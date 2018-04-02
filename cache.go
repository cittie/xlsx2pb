package xlsx2pb

import "os"

var (
	isCacheOn     bool // if cache mechanism will be used
	cacheFileName = "./cache/sheetcache"
	cacher        *Cacher
)

// Cacher is handler for cache
type Cacher struct {
	xlsxInfos []*XlsxInfo
}

// XlsxInfo contains sheet information in xlsx files
type XlsxInfo struct {
	SheetName string
	MD5       [16]byte
}

func cacheInit() {
	cacher = new(Cacher)

	if _, err := os.Stat(cacheFileName); err == nil {
		readCache()
	}
}

func readCache() {

}

func writeCache() {

}
