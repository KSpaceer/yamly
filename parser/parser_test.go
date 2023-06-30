package parser_test

import (
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/ast/utils"
	"github.com/KSpaceer/fastyaml/parser"
	"github.com/KSpaceer/fastyaml/token"
	"strings"
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
			expectedAST: ast.NewStreamNode(token.Position{}, token.Position{}, nil),
		},
		{
			tokens: []token.Token{
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 4},
					Origin: "key",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 1, Column: 4},
					End:    token.Position{Row: 1, Column: 5},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 5},
					End:    token.Position{Row: 1, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 6},
					End:    token.Position{Row: 1, Column: 11},
					Origin: "value",
				},
				{
					Type: token.EOFType,
				},
			},
			expectedAST: ast.NewStreamNode(
				token.Position{},
				token.Position{},
				nil),
		},
	}
	cmp := utils.NewComparator()
	for _, tc := range tcases {
		result := parser.Parse(&testTokenStream{
			tokens: tc.tokens,
			index:  0,
		})
		if !cmp.Compare(tc.expectedAST, result) {
			printer := utils.NewPrinter()
			var s strings.Builder
			if err := printer.Print(tc.expectedAST, &s); err != nil {
				t.Fatalf("failed to print AST: %v", err)
			}
			expected := s.String()
			s.Reset()
			if err := printer.Print(result, &s); err != nil {
				t.Fatalf("failed to print AST: %v", err)
			}
			got := s.String()
			s.Reset()
			t.Errorf("AST are not equal:\n\nExpected:\n%s\n\nGot:\n%s\n", expected, got)
			t.Fail()
		}
	}
}
