package xlsx2pb

import (
	"crypto/md5"
	"io"
	"log"
	"os"
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
