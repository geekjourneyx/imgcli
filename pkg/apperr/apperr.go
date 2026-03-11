package apperr

import "fmt"

type Error struct {
	Code     string `json:"code"`
	Message  string `json:"-"`
	ExitCode int    `json:"exit_code"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code, message string, exitCode int) *Error {
	return &Error{Code: code, Message: message, ExitCode: exitCode}
}

func Wrap(code string, exitCode int, err error, format string, args ...any) *Error {
	msg := fmt.Sprintf(format, args...)
	if err != nil {
		msg = msg + ": " + err.Error()
	}
	return &Error{Code: code, Message: msg, ExitCode: exitCode}
}

func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if ok := As(err, &appErr); ok {
		return appErr
	}
	return New("INTERNAL_ERROR", err.Error(), 1)
}
