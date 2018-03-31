package xlsx2pb

// XlsxFileInfo contains original xlsx file information
type XlsxFileInfo struct {
	Filename string
	MD5      [16]byte
}
