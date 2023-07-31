package parser

import (
	"fmt"
	"github.com/KSpaceer/yayamls/token"
)

type UnbalancedParenthesesError struct {
	isClosing bool
	tok       token.Token
}

func (u UnbalancedParenthesesError) Error() string {
	if u.isClosing {
		return fmt.Sprintf("parenthesis %q at position %s does not have preceding opening equivalent",
			u.tok.Origin, u.tok.Start)
	}
	return fmt.Sprintf("parentheses are not balanced")
}
