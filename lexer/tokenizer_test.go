package lexer_test

import (
	"github.com/KSpaceer/yayamls/lexer"
	"github.com/KSpaceer/yayamls/token"
	"testing"
)

func TestTokenizer(t *testing.T) {
	type tcase struct {
		name           string
		src            string
		expectedTokens []token.Token
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
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			ts := lexer.NewTokenStream(tc.src)

			var (
				tokens       []token.Token
				currentToken token.Token
			)
			for currentToken.Type != token.EOFType {
				currentToken = ts.Next()
				tokens = append(tokens, currentToken)
			}
			compareTokens(t, tc.expectedTokens, tokens)
		})
	}
}

func compareTokens(t *testing.T, expectedTokens, actualTokens []token.Token) {
	t.Helper()

	if len(expectedTokens) != len(actualTokens) {
		t.Fatalf(
			"expected and actual tokens have different length: %d and %d respectively\nexpected: %v\ngot: %v",
			len(expectedTokens),
			len(actualTokens),
			expectedTokens,
			actualTokens,
		)
	}

	for i := range expectedTokens {
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
