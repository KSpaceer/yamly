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

	var tcases = []tcase{
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
		{
			name: "simple sequence with mapping and simple single quoted value",
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
					Origin: "key1",
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
					Origin: "value1",
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
					Type:   token.StringType,
					Origin: "key2",
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
					Origin: "value2",
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
					Type:   token.SingleQuoteType,
					Origin: "'",
				},
				{
					Type:   token.StringType,
					Origin: "quotedvalue",
				},
				{
					Type:   token.SingleQuoteType,
					Origin: "'",
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
												"key1",
											),
											ast.NewTextNode(
												token.Position{},
												token.Position{},
												"value1",
											),
										),
										ast.NewMappingEntryNode(
											token.Position{},
											token.Position{},
											ast.NewTextNode(
												token.Position{},
												token.Position{},
												"key2",
											),
											ast.NewTextNode(
												token.Position{},
												token.Position{},
												"value2",
											),
										),
									},
								),
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"quotedvalue",
								),
							},
						),
					),
				},
			),
		},
		{
			name: "nested mapping with properties",
			tokens: []token.Token{
				{
					Type:   token.StringType,
					Origin: "mapping",
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
					Type:   token.TagType,
					Origin: "!",
				},
				{
					Type:   token.TagType,
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Origin: "map",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.AnchorType,
					Origin: "&",
				},
				{
					Type:   token.StringType,
					Origin: "ref",
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
					Type:   token.MappingKeyType,
					Origin: "?",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "innerkey",
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
					Type:   token.MappingValueType,
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "innervalue",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.StringType,
					Origin: "aliased",
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
					Type:   token.AliasType,
					Origin: "*",
				},
				{
					Type:   token.StringType,
					Origin: "ref",
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
										"mapping",
									),
									ast.NewCollectionNode(
										token.Position{},
										token.Position{},
										ast.NewPropertiesNode(
											token.Position{},
											token.Position{},
											ast.NewTagNode(
												token.Position{},
												token.Position{},
												"map",
											),
											ast.NewAnchorNode(
												token.Position{},
												token.Position{},
												"ref",
											),
										),
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
														"innerkey",
													),
													ast.NewTextNode(
														token.Position{},
														token.Position{},
														"innervalue",
													),
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
										"aliased",
									),
									ast.NewAliasNode(
										token.Position{},
										token.Position{},
										"ref",
									),
								),
							},
						),
					),
				},
			),
		},
		{
			name: "sequence with folded and literal",
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
					Type:   token.AnchorType,
					Origin: "&",
				},
				{
					Type:   token.StringType,
					Origin: "lit",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.LiteralType,
					Origin: "|",
				},
				{
					Type:   token.PlusType,
					Origin: "+",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "my_comment",
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
					Type:   token.StringType,
					Origin: "firstrow",
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
					Type:   token.StringType,
					Origin: "secondrow",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
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
					Type:   token.TagType,
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Origin: "primary",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.FoldedType,
					Origin: ">",
				},
				{
					Type:   token.StringType,
					Origin: "1",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
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
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "folded",
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
								ast.NewScalarNode(
									token.Position{},
									token.Position{},
									ast.NewPropertiesNode(
										token.Position{},
										token.Position{},
										nil,
										ast.NewAnchorNode(
											token.Position{},
											token.Position{},
											"lit",
										),
									),
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"firstrow\nsecondrow\n",
									),
								),
								ast.NewScalarNode(
									token.Position{},
									token.Position{},
									ast.NewPropertiesNode(
										token.Position{},
										token.Position{},
										ast.NewTagNode(
											token.Position{},
											token.Position{},
											"primary",
										),
										nil,
									),
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"\nfolded",
									),
								),
							},
						),
					),
				},
			),
		},
		{
			name: "several documents with comments",
			tokens: []token.Token{
				{
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Origin: "directives",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "comment",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.DirectiveType,
					Origin: "%",
				},
				{
					Type:   token.StringType,
					Origin: token.YAMLDirective,
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "2.2",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.DirectiveType,
					Origin: "%",
				},
				{
					Type:   token.StringType,
					Origin: token.TagDirective,
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.TagType,
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Origin: "yaml",
				},
				{
					Type:   token.TagType,
					Origin: "!",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "tag:yaml.org,2002:",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.DirectiveEndType,
					Origin: "---",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.DocumentEndType,
					Origin: "...",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.DoubleQuoteType,
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Origin: "aaaa",
				},
				{
					Type:   token.DoubleQuoteType,
					Origin: "\"",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type: token.EOFType,
				},
			},
			expectedAST: ast.NewStreamNode(
				token.Position{},
				token.Position{},
				[]ast.Node{
					ast.NewNullNode(token.Position{}),
					ast.NewTextNode(token.Position{}, token.Position{}, "aaaa"),
				},
			),
		},
		{
			name: "null nodes",
			tokens: []token.Token{
				{
					Type:   token.DirectiveEndType,
					Origin: "---",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.StringType,
					Origin: "mapping",
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
					Type:   token.DoubleQuoteType,
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Origin: "quoted",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "key",
				},
				{
					Type:   token.DoubleQuoteType,
					Origin: "\"",
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
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Origin: "empty",
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
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Origin: "empty",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "key",
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
					Type:   token.MappingKeyType,
					Origin: "?",
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
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Origin: "empty",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "key",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "and",
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
					Type:   token.MappingKeyType,
					Origin: "?",
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
					Type:   token.MappingValueType,
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
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
					Origin: "seqvalue",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.DocumentEndType,
					Origin: "...",
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
										"mapping",
									),
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
													ast.NewScalarNode(
														token.Position{},
														token.Position{},
														ast.NewInvalidNode(
															token.Position{},
															token.Position{},
														),
														ast.NewTextNode(
															token.Position{},
															token.Position{},
															"quoted key",
														),
													),
													ast.NewNullNode(token.Position{}),
												),
												ast.NewMappingEntryNode(
													token.Position{},
													token.Position{},
													ast.NewNullNode(token.Position{}),
													ast.NewTextNode(
														token.Position{},
														token.Position{},
														"value",
													),
												),
												ast.NewMappingEntryNode(
													token.Position{},
													token.Position{},
													ast.NewNullNode(token.Position{}),
													ast.NewNullNode(token.Position{}),
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
												ast.NewNullNode(token.Position{}),
												ast.NewTextNode(
													token.Position{},
													token.Position{},
													"seqvalue",
												),
											},
										),
									),
								),
							},
						),
					),
					ast.NewNullNode(token.Position{}),
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
