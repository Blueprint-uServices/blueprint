package linux

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

// A utility function to deterministically convert a string into a
// a valid linux environment variable name.  This is done by converting
// all punctuation characters to underscores, and converting alphabetic
// characters to uppercase (for convention), e.g.
//
//	a.grpc_addr becomes A_GRPC_ADDR.
//
// Punctuation is converted to underscores, and alpha are made uppercase.
func EnvVar(name string) string {
	return strings.ToUpper(ir.CleanName(name))
}

// A utility function for use when using commands.
// Converts a string to a compatible command name.
// Punctuation is converted to underscores, and alpha are made uppercase.
func FuncName(name string) string {
	return strings.ToLower(ir.CleanName(name))
}
