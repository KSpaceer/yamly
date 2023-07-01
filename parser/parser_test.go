package parser_test

import (
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/ast/astutils"
	"github.com/KSpaceer/fastyaml/parser"
	"github.com/KSpaceer/fastyaml/token"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	type tcase struct {
		name        string
		tokens      []token.Token
		expectedAST ast.Node
	}

	tcases := []tcase{
		{
			name: "empty YAML",
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
			name: "simple mapping entry",
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
				[]ast.Node{
					ast.NewCollectionNode(
						token.Position{},
						token.Position{},
						nil,
						ast.NewMappingNode(
							token.Position{},
							token.Position{},
							[]ast.Node{
								ast.NewMappingEntryNode(
									token.Position{},
									token.Position{},
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"key",
									),
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"value",
									),
								),
							},
						),
					),
				},
			),
		},
		{
			name: "simple sequence",
			tokens: []token.Token{
				{
					Type:   token.SequenceEntryType,
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "value1",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.SequenceEntryType,
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "value2",
				},
				{
					Type: token.EOFType,
				},
			},
			expectedAST: ast.NewStreamNode(
				token.Position{},
				token.Position{},
				[]ast.Node{
					ast.NewCollectionNode(
						token.Position{},
						token.Position{},
						nil,
						ast.NewSequenceNode(
							token.Position{},
							token.Position{},
							[]ast.Node{
								ast.NewTextNode(token.Position{}, token.Position{}, "value1"),
								ast.NewTextNode(token.Position{}, token.Position{}, "value2"),
							},
						),
					),
				},
			),
		},
		{
			name: "simple mapping with sequence and simple value",
			tokens: []token.Token{
				{
					Type:   token.StringType,
					Origin: "sequence",
				},
				{
					Type:   token.MappingValueType,
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "sequencevalue1",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "sequencevalue2",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.StringType,
					Origin: "simple",
				},
				{
					Type:   token.MappingValueType,
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "value",
				},
				{
					Type: token.EOFType,
				},
			},
			expectedAST: ast.NewStreamNode(
				token.Position{},
				token.Position{},
				[]ast.Node{
					ast.NewCollectionNode(
						token.Position{},
						token.Position{},
						nil,
						ast.NewMappingNode(
							token.Position{},
							token.Position{},
							[]ast.Node{
								ast.NewMappingEntryNode(
									token.Position{},
									token.Position{},
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"sequence",
									),
									ast.NewCollectionNode(
										token.Position{},
										token.Position{},
										nil,
										ast.NewSequenceNode(
											token.Position{},
											token.Position{},
											[]ast.Node{
												ast.NewTextNode(
													token.Position{},
													token.Position{},
													"sequencevalue1",
												),
												ast.NewTextNode(
													token.Position{},
													token.Position{},
													"sequencevalue2",
												),
											},
										),
									),
								),
								ast.NewMappingEntryNode(
									token.Position{},
									token.Position{},
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"simple",
									),
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"value",
									),
								),
							},
						),
					),
				},
			),
		},
	}
	cmp := astutils.NewComparator()
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.Parse(&testTokenStream{
				tokens: tc.tokens,
				index:  0,
			})
			if !cmp.Equal(tc.expectedAST, result) {
				printer := astutils.NewPrinter()
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
		})

	}
}
