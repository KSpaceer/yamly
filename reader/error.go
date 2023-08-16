package reader

import (
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
)

type DenyError struct {
	expecter expecter
	nt       ast.NodeType
}

func (de *DenyError) Error() string {
	return fmt.Sprintf("node %s was denied by expectancy rule %q", de.nt, de.expecter)
}
