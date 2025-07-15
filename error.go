package harness

import "fmt"

type ErrorCode int

const (
	_ ErrorCode = iota
	ErrInvalidSetting
	ErrContextCancelled
	ErrTimeout
)

type Error struct {
	Code   ErrorCode
	Actual error
}

func (e *Error) Error() string {
	return fmt.Sprintf("(%d) %s", e.Code, e.Actual.Error())
}

func (e *Error) Unwrap() error {
	return e.Actual
}

func Errorf(code ErrorCode, format string, a ...any) error {
	return &Error{
		Code:   code,
		Actual: fmt.Errorf(format, a...),
	}
}
