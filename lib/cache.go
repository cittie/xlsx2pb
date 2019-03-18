package lib

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	isCacheOn bool
	cacher    *Cacher
	changes   *Cacher
)

type CacheStatus int

const (
	None CacheStatus = iota
	Remained
	Updated
	New
)

// Cacher is handler for cache
type Cacher struct {
	XlsxInfos  map[string]*XlsxInfo  `json:"xlsx_info"`
	ProtoInfos map[string]*ProtoInfo `json:"proto_info"`
}

// XlsxInfo contains sheet information in xlsx files
type XlsxInfo struct {
	FileName string      `json:"filename"`
	MD5      []byte      `json:"md5"`
	State    CacheStatus `json:"-"` // 0: previous 1: updated 2: new
}

type ProtoInfo struct {
	Name  string      `json:"proto_name"`
	MD5   []byte      `json:"md5"`
	State CacheStatus `json:"-"` // 0: previous 1: updated 2: new
}

// CacheInit initialize cacher and read from file
func CacheInit() {
	cacher = newCacher()
	changes = newCacher()

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
	cacher.ProtoInfos = make(map[string]*ProtoInfo)

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
	// remove previous
	if err := os.RemoveAll(cfg.ChangeOutputPath); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(cfg.ChangeOutputPath+"/proto", 0777); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(cfg.ChangeOutputPath+"/data", 0777); err != nil {
		log.Fatal(err)
	}

	// clear files not appears in current time
	for fName, info := range c.XlsxInfos {
		switch info.State {
		case None:
			delete(c.XlsxInfos, fName)
		case Updated, New:
			changes.XlsxInfos[fName] = info
		}
	}

	for pName, info := range c.ProtoInfos {
		switch info.State {
		case None:
			delete(c.ProtoInfos, pName)
		case Updated, New:
			changes.ProtoInfos[pName] = info
			if err := CopyChangedProtoFiles(pName); err != nil {
				return err
			}
		}
	}

	rawData, err := json.MarshalIndent(cacher, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cfg.CacheFile, rawData, 0644)
	if err != nil {
		return err
	}

	changesRaw, err := json.MarshalIndent(cacher, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cfg.ChangeLog, changesRaw, 0644)
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

// CopyChangedProtoFiles if a file is changed, copy both proto and data file to output dir
func CopyChangedProtoFiles(fName string) error {
	fn := strings.ToLower(fName)
	srcDataFile := filepath.Join(cfg.DataOutPath, fn+cfg.DataOutExt)
	_, err := os.Stat(srcDataFile)
	if err != nil {
		return err
	}

	srcData, err := os.Open(srcDataFile)
	if err != nil {
		return err
	}
	defer srcData.Close()

	dstDataFile := filepath.Join(cfg.ChangeOutputPath, "data", fn+cfg.DataOutExt)
	dstData, err := os.Create(dstDataFile)
	if err != nil {
		return err
	}
	defer dstData.Close()

	_, err = io.Copy(dstData, srcData)

	srcProtoFile := filepath.Join(cfg.ProtoOutPath, fn+cfg.ProtoOutExt)
	_, err = os.Stat(srcProtoFile)
	if err != nil {
		return err
	}

	srcProto, err := os.Open(srcProtoFile)
	if err != nil {
		return err
	}
	defer srcProto.Close()

	dstProtoFile := filepath.Join(cfg.ChangeOutputPath, "proto", fn+cfg.ProtoOutExt)
	dstProto, err := os.Create(dstProtoFile)
	if err != nil {
		return err
	}
	defer dstProto.Close()

	_, err = io.Copy(dstProto, srcProto)

	return err
}
