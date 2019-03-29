package lib

import (
	"fmt"
	"sync"
)

// Run execute the xlsx2pb and output proto files and binary data files
func Run(isCacheOn, isUseGoroutine bool) {
	if isCacheOn {
		fmt.Println("cache on, init cache...")
		CacheInit()
	}

	if isUseGoroutine {
		runByGoroutine()
	} else {
		runOneByOne()
	}

	if isCacheOn {
		fmt.Println("saving cache ...")
		if err := cacher.Save(); err != nil {
			fmt.Println(err)
		}
	}
}

func runOneByOne() {
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
}

func runByGoroutine() {
	var wg sync.WaitGroup

	for filename, sheets := range sheetFileMap {
		if IsXlsxChanged(filename) {
			wg.Add(1)
			go func(filename string, sheets []string) {
				defer wg.Done()
				for _, sheet := range sheets {
					err := ReadSheet(filename, sheet)
					if err != nil {
						// panic or continue as needed.
						panic(err)
					}
				}
			}(filename, sheets)
		}
	}

	wg.Wait()
}
