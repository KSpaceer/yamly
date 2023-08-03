package parser_test

import (
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/ast/astutils"
	"github.com/KSpaceer/yayamls/parser"
	"github.com/KSpaceer/yayamls/token"
	"os"
	"strings"
	"testing"
)

func TestParseTokens(t *testing.T) {
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
					Origin: "",
				},
			},
			expectedAST: ast.NewStreamNode(nil),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("key"),
							ast.NewTextNode("value"),
						),
					},
				)),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewTextNode("value1"),
						ast.NewTextNode("value2"),
					},
				)),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("sequence"),
							ast.NewContentNode(
								nil,
								ast.NewSequenceNode(
									[]ast.Node{
										ast.NewTextNode("sequencevalue1"),
										ast.NewTextNode("sequencevalue2"),
									},
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("simple"),
							ast.NewTextNode("value"),
						),
					},
				)),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewMappingNode(
							[]ast.Node{
								ast.NewMappingEntryNode(
									ast.NewTextNode("key1"),
									ast.NewTextNode("value1"),
								),
								ast.NewMappingEntryNode(
									ast.NewTextNode("key2"),
									ast.NewTextNode("value2"),
								),
							},
						),
						ast.NewTextNode("quotedvalue"),
					},
				)),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("mapping"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("map"),
									ast.NewAnchorNode("ref"),
								),
								ast.NewMappingNode(
									[]ast.Node{
										ast.NewMappingEntryNode(
											ast.NewTextNode("innerkey"),
											ast.NewTextNode("innervalue"),
										),
									},
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("aliased"),
							ast.NewAliasNode("ref"),
						),
					},
				)),
			}),
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
					Type:   token.KeepChompingType,
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewContentNode(
							ast.NewPropertiesNode(
								nil,
								ast.NewAnchorNode("lit"),
							),
							ast.NewTextNode("firstrow\nsecondrow\n\n"),
						),
						ast.NewContentNode(
							ast.NewPropertiesNode(
								ast.NewTagNode("primary"),
								nil,
							),
							ast.NewTextNode("\nfolded"),
						),
					},
				)),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewNullNode(),
				ast.NewTextNode("aaaa "),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("mapping"),
							ast.NewContentNode(
								nil,
								ast.NewMappingNode(
									[]ast.Node{
										ast.NewMappingEntryNode(
											ast.NewContentNode(
												ast.NewInvalidNode(),
												ast.NewTextNode("quoted key"),
											),
											ast.NewNullNode(),
										),
										ast.NewMappingEntryNode(
											ast.NewNullNode(),
											ast.NewTextNode("value"),
										),
										ast.NewMappingEntryNode(
											ast.NewNullNode(),
											ast.NewNullNode(),
										),
									},
								),
							)),
						ast.NewMappingEntryNode(
							ast.NewTextNode("sequence"),
							ast.NewContentNode(
								nil,
								ast.NewSequenceNode(
									[]ast.Node{
										ast.NewNullNode(),
										ast.NewTextNode("seqvalue"),
									},
								),
							)),
					},
				)),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewSequenceNode(
					[]ast.Node{
						ast.NewTextNode("plain"),
						ast.NewTextNode("\\\"multi\\\"\n\nline"),
						ast.NewMappingEntryNode(
							ast.NewTextNode("flow"),
							ast.NewTextNode("pair"),
						),
						ast.NewMappingEntryNode(ast.NewNullNode(), ast.NewNullNode()),
					},
				),
			}),
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
					Origin: "multiline",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "\"ათჯერ",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "გაზომე",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "და",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "ერთხელ",
				},
				{
					Type:   token.SpaceType,
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Origin: "გაჭერი\"",
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("unquoted"),
							ast.NewTextNode(
								"''single quoted''\n\nmultiline \"ათჯერ გაზომე და ერთხელ გაჭერი\"",
							),
						),
						ast.NewMappingEntryNode(
							ast.NewContentNode(
								ast.NewInvalidNode(),
								ast.NewTextNode("explicit key"),
							),
							ast.NewTextNode("adjacent"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("novalue"),
							ast.NewNullNode(),
						),
						ast.NewMappingEntryNode(ast.NewNullNode(), ast.NewNullNode()),
						ast.NewMappingEntryNode(ast.NewNullNode(), ast.NewNullNode()),
					},
				),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewSequenceNode(
							[]ast.Node{
								ast.NewContentNode(
									ast.NewPropertiesNode(
										ast.NewTagNode("!baz"),
										ast.NewInvalidNode(),
									),
									ast.NewTextNode("entity"),
								),
								ast.NewTextNode("plain\nmulti line"),
							},
						),
						ast.NewContentNode(
							nil,
							ast.NewTextNode("\n\tspaced\n\n text"),
						),
					},
				)),
			}),
		},
		{
			name: "mix2",
			/*
				...
				<BOM> # stream comment
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewNullNode(),
				ast.NewContentNode(
					ast.NewPropertiesNode(
						ast.NewInvalidNode(),
						ast.NewAnchorNode("anchor"),
					),
					ast.NewNullNode(),
				),
				ast.NewSequenceNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewAliasNode("anchor"),
							ast.NewNullNode(),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("plain"),
							ast.NewSequenceNode(nil),
						),
						ast.NewMappingEntryNode(
							ast.NewNullNode(),
							ast.NewTextNode("emptyk"),
						),
						ast.NewMappingEntryNode(
							ast.NewContentNode(
								ast.NewInvalidNode(),
								ast.NewTextNode("adjacent"),
							),
							ast.NewTextNode("value"),
						),
					},
				),
			}),
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
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewTextNode("a"),
				ast.NewNullNode(),
			}),
		},
	}
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseTokens(tc.tokens)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			compareAST(t, tc.expectedAST, result)
		})

	}
}

func TestParseStringWithDefaultTokenStream(t *testing.T) {
	type tcase struct {
		name        string
		src         string
		expectedAST ast.Node
	}

	tcases := []tcase{
		{
			//https://netplan.readthedocs.io/en/stable/examples/
			name: "netplan example",
			src: `
                network:
                  version: 2
                  renderer: networkd
                  ethernets:
                    enp3s0:
                      addresses:
                        - 10.10.10.2/24
				`,
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil,
					ast.NewMappingNode([]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("network"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("version"),
										ast.NewTextNode("2"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("renderer"),
										ast.NewTextNode("networkd"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("ethernets"),
										ast.NewContentNode(nil,
											ast.NewMappingNode([]ast.Node{
												ast.NewMappingEntryNode(
													ast.NewTextNode("enp3s0"),
													ast.NewContentNode(nil,
														ast.NewMappingNode([]ast.Node{
															ast.NewMappingEntryNode(
																ast.NewTextNode("addresses"),
																ast.NewContentNode(nil,
																	ast.NewSequenceNode([]ast.Node{
																		ast.NewTextNode("10.10.10.2/24"),
																	}),
																),
															),
														}),
													),
												),
											}),
										),
									),
								}),
							),
						),
					}),
				),
			}),
		},
		{
			// https://kubernetes.io/docs/concepts/configuration/configmap/
			name: "k8s configmap",
			src: `
              apiVersion: v1
              kind: ConfigMap
              metadata: 
                name: game-demo
              data:
                # property-like keys; each key maps to a single value
                player_initial_lives: "3"
                ui_properties_file_name: "user-interface.properties"

                # file-like keys
                game.properties: |
                  enemy.types=aliens,monsters
                  player.maximum-lives=5
                user-interface.properties: |
                  color.good=purple
                  color.bad=yellow
                  allow.textmode=true
            `,
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil,
					ast.NewMappingNode([]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("apiVersion"),
							ast.NewTextNode("v1"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("kind"),
							ast.NewTextNode("ConfigMap"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("metadata"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("name"),
										ast.NewTextNode("game-demo"),
									),
								}),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("data"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("player_initial_lives"),
										ast.NewTextNode("3"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("ui_properties_file_name"),
										ast.NewTextNode("user-interface.properties"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("game.properties"),
										ast.NewContentNode(nil,
											ast.NewTextNode("enemy.types=aliens,monsters\nplayer.maximum-lives=5\n"),
										),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("user-interface.properties"),
										ast.NewContentNode(nil,
											ast.NewTextNode("color.good=purple\ncolor.bad=yellow\nallow.textmode=true\n"),
										),
									),
								}),
							),
						),
					}),
				),
			}),
		},
		{
			name: "tag and anchor in flow",
			src:  "{key: !!str &ref value, *ref: *ref}",
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewMappingNode([]ast.Node{
					ast.NewMappingEntryNode(
						ast.NewTextNode("key"),
						ast.NewContentNode(
							ast.NewPropertiesNode(
								ast.NewTagNode("str"),
								ast.NewAnchorNode("ref"),
							),
							ast.NewTextNode("value"),
						),
					),
					ast.NewMappingEntryNode(
						ast.NewAliasNode("ref"),
						ast.NewAliasNode("ref"),
					),
				}),
			}),
		},
		{
			// https://learnxinyminutes.com/docs/yaml/
			name: "large example",
			src: func() string {
				data, err := os.ReadFile("testdata/learnyaml.yaml")
				if err != nil {
					panic(err)
				}
				return string(data)
			}(),
			expectedAST: ast.NewStreamNode([]ast.Node{
				ast.NewContentNode(nil,
					ast.NewMappingNode([]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("key"),
							ast.NewTextNode("value"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("another_key"),
							ast.NewTextNode("Another value goes here."),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("a_number_value"),
							ast.NewTextNode("100"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("scientific_notation"),
							ast.NewTextNode("1e+12"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("hex_notation"),
							ast.NewTextNode("0x123"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("octal_notation"),
							ast.NewTextNode("0123"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("boolean"),
							ast.NewTextNode("true"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("null_value"),
							ast.NewTextNode("null"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("another_null_value"),
							ast.NewTextNode("~"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("key with spaces"),
							ast.NewTextNode("value"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("no"),
							ast.NewTextNode("no"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("yes"),
							ast.NewTextNode("No"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("not_enclosed"),
							ast.NewTextNode("yes"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("enclosed"),
							ast.NewTextNode("yes"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("however"),
							ast.NewTextNode("A string, enclosed in quotes."),
						),
						ast.NewMappingEntryNode(
							ast.NewContentNode(
								ast.NewInvalidNode(),
								ast.NewTextNode("Keys can be quoted too."),
							),
							ast.NewTextNode("Useful if you want to put a ':' in your key."),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("single quotes"),
							ast.NewTextNode("have ''one'' escape pattern"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("double quotes"),
							ast.NewTextNode(`have many: \", \0, \t, \u263A, \x0d\x0a == \r\n, and more.`),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("Superscript two"),
							ast.NewTextNode(`\u00B2`),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("special_characters"),
							ast.NewTextNode("[ John ] & { Jane } - <Doe>"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("literal_block"),
							ast.NewContentNode(nil,
								ast.NewTextNode("This entire block of text will be the value of the "+
									"'literal_block' key,\nwith line breaks being preserved.\n"+
									"\nThe literal continues until de-dented, and the leading indentation "+
									"is\nstripped.\n\n    Any lines that are 'more-indented' keep the rest "+
									"of their indentation -\n    these lines will be indented by 4 spaces.\n",
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("folded_style"),
							ast.NewContentNode(nil,
								ast.NewTextNode("This entire block of text will be the value of "+
									"'folded_style', but this time, all newlines will be replaced "+
									"with a single space.\nBlank lines, like above, are converted to a newline "+
									"character.\n\n    'More-indented' lines keep their newlines, too -\n\n"+
									"    this text will appear over two lines.\n",
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("literal_strip"),
							ast.NewContentNode(nil,
								ast.NewTextNode("This entire block of text will be the value of the "+
									"'literal_block' key,\nwith trailing blank line being stripped.",
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("block_strip"),
							ast.NewContentNode(nil,
								ast.NewTextNode("This entire block of text will be the value of "+
									"'folded_style', but this time, all newlines will be replaced with a "+
									"single space and trailing blank line being stripped.",
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("literal_keep"),
							ast.NewContentNode(nil,
								ast.NewTextNode("This entire block of text will be the value of the "+
									"'literal_block' key,\nwith trailing blank line being kept.\n\n"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("block_keep"),
							ast.NewContentNode(nil,
								ast.NewTextNode("This entire block of text will be the value of "+
									"'folded_style', but this time, all newlines will be replaced "+
									"with a single space and trailing blank line being kept.\n\n",
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("a_nested_map"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("key"),
										ast.NewTextNode("value"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("another_key"),
										ast.NewTextNode("Another Value"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("another_nested_map"),
										ast.NewContentNode(nil,
											ast.NewMappingNode([]ast.Node{
												ast.NewMappingEntryNode(
													ast.NewTextNode("hello"),
													ast.NewTextNode("hello"),
												),
											}),
										),
									),
								}),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("0.25"),
							ast.NewTextNode("a float key"),
						),
						ast.NewMappingEntryNode(
							ast.NewContentNode(nil,
								ast.NewTextNode("This is a key\nthat has multiple lines\n"),
							),
							ast.NewTextNode("and this is its value"),
						),
						ast.NewMappingEntryNode(
							ast.NewSequenceNode([]ast.Node{
								ast.NewTextNode("Manchester United"),
								ast.NewTextNode("Real Madrid"),
							}),
							ast.NewSequenceNode([]ast.Node{
								ast.NewTextNode("2001-01-01"),
								ast.NewTextNode("2002-02-02"),
							}),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("a_sequence"),
							ast.NewContentNode(nil,
								ast.NewSequenceNode([]ast.Node{
									ast.NewTextNode("Item 1"),
									ast.NewTextNode("Item 2"),
									ast.NewTextNode("0.5"),
									ast.NewTextNode("Item 4"),
									ast.NewMappingNode([]ast.Node{
										ast.NewMappingEntryNode(
											ast.NewTextNode("key"),
											ast.NewTextNode("value"),
										),
										ast.NewMappingEntryNode(
											ast.NewTextNode("another_key"),
											ast.NewTextNode("another_value"),
										),
									}),
									ast.NewSequenceNode([]ast.Node{
										ast.NewTextNode("This is a sequence"),
										ast.NewTextNode("inside another sequence"),
									}),
									ast.NewSequenceNode([]ast.Node{
										ast.NewSequenceNode([]ast.Node{
											ast.NewTextNode("Nested sequence indicators"),
											ast.NewTextNode("can be collapsed"),
										}),
									}),
								}),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("json_map"),
							ast.NewMappingNode([]ast.Node{
								ast.NewMappingEntryNode(
									ast.NewContentNode(
										ast.NewInvalidNode(),
										ast.NewTextNode("key"),
									),
									ast.NewTextNode("value"),
								),
							}),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("json_seq"),
							ast.NewSequenceNode([]ast.Node{
								ast.NewTextNode("3"),
								ast.NewTextNode("2"),
								ast.NewTextNode("1"),
								ast.NewTextNode("takeoff"),
							}),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("and quotes are optional"),
							ast.NewMappingNode([]ast.Node{
								ast.NewMappingEntryNode(
									ast.NewTextNode("key"),
									ast.NewSequenceNode([]ast.Node{
										ast.NewTextNode("3"),
										ast.NewTextNode("2"),
										ast.NewTextNode("1"),
										ast.NewTextNode("takeoff"),
									}),
								),
							}),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("anchored_content"),
							ast.NewContentNode(
								ast.NewPropertiesNode(nil,
									ast.NewAnchorNode("anchor_name"),
								),
								ast.NewTextNode("This string will appear as the value of two keys."),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("other_anchor"),
							ast.NewAliasNode("anchor_name"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("base"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewInvalidNode(),
									ast.NewAnchorNode("base"),
								),
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("name"),
										ast.NewTextNode("Everyone has same name"),
									),
								}),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("foo"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("<<"),
										ast.NewAliasNode("base"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("age"),
										ast.NewTextNode("10"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("name"),
										ast.NewTextNode("John"),
									),
								}),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("bar"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("<<"),
										ast.NewAliasNode("base"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("age"),
										ast.NewTextNode("20"),
									),
								}),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("explicit_boolean"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("bool"),
									ast.NewInvalidNode(),
								),
								ast.NewTextNode("true"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("explicit_integer"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("int"),
									ast.NewInvalidNode(),
								),
								ast.NewTextNode("42"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("explicit_float"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("float"),
									ast.NewInvalidNode(),
								),
								ast.NewTextNode("-42.24"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("explicit_string"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("str"),
									ast.NewInvalidNode(),
								),
								ast.NewTextNode("0.5"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("explicit_datetime"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("timestamp"),
									ast.NewInvalidNode(),
								),
								ast.NewTextNode("2022-11-17 12:34:56.78 +9"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("explicit_null"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("null"),
									nil,
								),
								ast.NewTextNode("null"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("python_complex_number"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("python/complex"),
									ast.NewInvalidNode(),
								),
								ast.NewTextNode("1+2j"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("python/tuple"),
									ast.NewInvalidNode(),
								),
								ast.NewSequenceNode([]ast.Node{
									ast.NewTextNode("5"),
									ast.NewTextNode("7"),
								}),
							),
							ast.NewTextNode("Fifty Seven"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("datetime_canonical"),
							ast.NewTextNode("2001-12-15T02:59:43.1Z"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("datetime_space_separated_with_time_zone"),
							ast.NewTextNode("2001-12-14 21:59:43.10 -5"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("date_implicit"),
							ast.NewTextNode("2002-12-14"),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("date_explicit"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("timestamp"),
									nil,
								),
								ast.NewTextNode("2002-12-14"),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("gif_file"),
							ast.NewContentNode(
								ast.NewPropertiesNode(
									ast.NewTagNode("binary"),
									ast.NewInvalidNode(),
								),
								ast.NewTextNode("R0lGODlhDAAMAIQAAP//9/X17unp5WZmZgAAAOfn515eXvPz7Y6OjuDg4J+fn5\n"+
									"OTk6enp56enmlpaWNjY6Ojo4SEhP/++f/++f/++f/++f/++f/++f/++f/++f/+\n"+
									"+f/++f/++f/++f/++f/++SH+Dk1hZGUgd2l0aCBHSU1QACwAAAAADAAMAAAFLC\n"+
									"AgjoEwnuNAFOhpEMTRiggcz4BNJHrv/zCFcLiwMWYNG84BwwEeECcgggoBADs=\n",
								),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("set"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("item1"),
										ast.NewNullNode(),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("item2"),
										ast.NewNullNode(),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("item3"),
										ast.NewNullNode(),
									),
								}),
							),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("or"),
							ast.NewMappingNode([]ast.Node{
								ast.NewMappingEntryNode(
									ast.NewTextNode("item1"),
									ast.NewNullNode(),
								),
								ast.NewMappingEntryNode(
									ast.NewTextNode("item2"),
									ast.NewNullNode(),
								),
								ast.NewMappingEntryNode(
									ast.NewTextNode("item3"),
									ast.NewNullNode(),
								),
							}),
						),
						ast.NewMappingEntryNode(
							ast.NewTextNode("set2"),
							ast.NewContentNode(nil,
								ast.NewMappingNode([]ast.Node{
									ast.NewMappingEntryNode(
										ast.NewTextNode("item1"),
										ast.NewTextNode("null"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("item2"),
										ast.NewTextNode("null"),
									),
									ast.NewMappingEntryNode(
										ast.NewTextNode("item3"),
										ast.NewTextNode("null"),
									),
								}),
							),
						),
					}),
				),
			}),
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.src)
			if err == nil {
				compareAST(t, tc.expectedAST, result)
			} else {
				t.Errorf("unexpected error: %s", err)
			}
		})

	}
}

func TestAboba(t *testing.T) {
	parser.ParseString("[[[[[[[[[[[[[[[[[[[[[[!=>")
}

func FuzzParseString(f *testing.F) {
	seeds := []string{
		"key:key",
		"hello",
		"key: value",
		"- value\n- value",
		"key: |\n  value\n  value",
		"key: >\n  value\n  value",
		"{\"key\":'value'}",
		"[value,value,value]",
		func() string {
			data, err := os.ReadFile("testdata/learnyaml.yaml")
			if err != nil {
				panic(err)
			}
			return string(data)
		}(),
	}
	for i := range seeds {
		f.Add(seeds[i])
	}
	f.Fuzz(func(t *testing.T, src string) {
		parser.ParseString(src)
	})
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
