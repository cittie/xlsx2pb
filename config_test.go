package xlsx2pb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigFiles(t *testing.T) {
	path := "./test"
	files := getConfigFiles(path)

	assert.Equal(t, 3, len(files))
	assert.Contains(t, files[0], "xlsx_sample.config")
	assert.Contains(t, files[1], "xlsx_sample_dummy.config")
	assert.Contains(t, files[2], "xlsx_sample_wrong.config")
}

func TestReadCfgFile(t *testing.T) {
	file1 := "./test/xlsx_sample.config"

	readCfgFile(file1)
	assert.Equal(t, 4, len(sheetNames))

	// panic if config files contains incorrect lines
	file2 := "./test/xlsx_sample_wrong.config"
	assert.Panics(t, func() { readCfgFile(file2) })
}
