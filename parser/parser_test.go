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
			/*
				key: value
			*/
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
			/*
				- value1
				- value2
			*/
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
			/*
				sequence:
				  - sequencevalue1
				  - sequencevalue2
				simple: value
			*/
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
			/*
				- key1: value1
				  key2: value2
				- 'quotedvalue'
			*/
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
			/*
				mapping: !!map &ref
				 ? innerkey
				 : innervalue
				aliased: *ref
			*/
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
			/*
				- &lit |+ # my_comment
				  firstrow
				  secondrow

				- !primary >1

				   folded
			*/
			tokens: []token.Token{
				{
					Type:   token.SequenceEntryType,
					Origin: "-",
				},
				{
					Type: token.SpaceType, Origin: " ",
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
			/*
				#directives comment
				%YAML 2.2
				%TAG !yaml! tag:yaml.org,2002:
				---
				...
				"aaaa \
				"

			*/
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
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "\\",
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
					ast.NewTextNode(token.Position{}, token.Position{}, "aaaa "),
				},
			),
		},
		{
			name: "null nodes",
			/*
				---
				mapping:
				  "quoted key": #empty value
				#empty key
				  ?
				  : value
				#empty key and value
				  ?
				  :
				sequence:
				  -
				  - seqvalue
				...
			*/
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
				},
			),
		},
		{
			name: "flow syntax sequence",
			/*
				[	plain	,"\"multi\"


				  line",flow: pair,	? ]
			*/
			tokens: []token.Token{
				{
					Type:   token.SequenceStartType,
					Origin: "[",
				},
				{
					Type:   token.TabType,
					Origin: "\t",
				},
				{
					Type:   token.StringType,
					Origin: "plain",
				},
				{
					Type:   token.TabType,
					Origin: "\t",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.DoubleQuoteType,
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Origin: "\\\"multi\\\"",
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
					Origin: "line",
				},
				{
					Type:   token.DoubleQuoteType,
					Origin: "\"",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.StringType,
					Origin: "flow",
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
					Origin: "pair",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.TabType,
					Origin: "\t",
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
					Type:   token.SequenceEndType,
					Origin: "]",
				},
				{
					Type: token.EOFType,
				},
			},
			expectedAST: ast.NewStreamNode(
				token.Position{},
				token.Position{},
				[]ast.Node{
					ast.NewSequenceNode(
						token.Position{},
						token.Position{},
						[]ast.Node{
							ast.NewTextNode(
								token.Position{},
								token.Position{},
								"plain",
							),
							ast.NewTextNode(
								token.Position{},
								token.Position{},
								"\\\"multi\\\"\n\nline",
							),
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"flow",
								),
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"pair",
								),
							),
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewNullNode(
									token.Position{},
								),
								ast.NewNullNode(
									token.Position{},
								),
							),
						},
					),
				},
			),
		},
		{
			name: "flow syntax mapping",
			/*
				{ unquoted:	'''single quoted''


				  multiline "ათჯერ გაზომე და ერთხელ გაჭერი"',? "explicit key":adjacent,novalue,:,?	}
			*/
			tokens: []token.Token{
				{
					Type:   token.MappingStartType,
					Origin: "{",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "unquoted",
				},
				{
					Type:   token.MappingValueType,
					Origin: ":",
				},
				{
					Type:   token.TabType,
					Origin: "\t",
				},
				{
					Type:   token.SingleQuoteType,
					Origin: "'",
				},
				{
					Type:   token.StringType,
					Origin: "''single quoted''",
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
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
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
					Origin: "multiline \"ათჯერ გაზომე და ერთხელ გაჭერი\"",
				},
				{
					Type:   token.SingleQuoteType,
					Origin: "'",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
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
					Type:   token.DoubleQuoteType,
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Origin: "explicit key",
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
					Type:   token.StringType,
					Origin: "adjacent",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.StringType,
					Origin: "novalue",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.MappingValueType,
					Origin: ":",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.MappingKeyType,
					Origin: "?",
				},
				{
					Type:   token.TabType,
					Origin: "\t",
				},

				{
					Type:   token.MappingEndType,
					Origin: "}",
				},
			},
			expectedAST: ast.NewStreamNode(
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
									"unquoted",
								),
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"''single quoted''\n\nmultiline \"ათჯერ გაზომე და ერთხელ გაჭერი\"",
								),
							),
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
										"explicit key",
									),
								),
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"adjacent",
								),
							),
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"novalue",
								),
								ast.NewNullNode(token.Position{}),
							),
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewNullNode(token.Position{}),
								ast.NewNullNode(token.Position{}),
							),
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewNullNode(token.Position{}),
								ast.NewNullNode(token.Position{}),
							),
						},
					),
				},
			),
		},
		{
			name: "mix",
			/*
				%RESERVED PARAMETER
				#directivecomment
				%TAG ! !local-tag-
				---
				-  -	!<!baz> entity
				   - plain

				     multi
				     line
				- >-

				  	spaced
				   text

				#trail
				#comments
			*/
			tokens: []token.Token{
				{
					Type:   token.DirectiveType,
					Origin: "%",
				},
				{
					Type:   token.StringType,
					Origin: "RESERVED",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "PARAMETER",
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
					Origin: "directivecomment",
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
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.TagType,
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Origin: "local-tag-",
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
					Type:   token.SequenceEntryType,
					Origin: "-",
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
					Type:   token.TabType,
					Origin: "\t",
				},
				{
					Type:   token.TagType,
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Origin: "<!baz>",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "entity",
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
					Type:   token.SequenceEntryType,
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "plain",
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
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "multi",
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
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "line",
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
					Type:   token.FoldedType,
					Origin: ">",
				},
				{
					Type:   token.StripChompingType,
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
					Type:   token.TabType,
					Origin: "\t",
				},
				{
					Type:   token.StringType,
					Origin: "spaced",
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
					Origin: "text",
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
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Origin: "trail",
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
					Origin: "comments",
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
												ast.NewTagNode(
													token.Position{},
													token.Position{},
													"!baz",
												),
												ast.NewInvalidNode(
													token.Position{},
													token.Position{},
												),
											),
											ast.NewTextNode(
												token.Position{},
												token.Position{},
												"entity",
											),
										),
										ast.NewTextNode(
											token.Position{},
											token.Position{},
											"plain\nmulti line",
										),
									},
								),
								ast.NewScalarNode(
									token.Position{},
									token.Position{},
									nil,
									ast.NewTextNode(
										token.Position{},
										token.Position{},
										"\n\tspaced\n\n text",
									),
								),
							},
						),
					),
				},
			),
		},
		{
			name: "mix2",
			/*
				...
				<BOM># stream comment
				---
				# document
				...
				&anchor
				...
				[*anchor : , plain: [],: emptyk,"adjacent":value]

			*/
			tokens: []token.Token{
				{
					Type:   token.DocumentEndType,
					Origin: "...",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.BOMType,
					Origin: "\uFEFF",
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
					Origin: "stream",
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
					Type:   token.DirectiveEndType,
					Origin: "---",
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
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "document",
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
					Type:   token.AnchorType,
					Origin: "&",
				},
				{
					Type:   token.StringType,
					Origin: "anchor",
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
					Type:   token.DocumentEndType,
					Origin: "...",
				},
				{
					Type:   token.LineBreakType,
					Origin: "\n",
				},
				{
					Type:   token.SequenceStartType,
					Origin: "[",
				},
				{
					Type:   token.AliasType,
					Origin: "*",
				},
				{
					Type:   token.StringType,
					Origin: "anchor",
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
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "plain",
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
					Type:   token.SequenceStartType,
					Origin: "[",
				},
				{
					Type:   token.SequenceEndType,
					Origin: "]",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
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
					Origin: "emptyk",
				},
				{
					Type:   token.CollectEntryType,
					Origin: ",",
				},
				{
					Type:   token.DoubleQuoteType,
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Origin: "adjacent",
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
					Type:   token.StringType,
					Origin: "value",
				},
				{
					Type:   token.SequenceEndType,
					Origin: "]",
				},
			},
			expectedAST: ast.NewStreamNode(
				token.Position{},
				token.Position{},
				[]ast.Node{
					ast.NewNullNode(token.Position{}),
					ast.NewScalarNode(
						token.Position{},
						token.Position{},
						ast.NewPropertiesNode(
							token.Position{},
							token.Position{},
							ast.NewInvalidNode(
								token.Position{},
								token.Position{},
							),
							ast.NewAnchorNode(
								token.Position{},
								token.Position{},
								"anchor",
							),
						),
						ast.NewNullNode(token.Position{}),
					),
					ast.NewSequenceNode(
						token.Position{},
						token.Position{},
						[]ast.Node{
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewAliasNode(
									token.Position{},
									token.Position{},
									"anchor",
								),
								ast.NewNullNode(token.Position{}),
							),
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"plain",
								),
								ast.NewSequenceNode(
									token.Position{},
									token.Position{},
									nil,
								),
							),
							ast.NewMappingEntryNode(
								token.Position{},
								token.Position{},
								ast.NewNullNode(token.Position{}),
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"emptyk",
								),
							),
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
										"adjacent",
									),
								),
								ast.NewTextNode(
									token.Position{},
									token.Position{},
									"value",
								),
							),
						},
					),
				},
			),
		},
		{
			name: "document with BOMs",
			tokens: []token.Token{
				{
					Type:   token.StringType,
					Origin: "a",
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
					Type:   token.BOMType,
					Origin: "\uFEFF",
				},
				{
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Origin: "comment",
				},
				{
					Type: token.EOFType,
				},
			},
			expectedAST: ast.NewStreamNode(
				token.Position{},
				token.Position{},
				[]ast.Node{
					ast.NewTextNode(
						token.Position{},
						token.Position{},
						"a",
					),
					ast.NewNullNode(token.Position{}),
				},
			),
		},
	}
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.Parse(&testTokenStream{
				tokens: tc.tokens,
				index:  0,
			})
			compareAST(t, tc.expectedAST, result)
		})

	}
}

func TestParserInvalidDocuments(t *testing.T) {
	t.Skip()

	type tcase struct {
		name   string
		tokens []token.Token
	}

	tcases := []tcase{
		{
			name: "broken explicit document",
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
					Type:   token.CommentType,
					Origin: "#",
				},
				{
					Type:   token.BOMType,
					Origin: "\uFEFF",
				},
			},
		},
	}

	expectedAST := ast.NewStreamNode(
		token.Position{},
		token.Position{},
		nil,
	)

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.Parse(&testTokenStream{
				tokens: tc.tokens,
				index:  0,
			})
			compareAST(t, expectedAST, result)
		})
	}

}

func compareAST(t *testing.T, expectedAST, gotAST ast.Node) {
	t.Helper()

	cmp := astutils.NewComparator()
	printer := astutils.NewPrinter()

	if !cmp.Equal(expectedAST, gotAST) {
		var s strings.Builder
		if err := printer.Print(expectedAST, &s); err != nil {
			t.Fatalf("failed to print AST: %v", err)
		}
		expected := s.String()
		s.Reset()
		if err := printer.Print(gotAST, &s); err != nil {
			t.Fatalf("failed to print AST: %v", err)
		}
		got := s.String()
		s.Reset()
		t.Errorf("AST are not equal:\n\nExpected:\n%s\n\nGot:\n%s\n", expected, got)
		t.Fail()
	}
}
