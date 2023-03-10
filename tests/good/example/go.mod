module github.com/pgpkg/example

go 1.20

require github.com/pgpkg/pgpkg v0.0.0-20230308035832-9dac5760b539

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gookit/color v1.5.2 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/pganalyze/pg_query_go/v4 v4.2.0 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/sys v0.0.0-20210330210617-4fbd30eecc44 // indirect
	google.golang.org/protobuf v1.23.0 // indirect
)

replace github.com/pgpkg/pgpkg => ../../../
