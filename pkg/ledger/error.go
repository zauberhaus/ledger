package ledger

import "fmt"

type Error struct {
	msg  string
	code int
}

func NewError(code int, format string, args ...interface{}) Error {
	return Error{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
}

func (e Error) Error() string {
	return string(e.msg)
}

func (e Error) Is(code int) bool {
	return e.code == code
}
