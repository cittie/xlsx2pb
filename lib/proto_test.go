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
		CommonInfo: CommonInfo{
			colIdx:   0,
			fieldNum: 1,
			name:     "TestField1",
			comment:  "* This is TestFiled1 *",
		},
		proto2Type:      "optional",
		typ:             "string",
		defaultValueStr: `"default"`,
	}

	pr.vars = append(pr.vars, sh)

	// opt struct
	optS := &OptStruct{
		CommonInfo: CommonInfo{
			colIdx:   3,
			fieldNum: 3,
			name:     "TestOptStruct",
			comment:  "TestOptStructData",
		},
	}

	optS.fields = append(optS.fields, sh)
	pr.optStructs = append(pr.optStructs, optS)

	// Repeat Field
	sh2 := &Val{
		CommonInfo: CommonInfo{
			colIdx:   1,
			fieldNum: 1,
			name:     "TestRepeat1",
			comment:  "** This is TestRepeat1 **",
		},
		proto2Type:      "optional",
		typ:             "int64",
		defaultValueStr: "0",
	}

	// Repeat Struct
	rp := &Repeat{
		CommonInfo: CommonInfo{
			colIdx:   2,
			fieldNum: 2,
			name:     "TestRepeatStruct",
			comment:  "TestRepeatStructData",
		},
		repeatIdx: 1,
	}

	rp2 := new(Repeat)
	*rp2 = *rp

	rp.val = sh2

	rp2.opts = optS
	rp2.val = nil

	pr.repeats = append(pr.repeats, rp, rp2)

	return pr
}

/*func TestProtoSheet_GenProto(t *testing.T) {
	pr := genTestProtoRow()
	pr.isProto3 = true

	pr.GenProto()

	for _, line := range pr.outProto {
		fmt.Println(line)
	}
}*/

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

	pr.AddOptionalStruct(pr.optStructs[0])
	checkOutput("  ")
	checkOutput("  message TestOptStruct {")
	checkOutput("    /* * This is TestFiled1 * */")
	checkOutput("    optional string TestField1 = 1 [default = \"default\"];")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  optional TestOptStruct TestOptStructData = 3;")

	pr.AddRepeat(pr.repeats[0])
	checkOutput("  ")
	checkOutput("  /* ** This is TestRepeat1 ** */")
	checkOutput("  repeated int64 TestRepeat1 = 2;")

	pr.AddRepeat(pr.repeats[1])
	checkOutput("  ")
	checkOutput("  message TestOptStruct {")
	checkOutput("    /* * This is TestFiled1 * */")
	checkOutput("    optional string TestField1 = 1 [default = \"default\"];")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  repeated TestRepeatStruct TestRepeatStructData = 2;")

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

	pr.AddOptionalStruct(pr.optStructs[0])
	checkOutput("  ")
	checkOutput("  message TestOptStruct {")
	checkOutput("    /* * This is TestFiled1 * */")
	checkOutput("    string TestField1 = 1 [default = \"default\"];")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  TestOptStruct TestOptStructData = 3;")

	pr.AddRepeat(pr.repeats[0])
	checkOutput("  ")
	checkOutput("  /* ** This is TestRepeat1 ** */")
	checkOutput("  repeated int64 TestRepeat1 = 2;")

	pr.AddRepeat(pr.repeats[1])
	checkOutput("  ")
	checkOutput("  message TestOptStruct {")
	checkOutput("    /* * This is TestFiled1 * */")
	checkOutput("    string TestField1 = 1 [default = \"default\"];")
	checkOutput("  }")
	checkOutput("  ")
	checkOutput("  repeated TestRepeatStruct TestRepeatStructData = 2;")

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

	assert.Equal(t, "a00323db48cda8b383f8d3f3c284e37c", hex.EncodeToString(pr.protoHash[:]))
}
