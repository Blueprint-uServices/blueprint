package blueprint

import (
	"errors"
	"fmt"
	"runtime"
)

const (
	MAX_ERR_SIZE = 2048
)

type BlueprintError struct {
	Stack []byte
	Err   error
}

func (e *BlueprintError) Error() string {
	return fmt.Sprintf("%v\n%v", e.Err, string(e.Stack))
}

func NewBlueprintError(msg string) error {
	bytes := make([]byte, MAX_ERR_SIZE)
	runtime.Stack(bytes, false)
	return &BlueprintError{
		Stack: bytes,
		Err:   errors.New(msg),
	}
}

func Errorf(format string, a ...any) error {
	bytes := make([]byte, MAX_ERR_SIZE)
	runtime.Stack(bytes, false)
	err := fmt.Errorf(format, a...)
	return &BlueprintError{
		Stack: bytes,
		Err:   err,
	}
}
