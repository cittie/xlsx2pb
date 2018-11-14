package lib

import (
	"fmt"
)

// Run execute the xlsx2pb and output proto files and binary data files
func Run(isCacheOn bool) {
	if isCacheOn {
		fmt.Println("cache on, init cache...")
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
		fmt.Println("saving cache ...")
		if err := cacher.Save(); err != nil {
			fmt.Println(err)
		}
	}
}
