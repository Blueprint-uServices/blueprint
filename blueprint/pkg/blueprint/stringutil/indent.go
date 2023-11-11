// Package stringutil implements utility methods for common string manipulations used
// by Blueprint plugins.
package stringutil

import (
	"strings"
)

// Indents str by the specified amount by prepending whitespace (space characters)
// at the beginning of every line.  str can be a multi-line string; whitespace will
// be prepended to every line.
func Indent(str string, amount int) string {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", amount) + line
	}
	return strings.Join(lines, "\n")
}

// Indents str by the specified amount by adding or removing whitespace (space characters)
// at the beginning of every line.  If str is already indented by some amount, then this
// method will add or remove whitespace so that the result is indented by amount total whitespace.
func Reindent(str string, amount int) string {
	splits := strings.Split(str, "\n")
	for i := range splits {
		if strings.TrimSpace(splits[i]) == "" {
			splits[i] = ""
		} else {
			splits[i] = strings.Replace(splits[i], "\t", "    ", -1)
		}
	}

	currentIndent := 10000
	for i := range splits {
		if len(splits[i]) > 0 {
			indent := len(splits[i]) - len(strings.TrimLeft(splits[i], " "))
			if indent < currentIndent {
				currentIndent = indent
			}
		}
	}

	prefix := strings.Repeat(" ", amount)
	for i := range splits {
		if len(splits[i]) > 0 {
			splits[i] = prefix + splits[i][currentIndent:]
		}
	}

	return strings.Join(splits, "\n")
}
