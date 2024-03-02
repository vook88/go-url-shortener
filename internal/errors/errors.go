package errors

import (
	errors1 "errors"
)

type DuplicateURLError struct {
	s string
}

func NewDuplicateURLError(text string) error {
	return &DuplicateURLError{text}
}

func (e *DuplicateURLError) Error() string {
	return e.s
}

var ErrURLDeleted = errors1.New("URL has been deleted")
