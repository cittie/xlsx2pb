package xlsx2pb

import (
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/tealeg/xlsx"
)

var (
	testSheet *xlsx.Sheet
)

func init() {
	xf, err := xlsx.OpenFile("./test/Sample.xlsx")
	if err != nil {
		panic(err)
	}

	if testSheet = xf.Sheet["SAMPLEONE"]; testSheet == nil {
		panic("Unable to find test sheet!")
	}
}

func TestReadHeads(t *testing.T) {
	pr := readHeads(testSheet)

	assert.Equal(t, 3, len(pr.vars))
	assert.Equal(t, "uint64", pr.vars[0].typ)
	assert.Equal(t, "SampleID", pr.vars[0].name)
	assert.Equal(t, "Comment1", pr.vars[0].comment)

	assert.Equal(t, 1, len(pr.repeats))
	assert.Equal(t, 3, pr.repeats[0].maxLength)
	assert.Equal(t, 3, len(pr.repeats[0].fields))
	assert.Equal(t, 3, pr.repeats[0].fieldLength)
	assert.Equal(t, "string", pr.repeats[0].fields[0].typ)
	assert.Equal(t, "RewardID", pr.repeats[0].fields[0].name)
	assert.Equal(t, "Comment4", pr.repeats[0].fields[0].comment)
}

func TestWriteProto(t *testing.T) {
	pr := readHeads(testSheet)
	pr.GenProto()
	assert.Equal(t, len(pr.outProto), 24)

	srcFile := "./test/sampleone.proto"
	tarFile := "./proto/sampleone.proto"

	pr.Write()
	assert.Equal(t, getFileMD5(srcFile), getFileMD5(tarFile))
	defer os.Remove(tarFile)
}

func TestReadCell(t *testing.T) {
	tests := []struct {
		cellValue string
		readType  string
		expected  []byte
	}{
		{"3", "int32", []byte{0, 3}},
		{"150", "uint32", []byte{0, 150, 1}},
		{"270", "int64", []byte{0, 142, 2}},
		{"86942", "uint64", []byte{0, 158, 167, 5}},
		{"128", "sint32", []byte{0, 128, 2}},
		{"-2", "sint64", []byte{0, 3}},
		{"0.125", "float", []byte{1, 0, 0, 0, 0, 0, 0, 192, 63}},
		{"testing", "string", []byte{2, 7, 116, 101, 115, 116, 105, 110, 103}},
	}

	cell := new(xlsx.Cell)
	buff := proto.NewBuffer([]byte{})
	val := new(Val)

	for _, test := range tests {
		buff.Reset()
		val.typ = test.readType
		cell.SetString(test.cellValue)
		readCell(buff, val, cell)
		assert.Equal(t, test.expected, buff.Bytes())
	}
}

func TestReadData(t *testing.T) {
	pr := readHeads(testSheet)
	pr.readData(testSheet)

	assert.Equal(t, 247, len(pr.buf.Bytes()))
}
