package pgpkg

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
	pgwasi "github.com/wasilibs/go-pgquery"
)

// Parse the given SQL statement into a parse tree (Go struct format)
func Parse(input string) (tree *pgquery.ParseResult, err error) {
	return pgwasi.Parse(input)
}

func Deparse(tree *pgquery.ParseResult) (output string, err error) {
	return pgwasi.Deparse(tree)
}
