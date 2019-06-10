package lib

import (
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/tealeg/xlsx"
)

var (
	testSheet *xlsx.Sheet
	xlsxPath  string
	protoPath string
	dataPath  string
)

func init() {
	xf, err := xlsx.OpenFile("../test/Sample.xlsx")
	if err != nil {
		panic(err)
	}

	if testSheet = xf.Sheet["SAMPLEONE"]; testSheet == nil {
		panic("Unable to find test sheet!")
	}
}

func MockUp() {
	xlsxPath = cfg.XlsxPath
	cfg.XlsxPath = "../test/"
	protoPath = cfg.ProtoOutPath
	cfg.ProtoOutPath = "../test/"
	dataPath = cfg.DataOutPath
	cfg.DataOutPath = "../test/"
}

func TearDown() {
	cfg.XlsxPath = xlsxPath
	cfg.ProtoOutPath = protoPath
	cfg.DataOutPath = dataPath
}

func TestReadHeads(t *testing.T) {
	pr := newProtoRow()
	pr.updateHeads(testSheet)

	assert.Equal(t, 3, len(pr.vars))
	assert.Equal(t, "uint64", pr.vars[0].typ)
	assert.Equal(t, "SampleID", pr.vars[0].name)
	assert.Equal(t, "Comment1", pr.vars[0].comment)

	assert.Equal(t, 1, len(pr.repeats))
	assert.Equal(t, 3, pr.repeats[0].maxLength)
	assert.Equal(t, 3, len(pr.repeats[0].opts.fields))
	assert.Equal(t, 3, pr.repeats[0].opts.maxLength)
	assert.Equal(t, "string", pr.repeats[0].opts.fields[0].typ)
	assert.Equal(t, "RewardID", pr.repeats[0].opts.fields[0].name)
	assert.Equal(t, "Comment4", pr.repeats[0].opts.fields[0].comment)
}

/*func TestWriteProto(t *testing.T) {
	tarFile := "../test/sampleone.proto"

	pr := newProtoRow()
	pr.updateHeads(testSheet)
	pr.GenProto()
	assert.Equal(t, len(pr.outProto), 27)
	pr.WriteProto()
	assert.Equal(t, getFileMD5("../test/sampleone2.proto"), getFileMD5(tarFile))

	pr = newProtoRow()
	pr.updateHeads(testSheet)
	pr.isProto3 = true
	pr.GenProto()
	pr.WriteProto()
	assert.Equal(t, getFileMD5("../test/sampleone3.proto"), getFileMD5(tarFile))

	//defer os.Remove(tarFile)
}*/

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
		{"-0.85", "float", []byte{5, 0x9a, 0x99, 0x59, 0xbf}},
		{"0.8", "double", []byte{1, 0x9a, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe9, 0x3f}},
		{"testing", "string", []byte{2, 7, 116, 101, 115, 116, 105, 110, 103}},
		{" testing \t", "string", []byte{2, 7, 116, 101, 115, 116, 105, 110, 103}},
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
	pr := newProtoRow()
	pr.updateHeads(testSheet)
	pr.readData(testSheet)

	assert.Equal(t, 260, len(pr.buf.Bytes()))
}

func TestReadSheet(t *testing.T) {
	MockUp()

	rmTmp := func() {
		for _, tmp := range []string{"../test/sampleone.data", "../test/sampleone.proto"} {
			if _, err := os.Stat(tmp); err == nil {
				os.Remove(tmp)
			}
		}
	}

	err := ReadSheet("Sample.xlsx", "SAMPLEONE")
	assert.Nil(t, err)

	rmTmp()

	TearDown()
}
