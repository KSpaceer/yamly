package parser_test

import (
	"github.com/KSpaceer/fastyaml/parser"
	"github.com/KSpaceer/fastyaml/token"
	"testing"
)

func TestTokenAccessor(t *testing.T) {
	type tcase struct {
		Name           string
		LoadedTokens   []token.Token
		ExpectedTokens []token.Token
		Checkpoints    []int
		Commits        []int
		Rollbacks      []int
	}

	tcases := []tcase{
		{
			Name: "Simple",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			Checkpoints: nil,
			Commits:     nil,
			Rollbacks:   nil,
		},
		{
			Name: "Single checkpoint only",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			Checkpoints: []int{0},
			Commits:     nil,
			Rollbacks:   nil,
		},
		{
			Name: "Commited checkpoint",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			Checkpoints: []int{0},
			Commits:     []int{1},
			Rollbacks:   nil,
		},
		{
			Name: "Checkpoint with rollback",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			Checkpoints: []int{0},
			Commits:     []int{2},
			Rollbacks:   []int{1},
		},
		{
			Name: "Two checkpoints, one rollback",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
			},
			Checkpoints: []int{0, 1},
			Commits:     []int{},
			Rollbacks:   []int{2},
		},
		{
			Name: "Nested commits",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
				{
					Type: token.TagType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
				{
					Type: token.TagType,
				},
			},
			Checkpoints: []int{0, 2},
			Commits:     []int{1, 3},
			Rollbacks:   nil,
		},
		{
			Name: "Nested rollbacks",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
				{
					Type: token.TagType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
				{
					Type: token.TagType,
				},
			},
			Checkpoints: []int{0, 1},
			Commits:     nil,
			Rollbacks:   []int{2, 3},
		},
		{
			Name: "Checkpoint, rollback, checkpoint, commit, checkpoint, checkpoint, commit, rollback",
			LoadedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
				{
					Type: token.TagType,
				},
				{
					Type: token.DirectiveType,
				},
				{
					Type: token.LineBreakType,
				},
				{
					Type: token.AliasType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.StringType,
				},
				{
					Type: token.SpaceType,
				},
				{
					Type: token.AnchorType,
				},
				{
					Type: token.TagType,
				},
				{
					Type: token.DirectiveType,
				},
				{
					Type: token.LineBreakType,
				},
				{
					Type: token.AliasType,
				},
				{
					Type: token.DirectiveType,
				},
				{
					Type: token.LineBreakType,
				},
				{
					Type: token.AliasType,
				},
			},
			Checkpoints: []int{0, 3, 6, 7},
			Commits:     []int{5, 8},
			Rollbacks:   []int{2, 9},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.Name, func(t *testing.T) {
			stream := &testTokenStream{
				tokens: tc.LoadedTokens,
				index:  0,
			}
			result := make([]token.Token, 0, len(tc.ExpectedTokens))
			tokenAccessor := parser.NewTokenAccessor(stream)

			var (
				checkpointIdx int
				commitIdx     int
				rollbackIdx   int
			)

			for i := 0; i < len(tc.ExpectedTokens); i++ {
				if commitIdx < len(tc.Commits) && i == tc.Commits[commitIdx] {
					tokenAccessor.Commit()
					commitIdx++
				}
				if rollbackIdx < len(tc.Rollbacks) && i == tc.Rollbacks[rollbackIdx] {
					tokenAccessor.Rollback()
					rollbackIdx++
				}
				if checkpointIdx < len(tc.Checkpoints) && i == tc.Checkpoints[checkpointIdx] {
					tokenAccessor.SetCheckpoint()
					checkpointIdx++
				}
				result = append(result, tokenAccessor.Next())
			}

			if len(result) != len(tc.ExpectedTokens) {
				t.Fatalf("expected result tokens to have length [%d] but got [%d]",
					len(tc.ExpectedTokens), len(result))
			}

			for i := range tc.ExpectedTokens {
				if tc.ExpectedTokens[i].Type != result[i].Type {
					t.Fatalf("expected token [%v] but got [%v] at position %d",
						tc.ExpectedTokens[i], result[i], i)
				}
			}
		})
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
