package xlsx2pb

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

// ProtoRow is used to generate proto file
type ProtoRow struct {
	varIdx  uint32
	vars    []*SheetHead
	repeats []*Repeat
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
	attr int
	typ  int
	name string
}

// Repeat contains repeat filed for .proto file
type Repeat struct {
	maxLength   int
	fieldName   string
	fieldLength int
	fields      []SheetHead
}

func readSheet(sheet *xlsx.Sheet) error {
	if sheet.MaxRow < RowData {
		return fmt.Errorf("Sheet contains no data")
	}

	return nil
}

func readHeads(sheet *xlsx.Sheet) {
	//sheetName := sheet.Name

	for colIdx := 0; colIdx < sheet.MaxCol; colIdx++ {

		// Attr
		switch sheet.Cell(0, colIdx).Value {
		case Req, Opt:
		case Rep:
		case OS:
		}
	}
}
