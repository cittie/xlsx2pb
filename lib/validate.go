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

	fname := filepath.Join(cfg.XlsxPath, filename + cfg.XlsxExt)
	if fInfo, ok := cacher.XlsxInfos[filename]; ok {
		if string(getFileMD5(fname)) == string(fInfo.MD5) {
			return false
		}
		cacher.XlsxInfos[filename] = &XlsxInfo{
			FileName: filename,
			MD5:      getFileMD5(fname),
		}
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
