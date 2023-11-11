package blueprint

import (
	"fmt"
	"runtime"
)

const (
	MAX_ERR_SIZE = 2048
)

// An error used by Blueprint's compiler that captures the calling
// stack so that we can tie errors back to the plugins that caused
// the error
type blueprintError struct {
	Stack []byte
	Err   error
}

func (e *blueprintError) Error() string {
	return fmt.Sprintf("%v\n%v", e.Err, string(e.Stack))
}

// Generates an error in the same way as fmt.Errorf but also includes
// the call stack.
//
// Plugins should generally use this method, because it enables us
// to more easily tie errors back to the plugins and wiring specs
// that caused the error.
func Errorf(format string, a ...any) error {
	bytes := make([]byte, MAX_ERR_SIZE)
	runtime.Stack(bytes, false)
	err := fmt.Errorf(format, a...)
	return &blueprintError{
		Stack: bytes,
		Err:   err,
	}
}
