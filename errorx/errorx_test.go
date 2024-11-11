package errorx

import (
	"errors"
	"fmt"
	"testing"
)

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}
func TestError(t *testing.T) {
	err := errors.New("test")
	errx := Wrap(err)
	_assert(errx.Cause() != nil, "test failed")
}
