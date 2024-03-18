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

	dsn := os.Getenv("PGPKG_DSN")

	switch os.Args[1] {
	case "deploy":
		doDeploy(dsn)

	case "try":
		doTry(dsn)

	case "repl":
		doRepl(dsn)

	case "export":
		doExport()

	case "import":
		doImport()

	default:
		usage()
		os.Exit(1)
	}
}
