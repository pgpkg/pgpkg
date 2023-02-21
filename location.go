package pgpkg

import (
	"fmt"
)

// utility for generating errors that express their location.

type LocatableObject interface {
	Location() string
}

// ErrorAt wraps an error with its location. The location is not
// added if a Location is already present.
func ErrorAt(l LocatableObject, err error) error {
	return fmt.Errorf("%s: %w", l.Location(), err)
}

func ErrorFat(l LocatableObject, err error, format string, args... any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %s: %s", l.Location(), msg, err)
}