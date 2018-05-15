package xlsx2pb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func genTestProtoRow() *ProtoRow {
	pr := new(ProtoRow)

	pr.Name = "TestProtoRow"

	// Field
	sh := &Val{
		colIdx:  0,
		typ:     "string",
		name:    "TestField1",
		comment: "* This is TestFiled1 *",
	}

	// Repeat Field
	sh2 := &Val{
		colIdx:  1,
		typ:     "int64",
		name:    "TestRepeat1",
		comment: "** This is TestRepeat1 **",
	}

	pr.vars = append(pr.vars, sh)

	// Repeat Struct
	rp := &Repeat{
		maxLength:   3,
		fieldName:   "TestRepeatStruct",
		fieldLength: 1,
	}

	rp.fields = append(rp.fields, sh2)

	pr.repeats = append(pr.repeats, rp)

	return pr
}

func TestAddMessage(t *testing.T) {
	pr := genTestProtoRow()
	var idx int

	checkOutput := func(out string) {
		assert.Equal(t, out, pr.outProto[idx])
		idx++
	}

	pr.AddMessageHead(pr.Name)
	checkOutput("message TestProtoRow {")

	pr.AddVal(pr.vars[0])
	checkOutput("  /* * This is TestFiled1 * */")
	checkOutput("  string TestField1 = 1;")

	pr.AddRepeat(pr.repeats[0])
	checkOutput("  ")
	checkOutput("  message TestRepeatStruct {")
	checkOutput("    /* ** This is TestRepeat1 ** */")
	checkOutput("    int64 TestRepeat1 = 1;")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  repeated TestRepeatStruct testRepeatStruct = 2;")

	pr.AddMessageTail()
	checkOutput("}")
	checkOutput("")

	pr.AddMessageArray()
	checkOutput("message TestProtoRow_ARRAY {")
	checkOutput("  repeated TestProtoRow testProtoRow = 1;")
	checkOutput("}")
	checkOutput("")
}
