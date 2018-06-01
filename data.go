package xlsx2pb

import (
	"fmt"
	"math"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/tealeg/xlsx"
	"strings"
	"bufio"
	"path/filepath"
)

var (
	dataOutPrefix = "./data/"
	dataOutSuffix = ".data"
	xlsxPrefix = "./xlsx/"
	xlsxSuffix = ".xlsx"
)

// ProtoSheet is used to generate proto file
type ProtoSheet struct {
	ProtoIn
	ProtoOut
	buf *proto.Buffer
}

// ProtoIn contains data read from xlsx
type ProtoIn struct {
	Name    string
	vars    []*Val
	repeats []*Repeat
	hash    []byte
}

// ProtoOut controls how to output proto file
type ProtoOut struct {
	varIdx   int
	tabCount int
	isProto3 bool
	outProto []string
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
	colIdx     int
	proto2Type string
	typ        string
	name       string
	comment    string
	fieldNum   int
}

// Repeat contains repeat filed for .proto file
type Repeat struct {
	colIdx      int
	maxLength   int
	fieldName   string
	fieldLength int
	fieldNum    int
	repeatIdx   int
	comment     string
	fields      []*Val
}

func newProtoRow() *ProtoSheet {
	pr := new(ProtoSheet)
	pr.varIdx = 1
	pr.buf = proto.NewBuffer([]byte{})

	return pr
}

func ReadSheet(fileName, sheetName string) error {
	xlsxFullName := filepath.Join(xlsxPrefix, fileName + xlsxSuffix)
	if _, err := os.Stat(xlsxFullName); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exists", fileName)
	}

	xlsxFile, err := xlsx.OpenFile(xlsxFullName)
	if err != nil {
		return err
	}

	xlsxSheet, ok := xlsxFile.Sheet[sheetName]
	if ok {
		readSheet(xlsxSheet)
	}

	return fmt.Errorf("xlsx file %s does not contain sheet %s", fileName, sheetName)
}

func readSheet(sheet *xlsx.Sheet) error {
	if sheet.MaxRow < RowData {
		return fmt.Errorf("Sheet contains no data")
	}

	// proto
	pr := readHeads(sheet)
	pr.GenProto()
	pr.Hash()

	// check if proto need update
	if string(pr.hash) != string(cacher.ProtoInfos[pr.Name]) {
		cacher.ProtoInfos[pr.Name] = pr.hash
		pr.WriteProto()
	}

	// always write data if xlsx file changed
	pr.readData(sheet)
	pr.WriteData()

	return nil
}

func readHeads(sheet *xlsx.Sheet) *ProtoSheet {
	pr := newProtoRow()
	pr.Name = sheet.Name
	var repeatLength, repeatStructLength int // These are counters for repeat structure
	var curRepeat *Repeat

	for colIdx := 0; colIdx < sheet.MaxCol; colIdx++ {
		headType := sheet.Cell(0, colIdx).Value

		if headType == "" {
			continue
		}

		switch headType {
		case Req, Opt:
			val := new(Val)
			val.colIdx = colIdx
			val.proto2Type = headType
			val.typ = sheet.Cell(RowType, colIdx).Value
			val.name = sheet.Cell(RowID, colIdx).Value
			val.comment = sheet.Cell(RowComment, colIdx).Value

			if repeatLength == 0 {
				pr.vars = append(pr.vars, val)
			} else {
				// read repeat struct but avoid duplicate
				if len(curRepeat.fields) != curRepeat.fieldLength {
					curRepeat.fields = append(curRepeat.fields, val)
				}
				// check if repeat ends
				repeatStructLength--
				if repeatStructLength == 0 {
					repeatLength--
				}
			}
		case Rep:
			rp := new(Repeat)
			rp.colIdx = colIdx
			repeatLength, _ = sheet.Cell(1, colIdx).Int()
			rp.maxLength = repeatLength
			pr.repeats = append(pr.repeats, rp)
			curRepeat = rp
		case OptStru:
			// struct length counter
			if repeatStructLength != 0 {
				repeatStructLength, _ = sheet.Cell(1, colIdx).Int()
			}

			// init struct
			if curRepeat.maxLength == repeatLength {
				curRepeat.fieldLength, _ = sheet.Cell(1, colIdx).Int()
				curRepeat.fieldName = sheet.Cell(2, colIdx).Value
			}
		}
	}

	return pr
}

func (pr *ProtoSheet) readData(sheet *xlsx.Sheet) {
	// Add Tag
	pr.buf.EncodeVarint(uint64(10)) // (1 << 3) | 2 = 10

	data := make([]byte, 0)
	for i := RowData; i < sheet.MaxRow; i++ {
		if row := sheet.Rows[i]; len(row.Cells) != 0 && row.Cells[0].Value != "" {
			data = append(data, pr.readRow(row)...)
		}
	}

	pr.buf.EncodeRawBytes(data)
}

func (rp *Repeat) getCount(row *xlsx.Row) int {
	valCount, err := row.Cells[rp.colIdx].Int()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 0
	}

	return valCount
}

// readRow Marshal a row of data into binary data
func (pr *ProtoSheet) readRow(row *xlsx.Row) []byte {
	rowBuff := proto.NewBuffer([]byte{})
	for _, val := range pr.vars {
		readCell(rowBuff, val, row.Cells[val.colIdx]) // Variable part of data
	}

	for _, repeat := range pr.repeats {
		if rowCount := repeat.getCount(row); rowCount > 0 {
			rowBuff.EncodeVarint(uint64((repeat.fieldNum << 3) | 2)) // Add Repeat Tag
			rpData := make([]byte, 0)
			fieldBuff := proto.NewBuffer([]byte{})
			for count := 0; count < rowCount; count++ {
				fieldBuff.Reset()
				for _, val := range repeat.fields {
					readCell(fieldBuff, val, row.Cells[val.colIdx+count*(repeat.fieldLength+1)]) // next variable position = current position + field length + 1
				}
				rpData = append(rpData, fieldBuff.Bytes()...)
			}
			rowBuff.EncodeRawBytes(rpData) // Add Repeat data
		}
	}

	return rowBuff.Bytes()
}

// readCell add "Tag - Value" or "Tag - Length - Value" to buffer according to var type
func readCell(b *proto.Buffer, val *Val, cell *xlsx.Cell) {
	tag := func(wireType int) {
		b.EncodeVarint(uint64((val.fieldNum << 3) | wireType))
	}

	switch val.typ {
	case "int32", "int64", "uint32", "uint64":
		tag(0)
		intVal, err := cell.Int()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		b.EncodeVarint(uint64(intVal))
	case "sint32", "sint64":
		tag(0)
		intVal, err := cell.Int()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		b.EncodeZigzag64(uint64(intVal))
	case "float", "float64":
		tag(1)
		floatVal, err := cell.Float()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		b.EncodeFixed64(math.Float64bits(floatVal))
	case "string":
		tag(2)
		b.EncodeStringBytes(cell.Value) // length included
	default:
		fmt.Fprintf(os.Stderr, "invalid var type: %v\n", val.typ)
		return
	}
}

func (pr *ProtoSheet) WriteData() {
	f, err := os.Create(dataOutPrefix + strings.ToLower(pr.Name) + dataOutSuffix)
	defer f.Close()

	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)
	w.Write(pr.buf.Bytes())
	w.Flush()
}