package pgpkg

// This file runs the user-defined tests. Tests are a bit more complicated than you might expect because
// we don't want any results to be written to the database, but we need to maintain
// various states so we can report errors and get stack traces.
//
// Nothing is overly complex; but it's not as simple as just executing the units directly.

import (
	"fmt"
	"os"
	"strings"
)

type Tests struct {
	*Bundle
	state       *stmtApplyState
	NamedTests  map[string]*Statement
	BeforeTests map[string]*Statement
}

type TestFunctionType int

const (
	TestFunctionOther  TestFunctionType = iota // utility function, declared but not executed
	TestFunctionTest                           // test function, called during testing
	TestFunctionBefore                         // before function, called once, before tests start.
)

// Given a function name, is it a test function, a before function, or a utility function?
func getTestFunctionType(name string) TestFunctionType {
	if strings.HasSuffix(name, "_test") {
		return TestFunctionTest
	}

	if strings.HasSuffix(name, "_before") {
		return TestFunctionBefore
	}

	if strings.HasSuffix(name, "_after") {
		_, _ = fmt.Fprintf(os.Stderr, "warning: test name %s is reserved for future use\n", name)
		return TestFunctionOther
	}

	if strings.HasPrefix(name, "test_") {
		_, _ = fmt.Fprintf(os.Stderr, "warning: test name %s is deprecated; use %s_test instead\n", name, name[5:])
		return TestFunctionTest
	}

	return TestFunctionOther
}

func (t *Tests) parse() error {
	var pending []*Statement

	namedTests := make(map[string]*Statement)
	beforeTests := make(map[string]*Statement)
	definitions := make(map[string]*Statement)

	for _, u := range t.Units {
		if Options.Verbose {
			Verbose.Println("parsing tests", u.Location())
		}

		if err := u.Parse(); err != nil {
			return fmt.Errorf("unable to parse tests: %w", err)
		}

		for _, stmt := range u.Statements {
			obj, err := stmt.GetManagedObject()
			if err != nil {
				return err
			}

			if obj.ObjectType != "function" {
				return PKGErrorf(stmt, nil, "only functions can be defined in tests; %s %s", obj.ObjectType, obj.ObjectName)
			}

			// Rewrite the statement to set the schema and security options.
			err = rewrite(stmt)
			if err != nil {
				return err
			}

			// Get the unqualified name of the function.
			fname := strings.ToLower(strings.TrimPrefix(obj.ObjectName, "\""+obj.ObjectSchema+"\"."))
			argIndex := strings.IndexRune(fname, '(')

			// strip the quotes and the args
			fname = fname[1 : argIndex-1]

			// Check for duplicate test definitions. This can be a subtle bug because
			// all the statements are probably "create or replace", so we make it explicit.
			objName := obj.ObjectType + ":" + obj.ObjectName
			dupeStmt, dupe := definitions[objName]
			if dupe {
				return PKGErrorf(stmt, nil,
					"duplicate declaration for %s %s; also defined in %s",
					obj.ObjectType, obj.ObjectName, dupeStmt.Location())
			}

			// We save and execute definitions for all functions in the test scripts,
			// even if they are non-test objects.
			definitions[objName] = stmt

			testFunctionType := getTestFunctionType(fname)

			if (testFunctionType != TestFunctionOther) && len(obj.ObjectArgs) != 0 {
				return PKGErrorf(stmt, nil, "test functions cannot receive arguments: %s %s", obj.ObjectType, obj.ObjectName)
			}

			switch testFunctionType {
			case TestFunctionTest:
				t.Package.StatTestCount++
				namedTests[obj.ObjectName] = stmt
			case TestFunctionBefore:
				beforeTests[obj.ObjectName] = stmt
			default:
				// do nothing.
			}

			pending = append(pending, stmt)
		}
	}

	t.NamedTests = namedTests
	t.BeforeTests = beforeTests
	t.state = &stmtApplyState{pending: pending}
	return nil
}

// testStmt is the statement containing the test function.
// testStmt was executed when the tests were parsed, so it is only used to work out where
// problems might have happened.
func (t *Tests) runTest(tx *PkgTx, testName string, testStmt *Statement) error {
	if _, spErr := tx.Exec("savepoint unittest"); spErr != nil {
		return fmt.Errorf("unable to begin savepoint for test %s: %w", testName, spErr)
	}

	cmd := fmt.Sprintf("select %s", testName)
	_, testErr := tx.Exec(cmd)

	_, rberr := tx.Exec("rollback to savepoint unittest")
	if rberr != nil {
		panic(rberr)
	}

	if testErr == nil {
		if Options.ShowTests {
			Stdout.Println("  [pass]", testName)
		}
		return nil
	}

	if Options.ShowTests {
		Stdout.Println("* [FAIL]", testName)
	}

	pe := PKGErrorf(testStmt, testErr, "test failed: %s", testName)
	pe.Context = getErrorContext(tx, cmd, testErr)
	tx = nil
	return pe
}

func (t *Tests) runBefore(tx *PkgTx, beforeName string, beforeStmt *Statement) error {
	cmd := fmt.Sprintf("select %s", beforeName)
	_, testErr := tx.Exec(cmd)

	if testErr == nil {
		return nil
	}

	pe := PKGErrorf(beforeStmt, testErr, "before-test script failed: %s", beforeName)
	pe.Context = getErrorContext(tx, cmd, testErr)
	tx = nil
	return pe
}

func (t *Tests) Run(tx *PkgTx) error {

	// Rollback, and return either the error or an error from the rollback.
	if !Options.KeepTestScripts {
		defer func() {
			if tx != nil {
				_, rberr := tx.Exec("rollback to savepoint test")
				if rberr != nil {
					panic(rberr)
				}
			}
		}()
	}

	// Create a savepoint for the entire set of tests. This savepoint ensures that the
	// test scripts are removed after testing is complete.
	// Note that there is a separate savepoint for the individual tests.
	if !Options.KeepTestScripts {
		_, err := tx.Exec("savepoint test")
		if err != nil {
			return fmt.Errorf("unable to begin test savepoint: %w", err)
		}
	}

	// Parse all the functions.
	err := t.parse()
	if err != nil {
		return err
	}

	err = applyState(tx, t.state)
	if err != nil {
		return err
	}

	// Run the before-tests. These are run in this outer savepoint so the results are
	// available to all the tests within. Note that before-functions are global, not just limited
	// to the current file.
	for beforeName, beforeStmt := range t.BeforeTests {
		if err = t.runBefore(tx, beforeName, beforeStmt); err != nil {
			return err
		}
	}

	// Run the actual tests.
	for testName, testStmt := range t.NamedTests {
		if Options.IncludePattern != nil {
			if !Options.IncludePattern.MatchString(testName) {
				if Options.ShowSkipped {
					Stdout.Println("- [skip]", testName)
				}
				continue
			}
		}

		if Options.ExcludePattern != nil {
			if Options.ExcludePattern.MatchString(testName) {
				if Options.ShowSkipped {
					Stdout.Println("- [skip]", testName)
				}
				continue
			}
		}

		if err = t.runTest(tx, testName, testStmt); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tests) PrintInfo(w InfoWriter) {
	w.Println("Test Bundle")
	t.Bundle.PrintInfo(w)
}
