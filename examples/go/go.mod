module github.com/pgpkg/example

go 1.20

require github.com/pgpkg/pgpkg v0.0.0-20240102003238-544ff3b3fd02

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/pganalyze/pg_query_go/v4 v4.2.3 // indirect
	google.golang.org/protobuf v1.23.0 // indirect
)

replace github.com/pgpkg/pgpkg => ../../
