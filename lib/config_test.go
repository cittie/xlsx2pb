package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigFiles(t *testing.T) {
	path := "../test"
	files := getConfigFiles(path)

	assert.Equal(t, 3, len(files))
	assert.Contains(t, files[0], "xlsx_sample.config")
	assert.Contains(t, files[1], "xlsx_sample_dummy.config")
	assert.Contains(t, files[2], "xlsx_sample_wrong.config")
}

func TestReadCfgFile(t *testing.T) {
	ResetConfigCache()

	file1 := "../test/xlsx_sample.config"

	readCfgFile(file1)
	assert.Equal(t, 4, len(sheetNames))

	// panic if config files contains incorrect lines
	file2 := "../test/xlsx_sample_wrong.config"
	assert.Panics(t, func() { readCfgFile(file2) })
}

func TestReadCfgLine(t *testing.T) {
	ResetConfigCache()

	tests := []struct {
		in      string
		isError bool
	}{
		{"SAMPLEONE Sample.xlsx", false},
		{"SAMPLETHREE,SAMPLEFOUR Sample.xlsx", false},
		{" SAMPLETWO Sample.xlsx ", false},
		{"SAMPLEONE  Sample.xlsx", true}, // Duplicate
		{"SAMPLEONE", true},
		{"SAMPLEONE SAMPLETWO Sample.xlsx", true},
	}

	for _, test := range tests {
		assert.Equal(t, test.isError, readCfgLine(test.in) != nil, "test: %v", test)
	}
}
