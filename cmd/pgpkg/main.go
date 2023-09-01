package main

import (
	"os"
)

// This command-line version of pgpkg takes one or more directories or ZIP files, and installs them into the database.
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "deploy":
		doDeploy()

	case "try":
		doTry()

	case "repl":
		doRepl()

	case "export":
		doExport()

	case "import":
		doImport()

	default:
		usage()
		os.Exit(1)
	}
}
