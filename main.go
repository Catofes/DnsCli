package main

import (
	"github.com/Catofes/DnsCli/src"
	"flag"
	"os"
	"fmt"
)

var _version_ string

func main() {
	versionFlag := flag.Bool("v", false, "Show version.")
	configPathFlag := flag.String("c","","Config path.")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("Git commit id: %s.\n", _version_)
		os.Exit(0)
	}
	dnscli.Do(*configPathFlag)
}
