package pgpkg

import (
	"log"
	"os"
)

var Stderr *log.Logger = log.New(os.Stderr, "pgpkg: ", log.LstdFlags)
var Stdout *log.Logger = log.New(os.Stdout, "pgpkg: ", log.LstdFlags)
var Verbose = Stdout