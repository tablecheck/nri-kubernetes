package data

import (
	"fmt"
	"strings"
)

// ErrorGroup groups errors that can be recoverable (the execution can continue) or not
type ErrorGroup struct {
	Recoverable bool
	Errors      []error
}

// Append appends the errors passed as argument to the errors slice of the receiver object.
func (g *ErrorGroup) Append(errs ...error) {
	g.Errors = append(g.Errors, errs...)
}

// String shows a comma-separated string representation of all the error messages
func (g ErrorGroup) String() string {
	strs := make([]string, 0, len(g.Errors))
	for _, err := range g.Errors {
		strs = append(strs, err.Error())
	}
	var recoverable string
	if g.Recoverable {
		recoverable = "Recoverable"
	} else {
		recoverable = "Non-recoverable"
	}
	return fmt.Sprintf("%s error group: %s", recoverable, strings.Join(strs, ", "))
}
