package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var (
	isCacheOn bool
	cacher    *Cacher
)

// Cacher is handler for cache
type Cacher struct {
	XlsxInfos  map[string]*XlsxInfo `json:"xlsxinfos"`
	ProtoInfos map[string][]byte    `json:"protoinfos"`
}

// XlsxInfo contains sheet information in xlsx files
type XlsxInfo struct {
	FileName string `json:"filename"`
	MD5      []byte `json:"md5"`
}

// CacheInit initialize cacher and read from file
func CacheInit() {
	cacher = newCacher()

	if _, err := os.Stat(cfg.CacheFile); err == nil {
		err := cacher.Load()
		if err != nil {
			log.Fatal("load cache fail", err)
		}
	}
}

func newCacher() *Cacher {
	cacher := new(Cacher)
	cacher.XlsxInfos = make(map[string]*XlsxInfo)
	cacher.ProtoInfos = make(map[string][]byte)

	return cacher
}

// Load read data from json cache file
func (c *Cacher) Load() error {
	rawData, err := ioutil.ReadFile(cfg.CacheFile)
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
	rawData, err := json.MarshalIndent(cacher, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cfg.CacheFile, rawData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// ClearCache remove saved records and initialize a new cacher
func ClearCache() {
	if _, err := os.Stat(cfg.CacheFile); err == nil {
		if err := os.Remove(cfg.CacheFile); err != nil {
			log.Fatal("remove cache failed, ", err)
			return
		}
	}

	cacher = newCacher()
}
