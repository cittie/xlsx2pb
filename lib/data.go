package lib

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/tealeg/xlsx"
)

// ProtoSheet is used to generate proto file
// buf used to store data binary
type ProtoSheet struct {
	ProtoIn
	ProtoOut
	buf *proto.Buffer
}

// ProtoIn contains data read from xlsx
type ProtoIn struct {
	Name     string
	vars     []*Val
	repeats  []*Repeat
	dataHash []byte
	fieldMap map[string]int // map[fieldName]index in vars/repeats
}

// ProtoOut controls how to output proto file
type ProtoOut struct {
	varIdx    int
	tabCount  int
	isProto3  bool
	protoHash []byte
	outProto  []string
}

// Row index in sheet
const (
	RowAttr = iota
	RowType
	RowID
	RowComment
	RowData
)

// Field type for proto2
const (
	Req     = "required"
	Opt     = "optional"
	Rep     = "repeated"
	OptStru = "optional_struct"
)

// Val contains fields for .proto file
type Val struct {
	colIdx          int
	proto2Type      string
	typ             string
	name            string
	comment         string
	fieldNum        int // proto index
	defaultValueStr string
}

// Repeat contains repeat filed for .proto file
type Repeat struct {
	colIdx          int // start position of column
	maxLength       int
	fieldName       string
	fieldLength     int
	fieldNum        int // proto index
	repeatIdx       int // proto repeat index
	comment         string
	fields          []*Val
	defaultValueStr string
}

func newProtoRow() *ProtoSheet {
	pr := new(ProtoSheet)
	pr.fieldMap = make(map[string]int)
	pr.isProto3 = cfg.UseProto3
	pr.varIdx = 1
	pr.buf = proto.NewBuffer([]byte{})

	return pr
}

func newRepeat() *Repeat {
	rp := new(Repeat)
	rp.repeatIdx = 1

	return rp
}

// ReadSheet Read data pair from *.config
func ReadSheet(fileName, sheetName string) error {
	fmt.Printf("reading %v for sheets %v\n", fileName, sheetName)

	xlsxFullName := filepath.Join(cfg.XlsxPath, fileName+cfg.XlsxExt)
	if _, err := os.Stat(xlsxFullName); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exists", fileName)
	}

	xlsxFile, err := xlsx.OpenFile(xlsxFullName)
	if err != nil {
		return err
	}

	// Verify all sheets exists in file
	sheetNames := strings.Split(sheetName, ",")
	sheets := make([]*xlsx.Sheet, 0)
	for _, sheetName := range sheetNames {
		xlsxSheet, ok := xlsxFile.Sheet[sheetName]
		if !ok {
			return fmt.Errorf("xlsx file %s does not contain sheet %s", fileName, sheetName)
		}

		sheets = append(sheets, xlsxSheet)
	}

	// Marshal data
	if err := readSheets(fileName, sheets); err != nil {
		return err
	}

	fmt.Printf("done for %v for sheets %v\n", fileName, sheetName)

	return nil
}

func readSheets(filename string, sheets []*xlsx.Sheet) error {
	pr := newProtoRow()

	hasGenProto := false

	// use filename instead of sheet name if sheets are more than 1
	if len(sheets) > 1 {
		pr.Name = strings.TrimSpace(strings.TrimSuffix(filename, ".xlsx"))
	}

	for _, sheet := range sheets {
		if sheet.MaxRow < RowData {
			return fmt.Errorf("sheet %v contains no data", sheet)
		}

		// update head for each sheet, avoiding empty columns changes the col index
		pr.updateHeads(sheet)

		// check if proto need update
		// only hash head which will be used to generate proto file
		if !hasGenProto {
			pr.GenProto()
			pr.ProtoHash()
			hasGenProto = true
		}

		pr.readData(sheet)
	}

	if IsProtoChanged(pr) {
		pr.WriteProto()
	}

	if pr != nil && IsDataChanged(pr) {
		pr.WriteData()
	}

	return nil
}

func (pr *ProtoSheet) updateHeads(sheet *xlsx.Sheet) {
	if pr.Name == "" {
		pr.Name = strings.TrimSpace(sheet.Name)
	}

	pr.resetAllIndex() // clear previous sheet data

	var repeatLength, repeatStructLength int // These are counters for repeat structure
	var curRepeat *Repeat

	for colIdx := 0; colIdx < sheet.MaxCol; colIdx++ {
		headType := sheet.Cell(RowAttr, colIdx).Value

		if strings.TrimSpace(headType) == "" {
			continue
		}

		switch headType {
		case Req, Opt:
			val := new(Val)
			val.colIdx = colIdx
			val.proto2Type = headType
			val.typ = sheet.Cell(RowType, colIdx).Value
			val.name = sheet.Cell(RowID, colIdx).Value
			val.defaultValueStr = "0"
			if val.typ == "string" {
				val.defaultValueStr = `""`
			}
			// If val name has default value
			if strings.Contains(val.name, "=") {
				parts := strings.Split(val.name, "=")
				val.name, val.defaultValueStr = parts[0], parts[1]
			}
			val.comment = sheet.Cell(RowComment, colIdx).Value

			if repeatLength == 0 {
				pr.updateVal(val)
			} else {
				// read repeat struct but avoid duplicate
				if len(curRepeat.fields) != curRepeat.fieldLength {
					val.fieldNum = curRepeat.repeatIdx
					curRepeat.repeatIdx++
					curRepeat.fields = append(curRepeat.fields, val)
				}
				// check if repeat ends
				repeatStructLength--
				if repeatStructLength == 0 {
					repeatLength--
				}
			}
		case Rep:
			rp := newRepeat()
			rp.colIdx = colIdx
			repeatLength, _ = sheet.Cell(RowType, colIdx).Int()
			rp.maxLength = repeatLength
			curRepeat = rp
		case OptStru:
			// struct length counter
			repeatStructLength, _ = sheet.Cell(RowType, colIdx).Int()

			// avoid optional_struct before repeat
			if curRepeat == nil {
				fmt.Printf("sheet %v col %v found optional struct without repeat\n", pr.Name, colIdx)
				repeatStructLength = 0
				continue
			}

			// init struct
			if curRepeat.maxLength == repeatLength {
				curRepeat.fieldLength, _ = sheet.Cell(RowType, colIdx).Int()
				curRepeat.fieldName = sheet.Cell(RowID, colIdx).Value
				curRepeat.comment = sheet.Cell(RowComment, colIdx).Value
				pr.updateRepeat(curRepeat)
			}
		}
	}
}

// updateVal if a variable is already in ProtoSheet, update its value, else add it
func (pr *ProtoSheet) updateVal(val *Val) {
	if idx, ok := pr.fieldMap[val.name]; ok {
		// update
		val.fieldNum = pr.vars[idx].fieldNum // get proto index from previous val
		pr.vars[idx] = val
		if val.defaultValueStr != pr.vars[idx].defaultValueStr {
			fmt.Printf("sheet %v variable %v default value %v differs from others %v\n",
				pr.Name, val.name, val.defaultValueStr, pr.vars[idx].defaultValueStr)
		}
	} else {
		// add
		idx = len(pr.vars)
		pr.vars = append(pr.vars, val)
		pr.fieldMap[val.name] = idx
		val.fieldNum = pr.varIdx
		pr.varIdx++
	}
}

// updateRepeat if a repeat has same optional struct name and maxLength as in ProtoSheet, update it, else add it
func (pr *ProtoSheet) updateRepeat(repeat *Repeat) {
	idx, ok := pr.fieldMap[repeat.fieldName]
	if ok {
		if repeat.maxLength == pr.repeats[idx].maxLength {
			// update
			repeat.fieldNum = pr.repeats[idx].fieldNum
			pr.repeats[idx] = repeat
		} else {
			fmt.Printf("sheet %v repeat length %v is not equal as others %v\n", pr.Name, repeat.maxLength, pr.repeats[idx].maxLength)
			return
		}
	} else {
		// add
		repeat.fieldNum = pr.varIdx // get proto index
		pr.varIdx++
		idx = len(pr.repeats)
		pr.repeats = append(pr.repeats, repeat)
		pr.fieldMap[repeat.fieldName] = idx
	}
}

// resetAllIndex will set all variables including which are inside repeat structure to -1
func (pr *ProtoSheet) resetAllIndex() {
	for _, val := range pr.vars {
		val.colIdx = -1
	}
	for _, rp := range pr.repeats {
		for _, val := range rp.fields {
			val.colIdx = -1
		}
	}
}

func (pr *ProtoSheet) readData(sheet *xlsx.Sheet) {
	for i := RowData; i < sheet.MaxRow; i++ {
		if row := sheet.Rows[i]; len(row.Cells) != 0 && strings.TrimSpace(row.Cells[0].Value) != "" {
			rawRowData := pr.readRow(row)
			if len(rawRowData) != 0 {
				// Add Tag
				err := pr.buf.EncodeVarint(uint64(10)) // (1 << 3) | 2 = 10
				if err != nil {
					log.Fatal(err)
				}
				// Add contents
				err = pr.buf.EncodeRawBytes(rawRowData)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	pr.DataHash()
}

func (rp *Repeat) getCount(row *xlsx.Row) int {
	// no repeat content
	if rp.colIdx >= len(row.Cells) {
		return 0
	}
	// not use
	if strings.TrimSpace(row.Cells[rp.colIdx].Value) == "" {
		return 0
	}
	// not number
	valCount, err := row.Cells[rp.colIdx].Int()
	if err != nil {
		log.Printf("convert %v to number fail, %v", row.Cells[rp.colIdx].Value, err)
		return 0
	}

	return valCount
}

// readRow Marshal a row of data into binary data
func (pr *ProtoSheet) readRow(row *xlsx.Row) []byte {
	rowBuff := proto.NewBuffer([]byte{})
	var err error
	for i, val := range pr.vars {
		if val.colIdx == -1 { // sheet has no field
			err = readCell(rowBuff, val, new(xlsx.Cell))
		} else if i < len(row.Cells) && val.colIdx < len(row.Cells) {
			err = readCell(rowBuff, val, row.Cells[val.colIdx]) // Variable part of data
		} else { // sheet cell is empty
			err = readCell(rowBuff, val, new(xlsx.Cell))
		}

		if err != nil {
			log.Printf("readCell to val %+v failed, %v", val, err)
		}
	}

	for _, repeat := range pr.repeats {
		if rowCount := repeat.getCount(row); rowCount > 0 {
			fieldBuff := proto.NewBuffer([]byte{})
			for count := 0; count < rowCount; count++ {
				fieldBuff.Reset()
				err := rowBuff.EncodeVarint(uint64((repeat.fieldNum << 3) | 2)) // Add Repeat Tag
				if err != nil {
					log.Fatal(err)
				}
				for _, val := range repeat.fields {
					// if the rest of a row is blank
					cell := new(xlsx.Cell)
					if len(row.Cells) > val.colIdx+count*(repeat.fieldLength+1) {
						cell = row.Cells[val.colIdx+count*(repeat.fieldLength+1)]
					}
					err := readCell(fieldBuff, val, cell) // next variable position = current position + field length + 1
					if err != nil {
						log.Printf("readCell to repeat %+v val %+v failed, %v", repeat, val, err)
					}
				}
				err = rowBuff.EncodeRawBytes(fieldBuff.Bytes())
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	return rowBuff.Bytes()
}

// readCell add "Tag - Value" or "Tag - Length - Value" to buffer according to var type
func readCell(b *proto.Buffer, val *Val, cell *xlsx.Cell) error {
	if strings.TrimSpace(cell.Value) == "" {
		return nil
	}

	switch val.typ {
	case "int32", "int64", "uint32", "uint64":
		err := b.EncodeVarint(uint64((val.fieldNum << 3) | proto.WireVarint))
		if err != nil {
			return err
		}
		intVal, err := cell.Int()
		if err != nil {
			return err
		}
		err = b.EncodeVarint(uint64(intVal))
		if err != nil {
			return err
		}
	case "sint32", "sint64":
		err := b.EncodeVarint(uint64((val.fieldNum << 3) | proto.WireVarint))
		if err != nil {
			return err
		}
		intVal, err := cell.Int()
		if err != nil {
			return err
		}
		err = b.EncodeZigzag64(uint64(intVal))
		if err != nil {
			return err
		}
	case "float", "float32":
		err := b.EncodeVarint(uint64((val.fieldNum << 3) | proto.WireFixed32))
		if err != nil {
			return err
		}
		floatVal, err := cell.Float()
		if err != nil {
			return err
		}
		err = b.EncodeFixed32(math.Float64bits(floatVal))
		if err != nil {
			return err
		}
	case "float64":
		err := b.EncodeVarint(uint64((val.fieldNum << 3) | proto.WireFixed64))
		if err != nil {
			return err
		}
		floatVal, err := cell.Float()
		if err != nil {
			return err
		}
		err = b.EncodeFixed64(uint64(math.Float32bits(float32(floatVal))))
		if err != nil {
			return err
		}
	case "string":
		err := b.EncodeVarint(uint64((val.fieldNum << 3) | proto.WireBytes)) // tag
		if err != nil {
			return err
		}
		err = b.EncodeStringBytes(cell.Value) // length included
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid var type: %v\n", val.typ)
	}

	return nil
}

func (pr *ProtoSheet) WriteData() {
	f, err := os.Create(filepath.Join(cfg.DataOutPath, strings.ToLower(pr.Name)+cfg.DataOutExt))
	defer f.Close()

	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)
	_, err = w.Write(pr.buf.Bytes())
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		fmt.Println(err.Error())
	}
}

// Hash generate hash of proto content
func (pr *ProtoSheet) DataHash() {
	hash := md5.New()
	hash.Write(pr.buf.Bytes())
	pr.dataHash = hash.Sum(nil)
}
