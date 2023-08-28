package decode

import (
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
)

type denyError struct {
	expecter expecter
	nt       ast.NodeType
}

func (de *denyError) Error() string {
	return fmt.Sprintf("node %s was denied by expectancy rule %q", de.nt, de.expecter.name())
}

func (de *denyError) Is(err error) bool {
	_, ok := err.(*denyError)
	return ok
}

type AliasDereferenceError struct {
	name string
}

func (ade AliasDereferenceError) Error() string {
	return fmt.Sprintf("failed to dereference alias %q", ade.name)
}
