package xlsx2pb

import (
	"flag"
)

func main() {
	isCacheOn = *flag.Bool("cache", true, "Use cache to check if .proto file needs to update")

	if isCacheOn {
		CacheInit()
	}

	for filename, sheets := range sheetFileMap {
		if IsXlsxChanged(filename) {
			for _, sheet := range sheets {
				err := ReadSheet(filename, sheet)
				if err != nil {
					// panic or continue as needed.
					panic(err)
				}
			}
		}
	}

	if isCacheOn {
		cacher.Save()
	}
}
