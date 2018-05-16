package xlsx2pb

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testCacher       *Cacher
	preCacheFileName string
)

func setup() {
	testCacher = newCacher()

	info := new(XlsxInfo)
	info.FileName = "CACHETESTSHEET"
	copy(info.MD5[:], []byte("f7ffd6e04e02a743fe8bec550e64cb71"))

	testCacher.XlsxInfos[info.FileName] = info

	// mock filename
	preCacheFileName = sheetCache
	sheetCache = "./cache/sheetcachetest.json"
}

func tearDown() {
	os.Remove(sheetCache)
	sheetCache = preCacheFileName
}

func TestCacheReadAndWrite(t *testing.T) {
	setup()

	// Save
	cacher = testCacher
	assert.NoError(t, cacher.Save())

	// Load
	cacher = newCacher()
	assert.NoError(t, cacher.Load())
	assert.Equal(t, testCacher, cacher)

	// Clear
	ClearCache()
	assert.Equal(t, newCacher(), cacher)
	_, err := os.Stat(sheetCache)
	assert.Error(t, err)

	tearDown()
}
