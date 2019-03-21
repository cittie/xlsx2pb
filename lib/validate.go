package lib

import (
	"crypto/md5"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getFileMD5(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}

	return hash.Sum(nil)[:]
}

// IsTitleValid check if the first letter of a xlsx title is uppercase.
func IsTitleValid(title string) bool {
	if len(title) == 0 {
		return false
	}

	return strings.Title(title) == title
}

// IsXlsxChanged check if file has the same hash value as before
func IsXlsxChanged(filename string) bool {
	if cacher == nil {
		return true
	}

	fname := filepath.Join(cfg.XlsxPath, filename+cfg.XlsxExt)
	if fInfo, ok := cacher.XlsxInfos[filename]; ok {
		fHash := getFileMD5(fname)
		if string(fHash) == string(fInfo.MD5) {
			fInfo.State = Remained
			return false
		}

		fInfo.MD5 = fHash
		fInfo.State = Updated
		return true
	}

	cacher.XlsxInfos[filename] = &DataInfo{
		Name:  filename,
		MD5:   getFileMD5(fname),
		State: New,
	}

	return true
}

// IsSheetExists check if file exist in xlsx folder
func IsSheetExists(xlsxName string) bool {
	xlsxFullName := filepath.Join(cfg.XlsxPath, xlsxName+cfg.XlsxExt)
	if _, err := os.Stat(xlsxFullName); err == nil {
		return true
	}

	return false
}

// IsProtoChanged check and update previous proto info
func IsProtoChanged(ps *ProtoSheet) bool {
	if cacher == nil {
		return true
	}

	if info, ok := cacher.ProtoInfos[ps.Name]; ok {
		if string(info.MD5) == string(ps.protoHash) {
			info.State = Remained
			return false
		}

		info.MD5 = ps.protoHash
		info.State = Updated
		return true
	}

	cacher.ProtoInfos[ps.Name] = &DataInfo{
		Name:  ps.Name,
		MD5:   ps.protoHash,
		State: New,
	}

	return true
}

func IsDataChanged(ps *ProtoSheet) bool {
	if cacher == nil {
		return true
	}

	if info, ok := cacher.DataInfos[ps.Name]; ok {
		if string(info.MD5) == string(ps.dataHash) {
			info.State = Remained
			return false
		}

		info.MD5 = ps.dataHash
		info.State = Updated
		return true
	}

	cacher.DataInfos[ps.Name] = &DataInfo{
		Name:  ps.Name,
		MD5:   ps.dataHash,
		State: New,
	}

	return true
}
