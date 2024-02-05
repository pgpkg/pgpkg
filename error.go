package pgpkg

// utilities for generating errors that express their location and context.

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Used when a --option requires the caller to quit, but there wasn't an error.
// e.g., --dry-run
var ErrUserRequest = errors.New("terminating due to user request")

// PKGObject is any object (statement, unit, package) that can tell us
// where a problem happened.
type PKGObject interface {
	Location() string
	DefaultContext() *PKGErrorContext
}

// PKGErrorContext represents the execution context in which an error occurred.
// Error context may be inside of pgpkg structures, but can also be found within
// stored procedures at runtime (eg during tests). For this reason the context may be
// independent of the object which caused the error.

type PKGErrorContext struct {
	Source     string
	LineNumber int
	Location   string
	Next       *PKGErrorContext // Indicates addtional stack traces.
}

// PKGError is the error type used internally by pgpkg.
type PKGError struct {
	Message string
	Object  PKGObject
	Context *PKGErrorContext
	Err     error

	// In the case of applying a migration, there may be multiple errors,
	// e.g. if a MOB can't be processed (typically because of dependencies).
	Errors []*PKGError
}

func (e *PKGError) Unwrap() error {
	return e.Err
}

func (e *PKGError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %s", e.Object.Location(), e.Message, e.Err.Error())
	}

	return fmt.Sprintf("%s: %s", e.Object.Location(), e.Message)
}

// UnwrapAll unwraps the errors until we get to the very last PKGError.
func (e *PKGError) UnwrapAll() []*PKGError {
	var err error
	var pkgErrors []*PKGError

	for err = e; err != nil; err = errors.Unwrap(err) {
		pkgErr, ok := err.(*PKGError)
		if ok {
			pkgErrors = append(pkgErrors, pkgErr)
		}
	}

	return pkgErrors
}

func (e *PKGError) GetContext() *PKGErrorContext {
	if e.Context != nil {
		return e.Context
	}

	return e.Object.DefaultContext()
}

// Print prints useful information about this error.
func (e *PKGError) PrintRootContext(contextLines int) {
	pkgErrors := e.UnwrapAll()

	for _, pkgErr := range pkgErrors {
		c := pkgErr.Context
		_, _ = fmt.Fprintln(os.Stderr, pkgErr.Error())

		for c != nil {
			c.Print(contextLines)
			c = c.Next
		}
	}
}

func PKGErrorf(object PKGObject, err error, msg string, args ...any) *PKGError {
	return &PKGError{
		Message: fmt.Sprintf(msg, args...),
		Object:  object,
		Err:     err,
		// Context: object.DefaultContext(),   // don't set default context, use pkgerr.GetContext() instead.
	}
}

func (c *PKGErrorContext) Print(contextLines int) {
	sourceLine := c.LineNumber - 1

	if c == nil {
		return
	}

	lines := strings.Split(c.Source, "\n")
	lineCount := len(lines)

	for cl := sourceLine - contextLines; cl <= sourceLine+contextLines; cl++ {
		if cl >= 0 && cl < lineCount {
			if cl != sourceLine {
				Stderr.Printf("    %4d: %s\n", cl+1, lines[cl])
			} else {
				Stderr.Printf("--> %4d: %s\n", cl+1, lines[cl])
			}
		}
	}

	trace := c
	for trace != nil {
		fmt.Fprintln(os.Stderr, trace.Location)
		trace = trace.Next
	}
}

// Exit usually prints the error message (with context, if available), and then exits immediately with status 1.
// However, err == ErrUserRequest then no message is printed and we exit with status 0.
// project.Open() will return ErrUserRequest if the command-line options indicate a dry run or other
// condition that should not result in a program printing an error.
//
// Applications should call pgpkg.Exit(err) after calling project.Open() if err is not nil.
func Exit(err error) {
	if err == ErrUserRequest {
		os.Exit(0)
	}
	PrintError(err)
	os.Exit(1)
}

func PrintError(err error) {
	var pkgErr *PKGError
	ok := errors.As(err, &pkgErr)
	if !ok {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}

	pkgErr.PrintRootContext(2)

	for _, subErr := range pkgErr.Errors {
		subErr.PrintRootContext(2)
	}
}
