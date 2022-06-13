package ledger

import (
	"fmt"
	"net/http"
)

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

func (e Error) HttpStatusCode() int {
	if e.code < 100 {
		return http.StatusBadRequest
	} else {
		return e.code
	}
}

func (e Error) IsError(code int) bool {
	return e.code == code
}
