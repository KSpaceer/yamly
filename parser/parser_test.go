package parser_test

import (
	"github.com/KSpaceer/yayamls/ast"
	"github.com/KSpaceer/yayamls/ast/astutils"
	"github.com/KSpaceer/yayamls/parser"
	"github.com/KSpaceer/yayamls/token"
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
				ast.NewCollectionNode(nil, ast.NewMappingNode(
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
				ast.NewCollectionNode(nil, ast.NewSequenceNode(
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
				ast.NewCollectionNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("sequence"),
							ast.NewCollectionNode(
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
				ast.NewCollectionNode(nil, ast.NewSequenceNode(
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
				ast.NewCollectionNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("mapping"),
							ast.NewCollectionNode(
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
				ast.NewCollectionNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewScalarNode(
							ast.NewPropertiesNode(
								nil,
								ast.NewAnchorNode("lit"),
							),
							ast.NewTextNode("firstrow\nsecondrow\n"),
						),
						ast.NewScalarNode(
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
				ast.NewCollectionNode(nil, ast.NewMappingNode(
					[]ast.Node{
						ast.NewMappingEntryNode(
							ast.NewTextNode("mapping"),
							ast.NewCollectionNode(
								nil,
								ast.NewMappingNode(
									[]ast.Node{
										ast.NewMappingEntryNode(
											ast.NewScalarNode(
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
							ast.NewCollectionNode(
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
							ast.NewScalarNode(
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
				ast.NewCollectionNode(nil, ast.NewSequenceNode(
					[]ast.Node{
						ast.NewSequenceNode(
							[]ast.Node{
								ast.NewScalarNode(
									ast.NewPropertiesNode(
										ast.NewTagNode("!baz"),
										ast.NewInvalidNode(),
									),
									ast.NewTextNode("entity"),
								),
								ast.NewTextNode("plain\nmulti line"),
							},
						),
						ast.NewScalarNode(
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
				ast.NewScalarNode(
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
							ast.NewScalarNode(
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

	expectedAST := ast.NewStreamNode(nil)

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

type testTokenStream struct {
	tokens []token.Token
	index  int
}

func (t *testTokenStream) Next() token.Token {
	if t.index >= len(t.tokens) {
		return token.Token{Type: token.EOFType}
	}
	tok := t.tokens[t.index]
	t.index++
	return tok
}

func (t *testTokenStream) SetRawMode() {}

func (*testTokenStream) UnsetRawMode() {}
