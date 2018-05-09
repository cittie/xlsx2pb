package xlsx2pb

import (
	"testing"

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
