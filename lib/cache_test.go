package lib

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

	info := new(DataInfo)
	info.Name = "CACHETESTSHEET"
	copy(info.MD5[:], []byte("f7ffd6e04e02a743fe8bec550e64cb71"))

	testCacher.XlsxInfos[info.Name] = info

	// mock filename
	preCacheFileName = cfg.CacheFile
	cfg.CacheFile = "../cache/sheetcachetest.json"
}

func tearDown() {
	os.Remove(cfg.CacheFile)
	cfg.CacheFile = preCacheFileName
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
	_, err := os.Stat(cfg.CacheFile)
	assert.Error(t, err)

	tearDown()
}
