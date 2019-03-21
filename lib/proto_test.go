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
		colIdx:          0,
		proto2Type:      "optional",
		typ:             "string",
		name:            "TestField1",
		comment:         "* This is TestFiled1 *",
		defaultValueStr: `"default"`,
	}

	// Repeat Field
	sh2 := &Val{
		colIdx:          1,
		proto2Type:      "optional",
		typ:             "int64",
		name:            "TestRepeat1",
		comment:         "** This is TestRepeat1 **",
		defaultValueStr: "0",
	}

	pr.vars = append(pr.vars, sh)

	// Repeat Struct
	rp := &Repeat{
		maxLength:       3,
		fieldName:       "TestRepeatStruct",
		fieldLength:     1,
		repeatIdx:       1,
		defaultValueStr: `""`,
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
	checkOutput("package ProtobufGen;")
	checkOutput("")

	pr.AddMessageHead(pr.Name)
	checkOutput("message TestProtoRow {")

	pr.AddVal(pr.vars[0])
	checkOutput("  /* * This is TestFiled1 * */")
	checkOutput("  optional string TestField1 = 1 [default = \"default\"];")

	pr.AddRepeat(pr.repeats[0])
	checkOutput("  ")
	checkOutput("  message TestRepeatStruct {")
	checkOutput("    /* ** This is TestRepeat1 ** */")
	checkOutput("    optional int64 TestRepeat1 = 1 [default = 0];")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  /* TestRepeatStruct */")
	checkOutput("  repeated TestRepeatStruct testrepeatstructs = 2;")

	pr.AddMessageTail()
	checkOutput("}")
	checkOutput("")

	pr.AddMessageArray()
	checkOutput("message TestProtoRow_ARRAY {")
	checkOutput("  repeated TestProtoRow items = 1;")
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
	checkOutput("package ProtobufGen;")
	checkOutput("")

	pr.AddMessageHead(pr.Name)
	checkOutput("message TestProtoRow {")

	pr.AddVal(pr.vars[0])
	checkOutput("  /* * This is TestFiled1 * */")
	checkOutput("  string TestField1 = 1 [default = \"default\"];")

	pr.AddRepeat(pr.repeats[0])
	checkOutput("  ")
	checkOutput("  message TestRepeatStruct {")
	checkOutput("    /* ** This is TestRepeat1 ** */")
	checkOutput("    int64 TestRepeat1 = 1 [default = 0];")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  /* TestRepeatStruct */")
	checkOutput("  repeated TestRepeatStruct testrepeatstructs = 2;")

	pr.AddMessageTail()
	checkOutput("}")
	checkOutput("")

	pr.AddMessageArray()
	checkOutput("message TestProtoRow_ARRAY {")
	checkOutput("  repeated TestProtoRow items = 1;")
	checkOutput("}")
	checkOutput("")
}

func TestHash(t *testing.T) {
	pr := genTestProtoRow()
	pr.GenProto()
	pr.ProtoHash()

	assert.Equal(t, "9dcd1a127603012b57f5f41681660392", hex.EncodeToString(pr.protoHash[:]))
}
