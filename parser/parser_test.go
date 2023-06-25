package parser_test

import (
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/parser"
	"github.com/KSpaceer/fastyaml/token"
	"testing"
)

func TestParser(t *testing.T) {
	type tcase struct {
		tokens      []token.Token
		expectedAST ast.Node
	}

	tcases := []tcase{
		{
			tokens: []token.Token{
				{
					Type:   token.EOFType,
					Start:  token.Position{},
					End:    token.Position{},
					Origin: "",
				},
			},
			expectedAST: nil,
		},
	}

	for _, tc := range tcases {
		result := parser.Parse(&testTokenStream{
			tokens: tc.tokens,
			index:  0,
		})

	}
}

type testComparingVisitor struct {
}