package lib

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genTestProtoRow() *ProtoSheet {
	pr := newProtoRow()

	pr.Name = "TestProtoRow"

	// Field
	sh := &Val{
		colIdx:     0,
		proto2Type: "optional",
		typ:        "string",
		name:       "TestField1",
		comment:    "* This is TestFiled1 *",
	}

	// Repeat Field
	sh2 := &Val{
		colIdx:     1,
		proto2Type: "optional",
		typ:        "int64",
		name:       "TestRepeat1",
		comment:    "** This is TestRepeat1 **",
	}

	pr.vars = append(pr.vars, sh)

	// Repeat Struct
	rp := &Repeat{
		maxLength:   3,
		fieldName:   "TestRepeatStruct",
		fieldLength: 1,
		repeatIdx: 1,
	}

	rp.fields = append(rp.fields, sh2)

	pr.repeats = append(pr.repeats, rp)

	return pr
}

func TestAddMessageProto2(t *testing.T) {
	pr := genTestProtoRow()
	var idx int

	checkOutput := func(out string) {
		assert.Equal(t, out, pr.outProto[idx])
		idx++
	}

	pr.AddPreHead()
	checkOutput("syntax = \"proto2\";")
	checkOutput("package xlsx2pb;")
	checkOutput("")

	pr.AddMessageHead(pr.Name)
	checkOutput("message TestProtoRow {")

	pr.AddVal(pr.vars[0])
	checkOutput("  /* * This is TestFiled1 * */")
	checkOutput("  optional string TestField1 = 1;")

	pr.AddRepeat(pr.repeats[0])
	checkOutput("  ")
	checkOutput("  message TestRepeatStruct {")
	checkOutput("    /* ** This is TestRepeat1 ** */")
	checkOutput("    optional int64 TestRepeat1 = 1;")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  repeated TestRepeatStruct testrepeatstructs = 2;")

	pr.AddMessageTail()
	checkOutput("}")
	checkOutput("")

	pr.AddMessageArray()
	checkOutput("message TestProtoRow_ARRAY {")
	checkOutput("  repeated TestProtoRow testprotorows = 1;")
	checkOutput("}")
	checkOutput("")
}

func TestAddMessageProto3(t *testing.T) {
	pr := genTestProtoRow()
	pr.isProto3 = true
	var idx int

	checkOutput := func(out string) {
		assert.Equal(t, out, pr.outProto[idx])
		idx++
	}

	pr.AddPreHead()
	checkOutput("syntax = \"proto3\";")
	checkOutput("package xlsx2pb;")
	checkOutput("")

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
	checkOutput("  repeated TestRepeatStruct testrepeatstructs = 2;")

	pr.AddMessageTail()
	checkOutput("}")
	checkOutput("")

	pr.AddMessageArray()
	checkOutput("message TestProtoRow_ARRAY {")
	checkOutput("  repeated TestProtoRow testprotorows = 1;")
	checkOutput("}")
	checkOutput("")
}

func TestHash(t *testing.T) {
	pr := genTestProtoRow()
	pr.GenProto()
	pr.Hash()

	assert.Equal(t, "5c4924a37f3600542bac12ffbe31dd70", hex.EncodeToString(pr.hash[:]))
}
