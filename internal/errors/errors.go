package errors

type DuplicateURLError struct {
	s string
}

func NewDuplicateURLError(text string) error {
	return &DuplicateURLError{text}
}

func (e *DuplicateURLError) Error() string {
	return e.s
}
