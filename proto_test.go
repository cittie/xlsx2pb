package xlsx2pb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func genTestProtoRow() *ProtoRow {
	pr := new(ProtoRow)

	pr.Name = "TestProtoRow"

	// Field
	sh := &SheetHead{
		attr:    0,
		typ:     "string",
		name:    "TestField1",
		comment: "* This is TestFiled1 *",
	}

	// Repeat Field
	sh2 := &SheetHead{
		attr:    1,
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

	pr.AddMessageHead(pr.Name)
	assert.Equal(t, "message TestProtoRow {", pr.outProto[0])

	pr.AddVal(pr.vars[0])
	assert.Equal(t, "  /* * This is TestFiled1 * */", pr.outProto[1])
	assert.Equal(t, "  string TestField1 = 1;", pr.outProto[2])

	pr.AddRepeat(pr.repeats[0])
	assert.Equal(t, "  ", pr.outProto[3])
	assert.Equal(t, "  message TestRepeatStruct {", pr.outProto[4])
	assert.Equal(t, "    /* ** This is TestRepeat1 ** */", pr.outProto[5])
	assert.Equal(t, "    int64 TestRepeat1 = 1;", pr.outProto[6])
	assert.Equal(t, "  }", pr.outProto[7])
	assert.Equal(t, "  ", pr.outProto[8])
	assert.Equal(t, "  repeated TestRepeatStruct testRepeatStruct = 2;", pr.outProto[9])

	pr.AddMessageTail()
	assert.Equal(t, "}", pr.outProto[10])
	assert.Equal(t, "", pr.outProto[11])

	pr.AddMessageArray()
	assert.Equal(t, "message TestProtoRow_ARRAY {", pr.outProto[12])
	assert.Equal(t, "  repeated TestProtoRow testProtoRow = 1;", pr.outProto[13])
	assert.Equal(t, "}", pr.outProto[14])
	assert.Equal(t, "", pr.outProto[15])
}
