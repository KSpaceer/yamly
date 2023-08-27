package decode

import (
	"errors"
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
)

var EndOfStream = errors.New("end of YAML stream")

type DenyError struct {
	expecter expecter
	nt       ast.NodeType
}

func (de *DenyError) Error() string {
	return fmt.Sprintf("node %s was denied by expectancy rule %q", de.nt, de.expecter.name())
}

func (de *DenyError) Is(err error) bool {
	_, ok := err.(*DenyError)
	return ok
}

type AliasDereferenceError struct {
	name string
}

func (ade AliasDereferenceError) Error() string {
	return fmt.Sprintf("failed to dereference alias %q", ade.name)
}
