package main

import (
	"flag"

	"github.com/cittie/xlsx2pb/lib"
)

func main() {
	var useCache = flag.Bool("cache", true, "Use cache for current xlsx")
	var useGoroutine = flag.Bool("goroutine", true, "Use goroutine for faster handling")

	lib.Run(*useCache, *useGoroutine)
}
