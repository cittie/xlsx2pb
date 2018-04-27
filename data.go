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
	Req = "required"
	Opt = "optional"
	Rep = "repeated"
	OS  = "optional_struct"
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

func readHeads(sheet *xlsx.Sheet) {
	pr := newProtoRow()
	pr.Name = sheet.Name

	for colIdx := 0; colIdx < sheet.MaxCol; colIdx++ {
		switch sheet.Cell(0, colIdx).Value {
		case Req, Opt:
			head := new(SheetHead)
			head.typ = sheet.Cell(1, colIdx).Value
			head.name = sheet.Cell(2, colIdx).Value
			head.comment = sheet.Cell(3, colIdx).Value

			pr.vars = append(pr.vars, head)
		case Rep:
		case OS:
		}
	}
}
