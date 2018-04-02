package xlsx2pb

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

// ProtoRow is used to generate proto file
type ProtoRow struct {
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

// Typ is type names for proto2
type Typ string

// Field type for proto2
const (
	Req Typ = "required"
	Opt Typ = "optional"
	Rep Typ = "repeated"
	OS  Typ = "optional_struct"
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
