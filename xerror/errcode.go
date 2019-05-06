package xerror

import "fmt"

type CodeError interface {
	Code() int
}

func New(code int, format string, args... interface{}) error {
	return &codeError{fmt.Errorf(format, args...), code}
}

type codeError struct {
	error
	code int
}

func (e *codeError) Code() int {
	return e.code
}

func Code(e error) int {
	ec, ok := e.(CodeError)
	if ok {
		return ec.Code()
	}

	return -1
}
