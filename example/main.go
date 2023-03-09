package main

import (
	"embed"
	"fmt"
	"github.com/pgpkg/pgpkg"
)

//go:embed pgpkg.toml *.sql schema
var hello embed.FS

func main() {
	pgpkg.ParseArgs()
	var p pgpkg.Project
	p.AddFS(hello)

	db, err := p.Open()
	if err != nil {
		pgpkg.Exit(err)
	}
	defer db.Close()

	var world string
	if err = db.QueryRow("select hello.world()").Scan(&world); err != nil {
		pgpkg.Exit(err)
	}

	fmt.Println("And the world is:", world)
}
