package lib

// Run execute the xlsx2pb and output proto files and binary data files
func Run(isCacheOn bool) {
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
