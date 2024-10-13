package pgpkg

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

var dsn = os.Getenv("PGPKG_DSN")

func applyProject(dsn string, commit bool, pkgPath string) error {
	Options.DryRun = !commit

	p, err := NewProjectFrom(pkgPath)
	if err != nil {
		return fmt.Errorf("unable to open project %s: %w", pkgPath, err)
	}

	err = p.Migrate(dsn)
	if err != nil {
		return fmt.Errorf("unable to migrate project %s: %w", pkgPath, err)
	}

	return nil
}

// Install or migrate a pgpkg package. Some test packages are designed to test that pgpkg
// fails in some circumstances. Set `expectFailure` to return success when they do fail, and
// to return failure if they don't.
func testProject(t *testing.T, dsn string, commit bool, expectFailure bool, pkgPath string) {
	err := applyProject(dsn, commit, pkgPath)

	// applyProject returns ErrDryRun if commit is not set. So "no error" means
	// either err is nil, or it's a dry run error - and we expected it.
	if err == nil || (errors.Is(err, ErrDryRun) && !commit) {
		if expectFailure {
			t.Fatal("test should have produced an error, but did not")
		}
		return
	}

	if err != nil && expectFailure {
		return
	}

	t.Fatal(err)
}

func TestCrossSchema(t *testing.T) {
	testProject(t, dsn, false, false, "tests/good/cross-schema")
}

func TestDependencies(t *testing.T) {
	testProject(t, dsn, false, false, "tests/good/dependencies")
}

func TestComplexProject(t *testing.T) {
	testProject(t, dsn, false, false, "tests/good/gl")
}

func TestSchemaName(t *testing.T) {
	testProject(t, dsn, false, false, "tests/good/good-schema-name")
}

func TestObjects(t *testing.T) {
	testProject(t, dsn, false, false, "tests/good/objects")
}

func TestPassingTests(t *testing.T) {
	testProject(t, dsn, false, false, "tests/good/passing-tests")
}

func TestQuotedSchema(t *testing.T) {
	testProject(t, dsn, false, false, "tests/good/quoted-schema")
}

func TestBadSchemaName(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/bad-schema-name")
}

func TestDuplicateMigrationName(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/duplicate-migration-name")
}

func TestBadEntitySyntax(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/entity-syntax")
}

func TestBadFailingTests(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/bad-failing-tests")
}

func TestBadFuncDuplicates(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/func-duplicates")
}

func TestBadFuncSyntax(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/func-syntax")
}

func TestBadFuncArgs(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/function-args")
}

func TestBadSQLSyntax(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/sql-syntax")
}

func TestBadTableRef(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/table-ref")
}

func TestBadTextException(t *testing.T) {
	testProject(t, dsn, false, true, "tests/bad/test-exception")
}
