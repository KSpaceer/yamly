package lexer_test

import (
	"github.com/KSpaceer/yamly/engines/yayamls/lexer"
	"github.com/KSpaceer/yamly/engines/yayamls/token"
	"testing"
)

func TestTokenizer(t *testing.T) {
	type tcase struct {
		name                 string
		src                  string
		expectedTokens       []token.Token
		rawModEnableIndices  []int
		rawModDisableIndices []int
	}

	tcases := []tcase{
		{
			name: "empty YAML",
			src:  "",
			expectedTokens: []token.Token{
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 1, Column: 1},
					End:   token.Position{Row: 1, Column: 1},
				},
			},
		},
		{
			name: "simple mapping entry",
			src:  "key: value",
			expectedTokens: []token.Token{
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 3},
					Origin: "key",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 1, Column: 4},
					End:    token.Position{Row: 1, Column: 4},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 5},
					End:    token.Position{Row: 1, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 6},
					End:    token.Position{Row: 1, Column: 10},
					Origin: "value",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 1, Column: 11},
					End:   token.Position{Row: 1, Column: 11},
				},
			},
		},
		{
			name: "simple sequence",
			src:  "- value1\n- value2",
			expectedTokens: []token.Token{
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 2},
					End:    token.Position{Row: 1, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 3},
					End:    token.Position{Row: 1, Column: 8},
					Origin: "value1",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 9},
					End:    token.Position{Row: 1, Column: 9},
					Origin: "\n",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 3},
					End:    token.Position{Row: 2, Column: 8},
					Origin: "value2",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 2, Column: 9},
					End:   token.Position{Row: 2, Column: 9},
				},
			},
		},
		{
			name: "simple mapping with sequence and simple value",
			src:  "sequence:\n  - sequencevalue1\n  - sequencevalue2\nsimple: value",
			expectedTokens: []token.Token{
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 8},
					Origin: "sequence",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 1, Column: 9},
					End:    token.Position{Row: 1, Column: 9},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 10},
					End:    token.Position{Row: 1, Column: 10},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 2, Column: 3},
					End:    token.Position{Row: 2, Column: 3},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 4},
					End:    token.Position{Row: 2, Column: 4},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 5},
					End:    token.Position{Row: 2, Column: 18},
					Origin: "sequencevalue1",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 19},
					End:    token.Position{Row: 2, Column: 19},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 3, Column: 3},
					End:    token.Position{Row: 3, Column: 3},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 4},
					End:    token.Position{Row: 3, Column: 4},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 5},
					End:    token.Position{Row: 3, Column: 18},
					Origin: "sequencevalue2",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 19},
					End:    token.Position{Row: 3, Column: 19},
					Origin: "\n",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 6},
					Origin: "simple",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 4, Column: 7},
					End:    token.Position{Row: 4, Column: 7},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 8},
					End:    token.Position{Row: 4, Column: 8},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 9},
					End:    token.Position{Row: 4, Column: 13},
					Origin: "value",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 4, Column: 14},
					End:   token.Position{Row: 4, Column: 14},
				},
			},
		},
		{
			name: "simple sequence with mapping and simple single quoted value",
			src:  "- key1: value1\n  key2: value2\n- 'quotedvalue'",
			expectedTokens: []token.Token{
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 2},
					End:    token.Position{Row: 1, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 3},
					End:    token.Position{Row: 1, Column: 6},
					Origin: "key1",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 1, Column: 7},
					End:    token.Position{Row: 1, Column: 7},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 8},
					End:    token.Position{Row: 1, Column: 8},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 9},
					End:    token.Position{Row: 1, Column: 14},
					Origin: "value1",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 15},
					End:    token.Position{Row: 1, Column: 15},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 3},
					End:    token.Position{Row: 2, Column: 6},
					Origin: "key2",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 2, Column: 7},
					End:    token.Position{Row: 2, Column: 7},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 8},
					End:    token.Position{Row: 2, Column: 8},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 9},
					End:    token.Position{Row: 2, Column: 14},
					Origin: "value2",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 15},
					End:    token.Position{Row: 2, Column: 15},
					Origin: "\n",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SingleQuoteType,
					Start:  token.Position{Row: 3, Column: 3},
					End:    token.Position{Row: 3, Column: 3},
					Origin: "'",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 4},
					End:    token.Position{Row: 3, Column: 14},
					Origin: "quotedvalue",
				},
				{
					Type:   token.SingleQuoteType,
					Start:  token.Position{Row: 3, Column: 15},
					End:    token.Position{Row: 3, Column: 15},
					Origin: "'",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 3, Column: 16},
					End:   token.Position{Row: 3, Column: 16},
				},
			},
		},
		{
			name: "nested mapping with properties",
			src:  "mapping: !!map &ref\n ? innerkey\n : innervalue\naliased: *ref",
			expectedTokens: []token.Token{
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 7},
					Origin: "mapping",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 1, Column: 8},
					End:    token.Position{Row: 1, Column: 8},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 9},
					End:    token.Position{Row: 1, Column: 9},
					Origin: " ",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 1, Column: 10},
					End:    token.Position{Row: 1, Column: 10},
					Origin: "!",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 1, Column: 11},
					End:    token.Position{Row: 1, Column: 11},
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 12},
					End:    token.Position{Row: 1, Column: 14},
					Origin: "map",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 15},
					End:    token.Position{Row: 1, Column: 15},
					Origin: " ",
				},
				{
					Type:   token.AnchorType,
					Start:  token.Position{Row: 1, Column: 16},
					End:    token.Position{Row: 1, Column: 16},
					Origin: "&",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 17},
					End:    token.Position{Row: 1, Column: 19},
					Origin: "ref",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 20},
					End:    token.Position{Row: 1, Column: 20},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.MappingKeyType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: "?",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 3},
					End:    token.Position{Row: 2, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 4},
					End:    token.Position{Row: 2, Column: 11},
					Origin: "innerkey",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 12},
					End:    token.Position{Row: 2, Column: 12},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 2},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 3},
					End:    token.Position{Row: 3, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 4},
					End:    token.Position{Row: 3, Column: 13},
					Origin: "innervalue",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 14},
					End:    token.Position{Row: 3, Column: 14},
					Origin: "\n",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 7},
					Origin: "aliased",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 4, Column: 8},
					End:    token.Position{Row: 4, Column: 8},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 9},
					End:    token.Position{Row: 4, Column: 9},
					Origin: " ",
				},
				{
					Type:   token.AliasType,
					Start:  token.Position{Row: 4, Column: 10},
					End:    token.Position{Row: 4, Column: 10},
					Origin: "*",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 11},
					End:    token.Position{Row: 4, Column: 13},
					Origin: "ref",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 4, Column: 14},
					End:   token.Position{Row: 4, Column: 14},
				},
			},
		},
		{
			name: "sequence with folded and literal",
			src:  "- &lit |+ # my_comment\n  firstrow\n  secondrow\n\n- !primary >1\n\n   folded",
			expectedTokens: []token.Token{
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 2},
					End:    token.Position{Row: 1, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.AnchorType,
					Start:  token.Position{Row: 1, Column: 3},
					End:    token.Position{Row: 1, Column: 3},
					Origin: "&",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 4},
					End:    token.Position{Row: 1, Column: 6},
					Origin: "lit",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 7},
					End:    token.Position{Row: 1, Column: 7},
					Origin: " ",
				},
				{
					Type:   token.LiteralType,
					Start:  token.Position{Row: 1, Column: 8},
					End:    token.Position{Row: 1, Column: 8},
					Origin: "|",
				},
				{
					Type:   token.KeepChompingType,
					Start:  token.Position{Row: 1, Column: 9},
					End:    token.Position{Row: 1, Column: 9},
					Origin: "+",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 10},
					End:    token.Position{Row: 1, Column: 10},
					Origin: " ",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 1, Column: 11},
					End:    token.Position{Row: 1, Column: 11},
					Origin: "#",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 12},
					End:    token.Position{Row: 1, Column: 12},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 13},
					End:    token.Position{Row: 1, Column: 22},
					Origin: "my_comment",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 23},
					End:    token.Position{Row: 1, Column: 23},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 3},
					End:    token.Position{Row: 2, Column: 10},
					Origin: "firstrow",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 11},
					End:    token.Position{Row: 2, Column: 11},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 3},
					End:    token.Position{Row: 3, Column: 11},
					Origin: "secondrow",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 12},
					End:    token.Position{Row: 3, Column: 12},
					Origin: "\n",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 1},
					Origin: "\n",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 5, Column: 1},
					End:    token.Position{Row: 5, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 2},
					End:    token.Position{Row: 5, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 5, Column: 3},
					End:    token.Position{Row: 5, Column: 3},
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 5, Column: 4},
					End:    token.Position{Row: 5, Column: 10},
					Origin: "primary",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 11},
					End:    token.Position{Row: 5, Column: 11},
					Origin: " ",
				},
				{
					Type:   token.FoldedType,
					Start:  token.Position{Row: 5, Column: 12},
					End:    token.Position{Row: 5, Column: 12},
					Origin: ">",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 5, Column: 13},
					End:    token.Position{Row: 5, Column: 13},
					Origin: "1",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 5, Column: 14},
					End:    token.Position{Row: 5, Column: 14},
					Origin: "\n",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 6, Column: 1},
					End:    token.Position{Row: 6, Column: 1},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 1},
					End:    token.Position{Row: 7, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 2},
					End:    token.Position{Row: 7, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 3},
					End:    token.Position{Row: 7, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 7, Column: 4},
					End:    token.Position{Row: 7, Column: 9},
					Origin: "folded",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 7, Column: 10},
					End:   token.Position{Row: 7, Column: 10},
				},
			},
			rawModEnableIndices:  []int{14, 18, 33},
			rawModDisableIndices: []int{15, 19, 34},
		},
		{
			name: "several documents with comments",
			src: "#directives comment\n%YAML 2.2\n%TAG !yaml! tag:yaml.org,2002:\n" +
				"---\n...\n\"aaaa \\\n\"\n",
			expectedTokens: []token.Token{
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 2},
					End:    token.Position{Row: 1, Column: 11},
					Origin: "directives",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 12},
					End:    token.Position{Row: 1, Column: 12},
					Origin: " ",
				},

				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 13},
					End:    token.Position{Row: 1, Column: 19},
					Origin: "comment",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 20},
					End:    token.Position{Row: 1, Column: 20},
					Origin: "\n",
				},
				{
					Type:   token.DirectiveType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: "%",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 5},
					Origin: "YAML",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 6},
					End:    token.Position{Row: 2, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 7},
					End:    token.Position{Row: 2, Column: 9},
					Origin: "2.2",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 10},
					End:    token.Position{Row: 2, Column: 10},
					Origin: "\n",
				},
				{
					Type:   token.DirectiveType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: "%",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 4},
					Origin: "TAG",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 5},
					End:    token.Position{Row: 3, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 3, Column: 6},
					End:    token.Position{Row: 3, Column: 6},
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 7},
					End:    token.Position{Row: 3, Column: 10},
					Origin: "yaml",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 3, Column: 11},
					End:    token.Position{Row: 3, Column: 11},
					Origin: "!",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 12},
					End:    token.Position{Row: 3, Column: 12},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 13},
					End:    token.Position{Row: 3, Column: 30},
					Origin: "tag:yaml.org,2002:",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 31},
					End:    token.Position{Row: 3, Column: 31},
					Origin: "\n",
				},
				{
					Type:   token.DirectiveEndType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 3},
					Origin: "---",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 4, Column: 4},
					End:    token.Position{Row: 4, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.DocumentEndType,
					Start:  token.Position{Row: 5, Column: 1},
					End:    token.Position{Row: 5, Column: 3},
					Origin: "...",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 5, Column: 4},
					End:    token.Position{Row: 5, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 6, Column: 1},
					End:    token.Position{Row: 6, Column: 1},
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 6, Column: 2},
					End:    token.Position{Row: 6, Column: 5},
					Origin: "aaaa",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 6},
					End:    token.Position{Row: 6, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 6, Column: 7},
					End:    token.Position{Row: 6, Column: 7},
					Origin: "\\",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 6, Column: 8},
					End:    token.Position{Row: 6, Column: 8},
					Origin: "\n",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 7, Column: 1},
					End:    token.Position{Row: 7, Column: 1},
					Origin: "\"",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 7, Column: 2},
					End:    token.Position{Row: 7, Column: 2},
					Origin: "\n",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 8, Column: 1},
					End:   token.Position{Row: 8, Column: 1},
				},
			},
			rawModEnableIndices:  []int{17},
			rawModDisableIndices: []int{18},
		},
		{
			name: "null nodes",
			src: "---\nmapping:\n  \"quoted key\": #empty value\n" +
				"#empty key\n  ?\n  : value\n" +
				"#empty key and value\n  ?\n  :\n" +
				"sequence:\n  -\n  - seqvalue\n...",
			expectedTokens: []token.Token{
				{
					Type:   token.DirectiveEndType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 3},
					Origin: "---",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 4},
					End:    token.Position{Row: 1, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 7},
					Origin: "mapping",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 2, Column: 8},
					End:    token.Position{Row: 2, Column: 8},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 9},
					End:    token.Position{Row: 2, Column: 9},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 3, Column: 3},
					End:    token.Position{Row: 3, Column: 3},
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 4},
					End:    token.Position{Row: 3, Column: 9},
					Origin: "quoted",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 10},
					End:    token.Position{Row: 3, Column: 10},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 11},
					End:    token.Position{Row: 3, Column: 13},
					Origin: "key",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 3, Column: 14},
					End:    token.Position{Row: 3, Column: 14},
					Origin: "\"",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 3, Column: 15},
					End:    token.Position{Row: 3, Column: 15},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 16},
					End:    token.Position{Row: 3, Column: 16},
					Origin: " ",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 3, Column: 17},
					End:    token.Position{Row: 3, Column: 17},
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 18},
					End:    token.Position{Row: 3, Column: 22},
					Origin: "empty",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 23},
					End:    token.Position{Row: 3, Column: 23},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 24},
					End:    token.Position{Row: 3, Column: 28},
					Origin: "value",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 29},
					End:    token.Position{Row: 3, Column: 29},
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 1},
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 2},
					End:    token.Position{Row: 4, Column: 6},
					Origin: "empty",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 7},
					End:    token.Position{Row: 4, Column: 7},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 8},
					End:    token.Position{Row: 4, Column: 10},
					Origin: "key",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 4, Column: 11},
					End:    token.Position{Row: 4, Column: 11},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 1},
					End:    token.Position{Row: 5, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 2},
					End:    token.Position{Row: 5, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.MappingKeyType,
					Start:  token.Position{Row: 5, Column: 3},
					End:    token.Position{Row: 5, Column: 3},
					Origin: "?",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 5, Column: 4},
					End:    token.Position{Row: 5, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 1},
					End:    token.Position{Row: 6, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 2},
					End:    token.Position{Row: 6, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 6, Column: 3},
					End:    token.Position{Row: 6, Column: 3},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 4},
					End:    token.Position{Row: 6, Column: 4},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 6, Column: 5},
					End:    token.Position{Row: 6, Column: 9},
					Origin: "value",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 6, Column: 10},
					End:    token.Position{Row: 6, Column: 10},
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 7, Column: 1},
					End:    token.Position{Row: 7, Column: 1},
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 7, Column: 2},
					End:    token.Position{Row: 7, Column: 6},
					Origin: "empty",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 7},
					End:    token.Position{Row: 7, Column: 7},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 7, Column: 8},
					End:    token.Position{Row: 7, Column: 10},
					Origin: "key",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 11},
					End:    token.Position{Row: 7, Column: 11},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 7, Column: 12},
					End:    token.Position{Row: 7, Column: 14},
					Origin: "and",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 15},
					End:    token.Position{Row: 7, Column: 15},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 7, Column: 16},
					End:    token.Position{Row: 7, Column: 20},
					Origin: "value",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 7, Column: 21},
					End:    token.Position{Row: 7, Column: 21},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 1},
					End:    token.Position{Row: 8, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 2},
					End:    token.Position{Row: 8, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.MappingKeyType,
					Start:  token.Position{Row: 8, Column: 3},
					End:    token.Position{Row: 8, Column: 3},
					Origin: "?",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 8, Column: 4},
					End:    token.Position{Row: 8, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 9, Column: 1},
					End:    token.Position{Row: 9, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 9, Column: 2},
					End:    token.Position{Row: 9, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 9, Column: 3},
					End:    token.Position{Row: 9, Column: 3},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 9, Column: 4},
					End:    token.Position{Row: 9, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 10, Column: 1},
					End:    token.Position{Row: 10, Column: 8},
					Origin: "sequence",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 10, Column: 9},
					End:    token.Position{Row: 10, Column: 9},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 10, Column: 10},
					End:    token.Position{Row: 10, Column: 10},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 11, Column: 1},
					End:    token.Position{Row: 11, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 11, Column: 2},
					End:    token.Position{Row: 11, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 11, Column: 3},
					End:    token.Position{Row: 11, Column: 3},
					Origin: "-",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 11, Column: 4},
					End:    token.Position{Row: 11, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 12, Column: 1},
					End:    token.Position{Row: 12, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 12, Column: 2},
					End:    token.Position{Row: 12, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 12, Column: 3},
					End:    token.Position{Row: 12, Column: 3},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 12, Column: 4},
					End:    token.Position{Row: 12, Column: 4},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 12, Column: 5},
					End:    token.Position{Row: 12, Column: 12},
					Origin: "seqvalue",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 12, Column: 13},
					End:    token.Position{Row: 12, Column: 13},
					Origin: "\n",
				},
				{
					Type:   token.DocumentEndType,
					Start:  token.Position{Row: 13, Column: 1},
					End:    token.Position{Row: 13, Column: 3},
					Origin: "...",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 13, Column: 4},
					End:   token.Position{Row: 13, Column: 4},
				},
			},
		},
		{
			name: "flow syntax sequence",
			src:  "[\tplain\t,\"\\\"multi\\\"\n\n\n  line\",flow: pair,\t? ]",
			expectedTokens: []token.Token{
				{
					Type:   token.SequenceStartType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "[",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 1, Column: 2},
					End:    token.Position{Row: 1, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 3},
					End:    token.Position{Row: 1, Column: 7},
					Origin: "plain",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 1, Column: 8},
					End:    token.Position{Row: 1, Column: 8},
					Origin: "\t",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 1, Column: 9},
					End:    token.Position{Row: 1, Column: 9},
					Origin: ",",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 1, Column: 10},
					End:    token.Position{Row: 1, Column: 10},
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 11},
					End:    token.Position{Row: 1, Column: 19},
					Origin: "\\\"multi\\\"",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 20},
					End:    token.Position{Row: 1, Column: 20},
					Origin: "\n",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: "\n",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 2},
					End:    token.Position{Row: 4, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 3},
					End:    token.Position{Row: 4, Column: 6},
					Origin: "line",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 4, Column: 7},
					End:    token.Position{Row: 4, Column: 7},
					Origin: "\"",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 4, Column: 8},
					End:    token.Position{Row: 4, Column: 8},
					Origin: ",",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 9},
					End:    token.Position{Row: 4, Column: 12},
					Origin: "flow",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 4, Column: 13},
					End:    token.Position{Row: 4, Column: 13},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 14},
					End:    token.Position{Row: 4, Column: 14},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 15},
					End:    token.Position{Row: 4, Column: 18},
					Origin: "pair",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 4, Column: 19},
					End:    token.Position{Row: 4, Column: 19},
					Origin: ",",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 4, Column: 20},
					End:    token.Position{Row: 4, Column: 20},
					Origin: "\t",
				},
				{
					Type:   token.MappingKeyType,
					Start:  token.Position{Row: 4, Column: 21},
					End:    token.Position{Row: 4, Column: 21},
					Origin: "?",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 22},
					End:    token.Position{Row: 4, Column: 22},
					Origin: " ",
				},
				{
					Type:   token.SequenceEndType,
					Start:  token.Position{Row: 4, Column: 23},
					End:    token.Position{Row: 4, Column: 23},
					Origin: "]",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 4, Column: 24},
					End:   token.Position{Row: 4, Column: 24},
				},
			},
		},
		{
			name: "flow syntax mapping",
			src: "{ unquoted: " +
				"'''single quoted''\n \n \n  multiline \"ათჯერ გაზომე და ერთხელ გაჭერი\"'" +
				",? \"explicit key\":adjacent,novalue,:,?\t}",
			expectedTokens: []token.Token{
				{
					Type:   token.MappingStartType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "{",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 2},
					End:    token.Position{Row: 1, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 3},
					End:    token.Position{Row: 1, Column: 10},
					Origin: "unquoted",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 1, Column: 11},
					End:    token.Position{Row: 1, Column: 11},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 12},
					End:    token.Position{Row: 1, Column: 12},
					Origin: " ",
				},
				{
					Type:   token.SingleQuoteType,
					Start:  token.Position{Row: 1, Column: 13},
					End:    token.Position{Row: 1, Column: 13},
					Origin: "'",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 14},
					End:    token.Position{Row: 1, Column: 21},
					Origin: "''single",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 22},
					End:    token.Position{Row: 1, Column: 22},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 23},
					End:    token.Position{Row: 1, Column: 30},
					Origin: "quoted''",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 31},
					End:    token.Position{Row: 1, Column: 31},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 2},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 2},
					End:    token.Position{Row: 4, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 3},
					End:    token.Position{Row: 4, Column: 11},
					Origin: "multiline",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 12},
					End:    token.Position{Row: 4, Column: 12},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 13},
					End:    token.Position{Row: 4, Column: 18},
					Origin: "\"ათჯერ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 19},
					End:    token.Position{Row: 4, Column: 19},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 20},
					End:    token.Position{Row: 4, Column: 25},
					Origin: "გაზომე",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 26},
					End:    token.Position{Row: 4, Column: 26},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 27},
					End:    token.Position{Row: 4, Column: 28},
					Origin: "და",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 29},
					End:    token.Position{Row: 4, Column: 29},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 30},
					End:    token.Position{Row: 4, Column: 35},
					Origin: "ერთხელ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 36},
					End:    token.Position{Row: 4, Column: 36},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 37},
					End:    token.Position{Row: 4, Column: 43},
					Origin: "გაჭერი\"",
				},
				{
					Type:   token.SingleQuoteType,
					Start:  token.Position{Row: 4, Column: 44},
					End:    token.Position{Row: 4, Column: 44},
					Origin: "'",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 4, Column: 45},
					End:    token.Position{Row: 4, Column: 45},
					Origin: ",",
				},
				{
					Type:   token.MappingKeyType,
					Start:  token.Position{Row: 4, Column: 46},
					End:    token.Position{Row: 4, Column: 46},
					Origin: "?",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 47},
					End:    token.Position{Row: 4, Column: 47},
					Origin: " ",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 4, Column: 48},
					End:    token.Position{Row: 4, Column: 48},
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 49},
					End:    token.Position{Row: 4, Column: 56},
					Origin: "explicit",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 57},
					End:    token.Position{Row: 4, Column: 57},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 58},
					End:    token.Position{Row: 4, Column: 60},
					Origin: "key",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 4, Column: 61},
					End:    token.Position{Row: 4, Column: 61},
					Origin: "\"",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 4, Column: 62},
					End:    token.Position{Row: 4, Column: 62},
					Origin: ":",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 63},
					End:    token.Position{Row: 4, Column: 70},
					Origin: "adjacent",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 4, Column: 71},
					End:    token.Position{Row: 4, Column: 71},
					Origin: ",",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 72},
					End:    token.Position{Row: 4, Column: 78},
					Origin: "novalue",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 4, Column: 79},
					End:    token.Position{Row: 4, Column: 79},
					Origin: ",",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 4, Column: 80},
					End:    token.Position{Row: 4, Column: 80},
					Origin: ":",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 4, Column: 81},
					End:    token.Position{Row: 4, Column: 81},
					Origin: ",",
				},
				{
					Type:   token.MappingKeyType,
					Start:  token.Position{Row: 4, Column: 82},
					End:    token.Position{Row: 4, Column: 82},
					Origin: "?",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 4, Column: 83},
					End:    token.Position{Row: 4, Column: 83},
					Origin: "\t",
				},
				{
					Type:   token.MappingEndType,
					Start:  token.Position{Row: 4, Column: 84},
					End:    token.Position{Row: 4, Column: 84},
					Origin: "}",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 4, Column: 85},
					End:   token.Position{Row: 4, Column: 85},
				},
			},
		},
		{
			name: "mix",
			src: "%RESERVED PARAMETER\n" +
				"#directivecomment\n" +
				"%TAG ! !local-tag-\n" +
				"---\n" +
				"-  -\t!<!baz> entity\n" +
				"   - plain\n\n" +
				"     multi\n" +
				"     line\n" +
				"- >-\n" +
				"  \n" +
				"  \tspaced\n" +
				"   text\n\n" +
				"#trail\n" +
				"#comments",
			expectedTokens: []token.Token{
				{
					Type:   token.DirectiveType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "%",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 2},
					End:    token.Position{Row: 1, Column: 9},
					Origin: "RESERVED",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 1, Column: 10},
					End:    token.Position{Row: 1, Column: 10},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 1, Column: 11},
					End:    token.Position{Row: 1, Column: 19},
					Origin: "PARAMETER",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 20},
					End:    token.Position{Row: 1, Column: 20},
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 17},
					Origin: "directivecomment",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 18},
					End:    token.Position{Row: 2, Column: 18},
					Origin: "\n",
				},
				{
					Type:   token.DirectiveType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: "%",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 4},
					Origin: "TAG",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 5},
					End:    token.Position{Row: 3, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 3, Column: 6},
					End:    token.Position{Row: 3, Column: 6},
					Origin: "!",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 7},
					End:    token.Position{Row: 3, Column: 7},
					Origin: " ",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 3, Column: 8},
					End:    token.Position{Row: 3, Column: 8},
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 9},
					End:    token.Position{Row: 3, Column: 18},
					Origin: "local-tag-",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 19},
					End:    token.Position{Row: 3, Column: 19},
					Origin: "\n",
				},
				{
					Type:   token.DirectiveEndType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 3},
					Origin: "---",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 4, Column: 4},
					End:    token.Position{Row: 4, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 5, Column: 1},
					End:    token.Position{Row: 5, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 2},
					End:    token.Position{Row: 5, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 3},
					End:    token.Position{Row: 5, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 5, Column: 4},
					End:    token.Position{Row: 5, Column: 4},
					Origin: "-",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 5, Column: 5},
					End:    token.Position{Row: 5, Column: 5},
					Origin: "\t",
				},
				{
					Type:   token.TagType,
					Start:  token.Position{Row: 5, Column: 6},
					End:    token.Position{Row: 5, Column: 6},
					Origin: "!",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 5, Column: 7},
					End:    token.Position{Row: 5, Column: 12},
					Origin: "<!baz>",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 13},
					End:    token.Position{Row: 5, Column: 13},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 5, Column: 14},
					End:    token.Position{Row: 5, Column: 19},
					Origin: "entity",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 5, Column: 20},
					End:    token.Position{Row: 5, Column: 20},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 1},
					End:    token.Position{Row: 6, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 2},
					End:    token.Position{Row: 6, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 3},
					End:    token.Position{Row: 6, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 6, Column: 4},
					End:    token.Position{Row: 6, Column: 4},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 5},
					End:    token.Position{Row: 6, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 6, Column: 6},
					End:    token.Position{Row: 6, Column: 10},
					Origin: "plain",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 6, Column: 11},
					End:    token.Position{Row: 6, Column: 11},
					Origin: "\n",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 7, Column: 1},
					End:    token.Position{Row: 7, Column: 1},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 1},
					End:    token.Position{Row: 8, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 2},
					End:    token.Position{Row: 8, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 3},
					End:    token.Position{Row: 8, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 4},
					End:    token.Position{Row: 8, Column: 4},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 5},
					End:    token.Position{Row: 8, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 8, Column: 6},
					End:    token.Position{Row: 8, Column: 10},
					Origin: "multi",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 8, Column: 11},
					End:    token.Position{Row: 8, Column: 11},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 9, Column: 1},
					End:    token.Position{Row: 9, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 9, Column: 2},
					End:    token.Position{Row: 9, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 9, Column: 3},
					End:    token.Position{Row: 9, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 9, Column: 4},
					End:    token.Position{Row: 9, Column: 4},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 9, Column: 5},
					End:    token.Position{Row: 9, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 9, Column: 6},
					End:    token.Position{Row: 9, Column: 9},
					Origin: "line",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 9, Column: 10},
					End:    token.Position{Row: 9, Column: 10},
					Origin: "\n",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 10, Column: 1},
					End:    token.Position{Row: 10, Column: 1},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 10, Column: 2},
					End:    token.Position{Row: 10, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.FoldedType,
					Start:  token.Position{Row: 10, Column: 3},
					End:    token.Position{Row: 10, Column: 3},
					Origin: ">",
				},
				{
					Type:   token.StripChompingType,
					Start:  token.Position{Row: 10, Column: 4},
					End:    token.Position{Row: 10, Column: 4},
					Origin: "-",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 10, Column: 5},
					End:    token.Position{Row: 10, Column: 5},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 11, Column: 1},
					End:    token.Position{Row: 11, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 11, Column: 2},
					End:    token.Position{Row: 11, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 11, Column: 3},
					End:    token.Position{Row: 11, Column: 3},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 12, Column: 1},
					End:    token.Position{Row: 12, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 12, Column: 2},
					End:    token.Position{Row: 12, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 12, Column: 3},
					End:    token.Position{Row: 12, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 12, Column: 4},
					End:    token.Position{Row: 12, Column: 9},
					Origin: "spaced",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 12, Column: 10},
					End:    token.Position{Row: 12, Column: 10},
					Origin: "\n",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 13, Column: 1},
					End:    token.Position{Row: 13, Column: 1},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 13, Column: 2},
					End:    token.Position{Row: 13, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 13, Column: 3},
					End:    token.Position{Row: 13, Column: 3},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 13, Column: 4},
					End:    token.Position{Row: 13, Column: 7},
					Origin: "text",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 13, Column: 8},
					End:    token.Position{Row: 13, Column: 8},
					Origin: "\n",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 14, Column: 1},
					End:    token.Position{Row: 14, Column: 1},
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 15, Column: 1},
					End:    token.Position{Row: 15, Column: 1},
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 15, Column: 2},
					End:    token.Position{Row: 15, Column: 6},
					Origin: "trail",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 15, Column: 7},
					End:    token.Position{Row: 15, Column: 7},
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 16, Column: 1},
					End:    token.Position{Row: 16, Column: 1},
					Origin: "#",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 16, Column: 2},
					End:    token.Position{Row: 16, Column: 9},
					Origin: "comments",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 16, Column: 10},
					End:   token.Position{Row: 16, Column: 10},
				},
			},
			rawModEnableIndices:  []int{1, 2, 9, 14, 24, 60, 65},
			rawModDisableIndices: []int{2, 4, 10, 15, 25, 62, 67},
		},
		{
			name: "mix2",
			src: "...\r\n\uFEFF # stream comment\r---\n# document\n...\n&anchor\n...\n" +
				"[*anchor : , plain: [],: emptyk,\"adjacent\":value]",
			expectedTokens: []token.Token{
				{
					Type:   token.DocumentEndType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 3},
					Origin: "...",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 4},
					End:    token.Position{Row: 1, Column: 5},
					Origin: "\r\n",
				},
				{
					Type:   token.BOMType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: "\uFEFF",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 2, Column: 3},
					End:    token.Position{Row: 2, Column: 3},
					Origin: "#",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 4},
					End:    token.Position{Row: 2, Column: 4},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 5},
					End:    token.Position{Row: 2, Column: 10},
					Origin: "stream",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 2, Column: 11},
					End:    token.Position{Row: 2, Column: 11},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 12},
					End:    token.Position{Row: 2, Column: 18},
					Origin: "comment",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 19},
					End:    token.Position{Row: 2, Column: 19},
					Origin: "\r",
				},
				{
					Type:   token.DirectiveEndType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 3},
					Origin: "---",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 4},
					End:    token.Position{Row: 3, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.CommentType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 1},
					Origin: "#",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 2},
					End:    token.Position{Row: 4, Column: 2},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 3},
					End:    token.Position{Row: 4, Column: 10},
					Origin: "document",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 4, Column: 11},
					End:    token.Position{Row: 4, Column: 11},
					Origin: "\n",
				},
				{
					Type:   token.DocumentEndType,
					Start:  token.Position{Row: 5, Column: 1},
					End:    token.Position{Row: 5, Column: 3},
					Origin: "...",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 5, Column: 4},
					End:    token.Position{Row: 5, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.AnchorType,
					Start:  token.Position{Row: 6, Column: 1},
					End:    token.Position{Row: 6, Column: 1},
					Origin: "&",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 6, Column: 2},
					End:    token.Position{Row: 6, Column: 7},
					Origin: "anchor",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 6, Column: 8},
					End:    token.Position{Row: 6, Column: 8},
					Origin: "\n",
				},
				{
					Type:   token.DocumentEndType,
					Start:  token.Position{Row: 7, Column: 1},
					End:    token.Position{Row: 7, Column: 3},
					Origin: "...",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 7, Column: 4},
					End:    token.Position{Row: 7, Column: 4},
					Origin: "\n",
				},
				{
					Type:   token.SequenceStartType,
					Start:  token.Position{Row: 8, Column: 1},
					End:    token.Position{Row: 8, Column: 1},
					Origin: "[",
				},
				{
					Type:   token.AliasType,
					Start:  token.Position{Row: 8, Column: 2},
					End:    token.Position{Row: 8, Column: 2},
					Origin: "*",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 8, Column: 3},
					End:    token.Position{Row: 8, Column: 8},
					Origin: "anchor",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 9},
					End:    token.Position{Row: 8, Column: 9},
					Origin: " ",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 8, Column: 10},
					End:    token.Position{Row: 8, Column: 10},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 11},
					End:    token.Position{Row: 8, Column: 11},
					Origin: " ",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 8, Column: 12},
					End:    token.Position{Row: 8, Column: 12},
					Origin: ",",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 13},
					End:    token.Position{Row: 8, Column: 13},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 8, Column: 14},
					End:    token.Position{Row: 8, Column: 18},
					Origin: "plain",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 8, Column: 19},
					End:    token.Position{Row: 8, Column: 19},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 20},
					End:    token.Position{Row: 8, Column: 20},
					Origin: " ",
				},
				{
					Type:   token.SequenceStartType,
					Start:  token.Position{Row: 8, Column: 21},
					End:    token.Position{Row: 8, Column: 21},
					Origin: "[",
				},
				{
					Type:   token.SequenceEndType,
					Start:  token.Position{Row: 8, Column: 22},
					End:    token.Position{Row: 8, Column: 22},
					Origin: "]",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 8, Column: 23},
					End:    token.Position{Row: 8, Column: 23},
					Origin: ",",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 8, Column: 24},
					End:    token.Position{Row: 8, Column: 24},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 25},
					End:    token.Position{Row: 8, Column: 25},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 8, Column: 26},
					End:    token.Position{Row: 8, Column: 31},
					Origin: "emptyk",
				},
				{
					Type:   token.CollectEntryType,
					Start:  token.Position{Row: 8, Column: 32},
					End:    token.Position{Row: 8, Column: 32},
					Origin: ",",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 8, Column: 33},
					End:    token.Position{Row: 8, Column: 33},
					Origin: "\"",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 8, Column: 34},
					End:    token.Position{Row: 8, Column: 41},
					Origin: "adjacent",
				},
				{
					Type:   token.DoubleQuoteType,
					Start:  token.Position{Row: 8, Column: 42},
					End:    token.Position{Row: 8, Column: 42},
					Origin: "\"",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 8, Column: 43},
					End:    token.Position{Row: 8, Column: 43},
					Origin: ":",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 8, Column: 44},
					End:    token.Position{Row: 8, Column: 48},
					Origin: "value",
				},
				{
					Type:   token.SequenceEndType,
					Start:  token.Position{Row: 8, Column: 49},
					End:    token.Position{Row: 8, Column: 49},
					Origin: "]",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 8, Column: 50},
					End:   token.Position{Row: 8, Column: 50},
				},
			},
		},
		{
			name: "real example",
			src: `
				network:
				  version: 2
				  renderer: networkd
				  ethernets:
				    enp3s0:
					  addresses:
					    - 10.10.10.2/24
				`,
			expectedTokens: []token.Token{
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 1, Column: 1},
					End:    token.Position{Row: 1, Column: 1},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 2, Column: 1},
					End:    token.Position{Row: 2, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 2, Column: 2},
					End:    token.Position{Row: 2, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 2, Column: 3},
					End:    token.Position{Row: 2, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 2, Column: 4},
					End:    token.Position{Row: 2, Column: 4},
					Origin: "\t",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 2, Column: 5},
					End:    token.Position{Row: 2, Column: 11},
					Origin: "network",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 2, Column: 12},
					End:    token.Position{Row: 2, Column: 12},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 2, Column: 13},
					End:    token.Position{Row: 2, Column: 13},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 3, Column: 1},
					End:    token.Position{Row: 3, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 3, Column: 2},
					End:    token.Position{Row: 3, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 3, Column: 3},
					End:    token.Position{Row: 3, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 3, Column: 4},
					End:    token.Position{Row: 3, Column: 4},
					Origin: "\t",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 5},
					End:    token.Position{Row: 3, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 6},
					End:    token.Position{Row: 3, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 7},
					End:    token.Position{Row: 3, Column: 13},
					Origin: "version",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 3, Column: 14},
					End:    token.Position{Row: 3, Column: 14},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 3, Column: 15},
					End:    token.Position{Row: 3, Column: 15},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 3, Column: 16},
					End:    token.Position{Row: 3, Column: 16},
					Origin: "2",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 3, Column: 17},
					End:    token.Position{Row: 3, Column: 17},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 4, Column: 1},
					End:    token.Position{Row: 4, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 4, Column: 2},
					End:    token.Position{Row: 4, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 4, Column: 3},
					End:    token.Position{Row: 4, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 4, Column: 4},
					End:    token.Position{Row: 4, Column: 4},
					Origin: "\t",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 5},
					End:    token.Position{Row: 4, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 6},
					End:    token.Position{Row: 4, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 7},
					End:    token.Position{Row: 4, Column: 14},
					Origin: "renderer",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 4, Column: 15},
					End:    token.Position{Row: 4, Column: 15},
					Origin: ":",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 4, Column: 16},
					End:    token.Position{Row: 4, Column: 16},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 4, Column: 17},
					End:    token.Position{Row: 4, Column: 24},
					Origin: "networkd",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 4, Column: 25},
					End:    token.Position{Row: 4, Column: 25},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 5, Column: 1},
					End:    token.Position{Row: 5, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 5, Column: 2},
					End:    token.Position{Row: 5, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 5, Column: 3},
					End:    token.Position{Row: 5, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 5, Column: 4},
					End:    token.Position{Row: 5, Column: 4},
					Origin: "\t",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 5},
					End:    token.Position{Row: 5, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 5, Column: 6},
					End:    token.Position{Row: 5, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 5, Column: 7},
					End:    token.Position{Row: 5, Column: 15},
					Origin: "ethernets",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 5, Column: 16},
					End:    token.Position{Row: 5, Column: 16},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 5, Column: 17},
					End:    token.Position{Row: 5, Column: 17},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 6, Column: 1},
					End:    token.Position{Row: 6, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 6, Column: 2},
					End:    token.Position{Row: 6, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 6, Column: 3},
					End:    token.Position{Row: 6, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 6, Column: 4},
					End:    token.Position{Row: 6, Column: 4},
					Origin: "\t",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 5},
					End:    token.Position{Row: 6, Column: 5},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 6},
					End:    token.Position{Row: 6, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 7},
					End:    token.Position{Row: 6, Column: 7},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 6, Column: 8},
					End:    token.Position{Row: 6, Column: 8},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 6, Column: 9},
					End:    token.Position{Row: 6, Column: 14},
					Origin: "enp3s0",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 6, Column: 15},
					End:    token.Position{Row: 6, Column: 15},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 6, Column: 16},
					End:    token.Position{Row: 6, Column: 16},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 7, Column: 1},
					End:    token.Position{Row: 7, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 7, Column: 2},
					End:    token.Position{Row: 7, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 7, Column: 3},
					End:    token.Position{Row: 7, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 7, Column: 4},
					End:    token.Position{Row: 7, Column: 4},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 7, Column: 5},
					End:    token.Position{Row: 7, Column: 5},
					Origin: "\t",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 6},
					End:    token.Position{Row: 7, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 7, Column: 7},
					End:    token.Position{Row: 7, Column: 7},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 7, Column: 8},
					End:    token.Position{Row: 7, Column: 16},
					Origin: "addresses",
				},
				{
					Type:   token.MappingValueType,
					Start:  token.Position{Row: 7, Column: 17},
					End:    token.Position{Row: 7, Column: 17},
					Origin: ":",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 7, Column: 18},
					End:    token.Position{Row: 7, Column: 18},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 8, Column: 1},
					End:    token.Position{Row: 8, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 8, Column: 2},
					End:    token.Position{Row: 8, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 8, Column: 3},
					End:    token.Position{Row: 8, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 8, Column: 4},
					End:    token.Position{Row: 8, Column: 4},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 8, Column: 5},
					End:    token.Position{Row: 8, Column: 5},
					Origin: "\t",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 6},
					End:    token.Position{Row: 8, Column: 6},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 7},
					End:    token.Position{Row: 8, Column: 7},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 8},
					End:    token.Position{Row: 8, Column: 8},
					Origin: " ",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 9},
					End:    token.Position{Row: 8, Column: 9},
					Origin: " ",
				},
				{
					Type:   token.SequenceEntryType,
					Start:  token.Position{Row: 8, Column: 10},
					End:    token.Position{Row: 8, Column: 10},
					Origin: "-",
				},
				{
					Type:   token.SpaceType,
					Start:  token.Position{Row: 8, Column: 11},
					End:    token.Position{Row: 8, Column: 11},
					Origin: " ",
				},
				{
					Type:   token.StringType,
					Start:  token.Position{Row: 8, Column: 12},
					End:    token.Position{Row: 8, Column: 24},
					Origin: "10.10.10.2/24",
				},
				{
					Type:   token.LineBreakType,
					Start:  token.Position{Row: 8, Column: 25},
					End:    token.Position{Row: 8, Column: 25},
					Origin: "\n",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 9, Column: 1},
					End:    token.Position{Row: 9, Column: 1},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 9, Column: 2},
					End:    token.Position{Row: 9, Column: 2},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 9, Column: 3},
					End:    token.Position{Row: 9, Column: 3},
					Origin: "\t",
				},
				{
					Type:   token.TabType,
					Start:  token.Position{Row: 9, Column: 4},
					End:    token.Position{Row: 9, Column: 4},
					Origin: "\t",
				},
				{
					Type:  token.EOFType,
					Start: token.Position{Row: 9, Column: 5},
					End:   token.Position{Row: 9, Column: 5},
				},
			},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			tokenizer := lexer.NewTokenizer(tc.src)

			var rawModEnableIndex, rawModDisableIndex int

			var (
				tokens       []token.Token
				currentToken token.Token
			)
			for i := 0; currentToken.Type != token.EOFType; i++ {
				if rawModDisableIndex != len(tc.rawModDisableIndices) && i == tc.rawModDisableIndices[rawModDisableIndex] {
					tokenizer.UnsetRawMode()
					rawModDisableIndex++
				}

				if rawModEnableIndex != len(tc.rawModEnableIndices) && i == tc.rawModEnableIndices[rawModEnableIndex] {
					tokenizer.SetRawMode()
					rawModEnableIndex++
				}

				currentToken = tokenizer.Next()
				tokens = append(tokens, currentToken)
			}
			compareTokens(t, tc.expectedTokens, tokens)
		})
	}
}

func compareTokens(t *testing.T, expectedTokens, actualTokens []token.Token) {
	t.Helper()

	if len(expectedTokens) != len(actualTokens) {
		t.Errorf(
			"expected and actual tokens have different length: %d and %d respectively\nexpected: %v\ngot: %v",
			len(expectedTokens),
			len(actualTokens),
			expectedTokens,
			actualTokens,
		)
	}

	n := len(expectedTokens)
	if len(actualTokens) < n {
		n = len(actualTokens)
	}

	for i := 0; i < n; i++ {
		if !areTokensEqual(expectedTokens[i], actualTokens[i]) {
			t.Fatalf(
				"tokens at index %d differ:\nexpected: %v\nactual: %v",
				i,
				expectedTokens[i],
				actualTokens[i],
			)
		}
	}
}

func areTokensEqual(a, b token.Token) bool {
	switch {
	case a.Type != b.Type:
		return false
	case a.Origin != b.Origin:
		return false
	case a.Start != b.Start:
		return false
	case a.End != b.End:
		return false
	}
	return true
}
