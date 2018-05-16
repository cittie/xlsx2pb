package xlsx2pb

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var (
	isCacheOn  bool
	sheetCache = "./cache/sheetcache.json"
	cacher     *Cacher
)

// Cacher is handler for cache
type Cacher struct {
	XlsxInfos  map[string]*XlsxInfo `json:"xlsxinfos"`
	ProtoInfos map[string]string    `json:"protoinfos"`
}

// XlsxInfo contains sheet information in xlsx files
type XlsxInfo struct {
	FileName string   `json:"filename"`
	MD5      [16]byte `json:"md5"`
}

// CacheInit initialize cacher and read from file
func CacheInit() {
	cacher = newCacher()

	if _, err := os.Stat(sheetCache); err == nil {
		cacher.Load()
	}
}

func newCacher() *Cacher {
	cacher := new(Cacher)
	cacher.XlsxInfos = make(map[string]*XlsxInfo)
	cacher.ProtoInfos = make(map[string]string)

	return cacher
}

// Load read data from json cache file
func (c *Cacher) Load() error {
	rawData, err := ioutil.ReadFile(sheetCache)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawData, cacher)
	if err != nil {
		return err
	}

	return nil
}

// Save write current data to cache file as json
func (c *Cacher) Save() error {
	rawData, err := json.Marshal(cacher)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(sheetCache, rawData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// ClearCache remove saved records and initialize a new cacher
func ClearCache() {
	if _, err := os.Stat(sheetCache); err == nil {
		os.Remove(sheetCache)
	}

	cacher = newCacher()
}
