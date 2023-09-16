package decode

import (
	"fmt"
)

type denyError struct {
	expecter    expecter
	nodeContent string
}

func (de *denyError) Error() string {
	return fmt.Sprintf(
		"node was denied by expectancy rule %q\nnode content: %s",
		de.expecter.name(),
		de.nodeContent,
	)
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
