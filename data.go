package xlsx2pb

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

// ProtoRow is used to generate proto file
type ProtoRow struct {
	ProtoIn
	ProtoOut
}

// ProtoIn contains data read from xlsx
type ProtoIn struct {
	Name    string
	vars    []*SheetHead
	repeats []*Repeat
}

// ProtoOut controls how to output proto file
type ProtoOut struct {
	varIdx   int
	tabCount int
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

// SheetHead contains fields for .proto file
type SheetHead struct {
	attr    int
	typ     string
	name    string
	comment string
}

// Repeat contains repeat filed for .proto file
type Repeat struct {
	maxLength   int
	fieldName   string
	fieldLength int
	repeatIdx   int
	comment     string
	fields      []*SheetHead
}

func newProtoRow() *ProtoRow {
	pr := new(ProtoRow)
	pr.varIdx = 1

	return pr
}

func readSheet(sheet *xlsx.Sheet) error {
	if sheet.MaxRow < RowData {
		return fmt.Errorf("Sheet contains no data")
	}

	return nil
}

func readHeads(sheet *xlsx.Sheet) *ProtoRow {
	pr := newProtoRow()
	pr.Name = sheet.Name
	var repeatLength, repeatStructLength int // These are counters for repeat structure
	var curRepeat *Repeat

	for colIdx := 0; colIdx < sheet.MaxCol; colIdx++ {
		switch sheet.Cell(0, colIdx).Value {
		case Req, Opt:
			head := new(SheetHead)
			head.typ = sheet.Cell(1, colIdx).Value
			head.name = sheet.Cell(2, colIdx).Value
			head.comment = sheet.Cell(3, colIdx).Value

			if repeatLength == 0 {
				pr.vars = append(pr.vars, head)
			} else {
				// read repeat struct but avoid duplicate
				if len(curRepeat.fields) != curRepeat.fieldLength {
					curRepeat.fields = append(curRepeat.fields, head)
				}
				// check if repeat ends
				repeatStructLength--
				if repeatStructLength == 0 {
					repeatLength--
				}
			}
		case Rep:
			rp := new(Repeat)
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
