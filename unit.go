package pgpkg

//
// This file is the unit parser for pgpkg. It uses the official postgresql parser to
// parse a package into a set of build units (Units), consisting of SQL statements.
//

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"io"
	"strings"
)

// Unit (ie, build unit) represents potentially parsable tree of SQL source code
// taken from a single file. Units are lazily loaded, and don't parse their
// contents until requested, with the Statements() function.
// Once the unit is compiled, individual statements contain line number and other
// debugging information.
type Unit struct {

	// The Bundle that this unit belongs to.
	Bundle *Bundle

	// Path is the filename within the Bundle FS that this Unit
	// should read from when it's parsed.
	Path string

	// The contents (SQL statements) declared in the unit.
	Source string

	// The list of parsed statements in this unit.
	Statements []*Statement
}

// Add a statement to a unit. The parser will include all whitespace and comments prefixing
// the first line of code, so we remove that and increase the line number to give us the first
// line of the actual code.
func (u *Unit) addStatement(lineNumber int, sql string, tree *pg_query.RawStmt) {

	// Skip over empty lines or lines that start with "--".
	lines := strings.Split(sql, "\n")
	skipped := 0
	offset := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			break
		}
		skipped++
		offset += len(line) + 1
	}

	statement := &Statement{
		Unit:       u,
		LineNumber: lineNumber + skipped,
		Tree:       tree,
	}

	if offset > len(sql) {
		statement.Source = sql
	} else {
		statement.Source = sql[offset:]
	}

	u.Statements = append(u.Statements, statement)
}

// Parse a unit.
func (u *Unit) Parse() error {
	r, err := u.Bundle.Open(u.Path)
	if err != nil {
		return PKGErrorf(u, err, "unable to open")
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return PKGErrorf(u, err, "unable to read")
	}

	source := strings.TrimSpace(string(b))

	// Empty files are OK.
	if source == "" {
		return nil
	}

	// Files starting with "--pgpkg:ignore" are ignored
	if strings.HasPrefix(source, "--pgpkg:ignore") {
		return nil
	}

	// Automatically add a semicolon to the source if one
	// isn't there already.
	if source[len(source)-1] != ';' {
		source = source + ";"
	}
	u.Source = source

	parseResult, err := pg_query.Parse(source)
	if err != nil {
		// unfortunately parser errors return almost no information, so the best
		// we can do is identify the build unit. This seems to be a problem with
		// pg_query_go rather than the underlying PG parser itself.
		return PKGErrorf(u, err, "unable to parse")
	}

	lineNumber := 1

	// Add the statements to the unit.
	for _, stmt := range parseResult.Stmts {
		sql := source[stmt.StmtLocation : stmt.StmtLocation+stmt.StmtLen]
		u.addStatement(lineNumber, sql, stmt)

		// find all the \n's in the statement, which will give us the new line number.
		lineNumber = lineNumber + strings.Count(sql, "\n")
	}

	return nil
}

func (u *Unit) Location() string {
	if u != nil {
		return u.Bundle.Location() + "/" + u.Path
	} else {
		return "<internal>"
	}
}

func (u *Unit) DefaultContext() *PKGErrorContext {
	return &PKGErrorContext{
		Source:     u.Source,
		LineNumber: 1,
	}
}
