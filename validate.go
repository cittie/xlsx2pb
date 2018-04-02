package xlsx2pb

import (
	"crypto/md5"
	"io"
	"log"
	"os"
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
