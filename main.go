package main

import (
	"flag"

	"github.com/cittie/xlsx2pb/lib"
)

func main() {
	var useCache = flag.Bool("cache", true, "Use cache for current xlsx")

	lib.Run(*useCache)
}
