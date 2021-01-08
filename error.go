package test

import (
	"errors"
	"fmt"
)

var (
	ErrNotEqual        = errors.New("not equal")
	ErrUnsupportedType = errors.New("type not supported for evaluation")
	ErrDifferingTypes = errors.New("given values were not of same type")

	DifferentTypesFmt = `
a: (%v) %v
b: (%v) %v
`
)

type Error struct {
	err     error
	details string
}

var _ error = Error{}

func NewError(err error, format string, v ...interface{}) Error {
	return Error{
		err:     err,
		details: fmt.Sprintf(format, v...),
	}
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.err, err.details)
}

func (err Error) Unwrap() error {
	return err.err
}
