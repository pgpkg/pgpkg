package pgpkg

import (
	"log"
	"os"
)

var Stderr = log.New(os.Stderr, "pgpkg: ", log.LstdFlags)
var Stdout = log.New(os.Stdout, "pgpkg: ", log.LstdFlags)
var Verbose = Stdout
