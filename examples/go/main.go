package main

import (
	"embed"
	"fmt"
	"github.com/pgpkg/pgpkg"
)

//go:embed pgpkg.toml *.sql schema
var examplePkgFS embed.FS

// TODO
// - [ ] document each line
// - [ ] add a test case

func main() {
	var err error

	if err = pgpkg.ParseArgs("pgpkg"); err != nil {
		pgpkg.Exit(err)
	}

	p := pgpkg.NewProject()
	if _, err = p.AddEmbeddedFS(examplePkgFS, ""); err != nil {
		pgpkg.Exit(err)
	}

	db, err := p.Open("")
	if err != nil {
		pgpkg.Exit(err)
	}
	defer db.Close()

	contactId := "90069E0B-2998-45E0-B8FC-91610761B429"
	var contactName string
	if err = db.QueryRow("select example.get_contact($1)", contactId).Scan(&contactName); err != nil {
		pgpkg.Exit(err)
	}

	fmt.Printf("Found contact %s: %s\n", contactId, contactName)
}
