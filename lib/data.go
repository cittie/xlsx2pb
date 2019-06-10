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
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/tealeg/xlsx"
)

// ProtoSheet is used to generate proto file
// buf used to store data binary
type ProtoSheet struct {
	ProtoIn
	ProtoOut
	buf *proto.Buffer

	mutex sync.RWMutex
}

// ProtoIn contains data read from xlsx
type ProtoIn struct {
	Name       string
	vars       []*Val
	repeats    []*Repeat
	optStructs []*OptStruct
	dataHash   []byte
	fieldMap   map[string]int                 // map[fieldName]index in vars/opt structs/repeats
	dupMap     map[string]struct{}            // map[fieldName] to check if field name has been used in current sheet
	uniqueMap  map[string]map[string]struct{} // map[fieldName][uniqueName] to check if unique type of variant has duplicates
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
	Unique  = "unique"
)

// Val contains fields for .proto file
type Val struct {
	CommonInfo
	proto2Type      string
	typ             string
	defaultValueStr string
}

// OptStruct is a struct contains one or more variants
type OptStruct struct {
	CommonInfo
	messageIdx int // proto message index
	fields     []*Val
}

// Repeat contains repeat fields for .proto file
type Repeat struct {
	CommonInfo
	repeatIdx int // proto repeat index
	val       *Val
	opts      *OptStruct
}

// CommonInfo is common info for val, opt struct and repeat
type CommonInfo struct {
	name      string
	comment   string
	colIdx    int // position of column
	fieldNum  int // proto index
	curLength int // use for counter
	maxLength int // use for proto define
}

func newProtoRow() *ProtoSheet {
	pr := new(ProtoSheet)
	pr.fieldMap = make(map[string]int)
	pr.isProto3 = cfg.UseProto3
	pr.varIdx = 1
	pr.buf = proto.NewBuffer([]byte{})
	pr.uniqueMap = make(map[string]map[string]struct{})
	return pr
}

func newRepeat() *Repeat {
	rp := new(Repeat)
	rp.repeatIdx = 1

	return rp
}

func newOptStruct() *OptStruct {
	optS := new(OptStruct)
	optS.messageIdx = 1

	return optS
}

// ReadSheet Read data pair from *.config
func ReadSheet(fileName, sheetName string) error {
	sheets := make([]*xlsx.Sheet, 0)
	var preName string

	files := strings.Split(fileName, "|")
	if len(files) > 1 {
		preName = sheetName
	}

	for _, fn := range files {
		fmt.Printf("reading %v for sheets %v\n", fn, sheetName)

		xlsxFullName := filepath.Join(cfg.XlsxPath, fn+cfg.XlsxExt)
		if _, err := os.Stat(xlsxFullName); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exists", fn)
		}

		xlsxFile, err := xlsx.OpenFile(xlsxFullName)
		if err != nil {
			return err
		}

		// Verify all sheets exists in file
		sheetNames := strings.Split(sheetName, ",")
		if len(sheetNames) > 1 {
			sections := strings.Split(fileName, ".")
			preName = sections[0]
		}

		for _, sheetName := range sheetNames {
			xlsxSheet, ok := xlsxFile.Sheet[sheetName]
			if !ok {
				return fmt.Errorf("xlsx file %s does not contain sheet %s", fn, sheetName)
			}

			sheets = append(sheets, xlsxSheet)
		}
	}

	// Marshal data
	if err := readSheets(preName, sheets); err != nil {
		return err
	}

	fmt.Printf("done for %v for sheets %v\n", fileName, sheetName)

	return nil
}

func readSheets(preName string, sheets []*xlsx.Sheet) error {
	pr := newProtoRow()

	hasGenProto := false

	// use filename instead of sheet name if sheets are more than 1
	if len(sheets) > 1 {
		pr.Name = preName
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

	var curRepeat *Repeat
	var curOptS *OptStruct

	for colIdx := 0; colIdx < sheet.MaxCol; colIdx++ {
		headType := sheet.Cell(RowAttr, colIdx).Value

		if strings.TrimSpace(headType) == "" {
			continue
		}

		switch headType {
		case Req, Opt, Unique:
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

			switch {
			case curOptS != nil: // Check opts
				if len(curOptS.fields) < curOptS.maxLength {
					val.fieldNum = curOptS.messageIdx
					curOptS.messageIdx++
					curOptS.fields = append(curOptS.fields, val)
				}

				curOptS.curLength--
				if curOptS.curLength <= 0 {
					if curRepeat != nil {
						// TODO: check if all OptS are the same
						if curRepeat.opts == nil {
							curRepeat.opts = curOptS
						}
						curRepeat.curLength--

						if curRepeat.curLength <= 0 {
							pr.updateRepeat(curRepeat)
							curRepeat = nil
						}
					} else {
						fmt.Printf("%v col %v curOptS %+v\n", pr.Name, colIdx, curOptS)
						pr.updateOptStruct(curOptS)
					}

					curOptS = nil
				}
			case curRepeat != nil && curOptS == nil: // Check repeat variant
				if curRepeat.opts != nil {
					panic(fmt.Errorf("sheet struct invalid, max repeat value exceed"))
				}

				if curRepeat.val == nil {
					curRepeat.val = val
				}

				curRepeat.curLength--

				if curRepeat.curLength <= 0 {
					pr.updateRepeat(curRepeat)
					curRepeat = nil
				}
			default: // update val
				pr.updateVal(val)
			}
		case Rep:
			curRepeat = newRepeat()
			curRepeat.colIdx = colIdx
			curRepeat.maxLength, _ = sheet.Cell(RowType, colIdx).Int()
			curRepeat.curLength = curRepeat.maxLength
		case OptStru:
			curOptS = newOptStruct()
			curOptS.colIdx = colIdx
			curOptS.name = sheet.Cell(RowID, colIdx).Value
			curOptS.comment = sheet.Cell(RowComment, colIdx).Value
			curOptS.maxLength, _ = sheet.Cell(RowType, colIdx).Int()
			curOptS.curLength = curOptS.maxLength
		}
	}

	// sometimes people will not full fill all repeat structs they designed...
	if curRepeat != nil {
		pr.updateRepeat(curRepeat)
		curRepeat = nil
	}

	/*
		// Debug
		fmt.Printf("sheetName %v\n", sheet.Name)
		for _, val := range pr.vars {
			fmt.Printf("val %+v\n", val)
		}
		fmt.Printf("--------------------------")
	*/
}

func (pr *ProtoSheet) checkDupHead(info *CommonInfo) {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if _, ok := pr.dupMap[info.name]; ok {
		panic(fmt.Errorf("Verify failed! Duplicate name %v found in column %v of %v!!!\n", info.name, info.colIdx, pr.Name))
	}
	pr.dupMap[info.name] = struct{}{}
}

func (pr *ProtoSheet) checkDupUnique(varName, varValue string) error {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if _, ok := pr.uniqueMap[varName]; !ok {
		pr.uniqueMap[varName] = make(map[string]struct{}, 256)
	}

	if _, ok := pr.uniqueMap[varName][varValue]; ok {
		return fmt.Errorf("duplicate unique %s in %s", varValue, varName)
	}

	pr.uniqueMap[varName][varValue] = struct{}{}
	return nil
}

// updateVal if a variable is already in ProtoSheet, update its value, else add it
func (pr *ProtoSheet) updateVal(val *Val) {
	if val.name == "" {
		log.Printf("val %+v without name\n", val)
		return
	}

	pr.checkDupHead(&val.CommonInfo)

	pr.mutex.Lock()
	defer pr.mutex.Unlock()

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

func (pr *ProtoSheet) updateOptStruct(optS *OptStruct) {
	if optS == nil {
		log.Printf("optional struct %+v nil\n", optS)
		return
	}
	if optS.name == "" {
		log.Printf("optional struct %+v without name\n", optS)
		return
	}

	pr.checkDupHead(&optS.CommonInfo)

	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	idx, ok := pr.fieldMap[optS.name]
	if ok {
		// check if they are same opt struct
		if optS.maxLength == pr.optStructs[idx].maxLength {
			optS.fieldNum = pr.optStructs[idx].fieldNum
			pr.optStructs[idx] = optS
		} else {
			fmt.Printf("sheet %v opt struct length %v is not equal as others %v\n",
				pr.Name, optS.maxLength, pr.optStructs[idx].maxLength)
			return
		}
	} else {
		optS.fieldNum = pr.varIdx
		pr.varIdx++
		newIdx := len(pr.optStructs)
		pr.optStructs = append(pr.optStructs, optS)
		pr.fieldMap[optS.name] = newIdx
	}
}

// updateRepeat if a repeat has same optional struct name and maxLength as in ProtoSheet, update it, else add it
func (pr *ProtoSheet) updateRepeat(repeat *Repeat) {
	if repeat.opts != nil {
		repeat.name = repeat.opts.name
		repeat.comment = repeat.opts.comment
	} else if repeat.val != nil {
		repeat.name = repeat.val.name
	} else {
		log.Printf("[updateRepeat] empty repeat\n")
		return
	}

	if repeat.name == "" {
		log.Printf("[updateRepeat] empty repeat name from val or opt struct/n")
		return
	}

	pr.checkDupHead(&repeat.CommonInfo)

	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	idx, ok := pr.fieldMap[repeat.name]
	if ok {
		if repeat.maxLength == pr.repeats[idx].maxLength {
			// update
			repeat.fieldNum = pr.repeats[idx].fieldNum
			pr.repeats[idx] = repeat
		} else {
			fmt.Printf("sheet %v repeat length %v is not equal as others %v\n",
				pr.Name, repeat.maxLength, pr.repeats[idx].maxLength)
			return
		}
	} else {
		// add
		repeat.fieldNum = pr.varIdx // get proto index
		pr.varIdx++
		newIdx := len(pr.repeats)
		pr.repeats = append(pr.repeats, repeat)
		pr.fieldMap[repeat.name] = newIdx
	}

	if repeat.val != nil {
		repeat.val.fieldNum = repeat.fieldNum
	}
}

// resetAllIndex will set all variables including which are inside repeat structure to -1
func (pr *ProtoSheet) resetAllIndex() {
	pr.dupMap = make(map[string]struct{})

	for _, val := range pr.vars {
		val.colIdx = -1
	}
	for _, optS := range pr.optStructs {
		optS.colIdx = -1
		for _, val := range optS.fields {
			val.colIdx = -1
		}
	}
	for _, rp := range pr.repeats {
		rp.colIdx = -1
		if rp.val != nil {
			rp.val.colIdx = -1
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

	// if first cell of a line is empty, ignore this line
	if len(row.Cells) > 0 && strings.TrimSpace(row.Cells[0].Value) == "" {
		return nil
	}

	readval := func(idx int, val *Val, b *proto.Buffer) error {
		var e error
		if val.colIdx == -1 { // sheet has no field
			e = readCell(b, val, new(xlsx.Cell))
		} else if val.colIdx < len(row.Cells) {
			// check unique type data is really unique
			if val.proto2Type == Unique {
				if err := pr.checkDupUnique(val.name, row.Cells[val.colIdx].Value); err != nil {
					panic(err)
				}
			}
			e = readCell(b, val, row.Cells[val.colIdx]) // Variable part of data
		} else { // sheet cell is empty
			e = readCell(b, val, new(xlsx.Cell))
		}
		return e
	}

	for i, val := range pr.vars {
		err = readval(i, val, rowBuff)
		if err != nil {
			log.Printf("readCell to val %+v failed, %v", val, err)
		}
	}

	for _, optS := range pr.optStructs {
		// Tag
		err := rowBuff.EncodeVarint(uint64((optS.fieldNum << 3) | 2)) // Add Repeat Tag
		if err != nil {
			log.Fatal(err)
		}

		// vals
		fieldBuff := proto.NewBuffer([]byte{})
		for i, val := range optS.fields {
			err = readval(i, val, fieldBuff)
			if err != nil {
				log.Printf("readCell to val %v in opt struct %+v failed, %v", val, optS, err)
			}
		}

		err = rowBuff.EncodeRawBytes(fieldBuff.Bytes())
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, repeat := range pr.repeats {
		if repeat.colIdx == -1 {
			continue
		}
		// read the value of copy number
		if rowCount := repeat.getCount(row); rowCount > 0 {
			// repeat with struct, [RepeatTag][Tag1][Value1][Tag2]Value2]...
			if repeat.opts != nil {
				fieldBuff := proto.NewBuffer([]byte{})
				for count := 0; count < rowCount; count++ {
					fieldBuff.Reset()

					err := rowBuff.EncodeVarint(uint64((repeat.fieldNum << 3) | 2)) // Add Repeat Tag
					if err != nil {
						log.Fatal(err)
					}
					for _, val := range repeat.opts.fields {
						// if the rest of a row is blank
						cell := new(xlsx.Cell)
						if len(row.Cells) > val.colIdx+count*(repeat.opts.maxLength+1) {
							cell = row.Cells[val.colIdx+count*(repeat.opts.maxLength+1)]
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
			} else if repeat.val != nil { // repeat without struct [Tag][Value][Tag][Value]...
				for count := 0; count < rowCount; count++ {
					cell := new(xlsx.Cell)
					if len(row.Cells) > repeat.colIdx+count+1 {
						cell = row.Cells[repeat.colIdx+count+1]
					}

					err := readCell(rowBuff, repeat.val, cell) // next variable position = current position + field length + 1
					if err != nil {
						log.Printf("readCell to repeat %+v val %+v failed, %v", repeat, repeat.val, err)
					}
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
		err = b.EncodeFixed32(uint64(math.Float32bits(float32(floatVal))))
		if err != nil {
			return err
		}
	case "float64", "double":
		err := b.EncodeVarint(uint64((val.fieldNum << 3) | proto.WireFixed64))
		if err != nil {
			return err
		}
		floatVal, err := cell.Float()
		if err != nil {
			return err
		}
		err = b.EncodeFixed64(uint64(math.Float64bits(floatVal)))
		if err != nil {
			return err
		}
	case "string":
		err := b.EncodeVarint(uint64((val.fieldNum << 3) | proto.WireBytes)) // tag
		if err != nil {
			return err
		}
		err = b.EncodeStringBytes(strings.TrimSpace(cell.Value)) // length included, and also remove extra spaces for string type
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
