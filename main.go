package xlsx2pb

import (
	"flag"
)

func main() {
	if len(flag.Args()) != 2 {
		panic("Source path and target path is needed!")
	}

	isCacheOn = *flag.Bool("cache", true, "Use cache to check if .proto file needs to update")

	if isCacheOn {
		CacheInit()
	}

	if isCacheOn {

	}
}
