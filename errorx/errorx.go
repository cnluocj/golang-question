package errorx

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

type Error interface {
	error
	fmt.Formatter
	Unwrap() error
	Cause() error
	Code() int
	Type() ErrType
	Stack() Stack
}

type Err struct {
	cause error
	code  int
	typ   ErrType
	stack Stack
}

func (e Err) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", e.Cause())
			for i := 0; i < len(e.stack); i++ {
				fmt.Fprintf(s, "\n%+v %+v %+v\n",
					e.stack[i].File, e.stack[i].Line, e.stack[i].Name)
			}
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

func (e Err) Error() string {
	return e.cause.Error()
}

func (e Err) Cause() error {
	return e.cause
}

func (e Err) Unwrap() error {
	return e
}

func (e Err) Code() int {
	return e.code
}

func (e Err) Type() ErrType {
	return e.typ
}

func (e Err) Stack() Stack {
	return e.stack
}

type ErrType string

const (
	ErrTypeNotFound ErrType = "not_found"
	ErrTypeTimeout  ErrType = "timeout"
	// TODO: add more error types
)

type Frame struct {
	Name string
	File string
	Line int
}

func newFrame(f uintptr) *Frame {
	fn := runtime.FuncForPC(uintptr(f) - 1)
	name, file, line := "unknown", "unknown", 0
	if fn != nil {
		name = fn.Name()
		file, line = fn.FileLine(uintptr(f) - 1)
	}
	return &Frame{
		Name: name,
		File: file,
		Line: line,
	}
}

type Stack []Frame

type stack []uintptr

func (s *stack) StackTrace() Stack {
	st := make([]Frame, len(*s))
	for i := 0; i < len(st); i++ {
		st[i] = *newFrame((*s)[i])
	}
	return st
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

func Wrap(err error) Error {
	if err == nil {
		return nil
	}
	return Err{
		cause: err,
		stack: callers().StackTrace(),
	}
}

func New(msg string) Error {
	return Err{
		cause: errors.New(msg),
		stack: callers().StackTrace(),
	}
}

func C(code int, msg string) Error {
	return Err{
		cause: errors.New(msg),
		code:  code,
		stack: callers().StackTrace(),
	}
}

func Cf(code int, format string, args ...interface{}) Error {
	return Err{
		cause: fmt.Errorf(format, args...),
		code:  code,
		stack: callers().StackTrace(),
	}
}
