package cpaccessor_test

import (
	"github.com/KSpaceer/yayamls/cpaccessor"
	"github.com/KSpaceer/yayamls/token"
	"testing"
)

func TestTokenAccessor(t *testing.T) {
	type rollbackInfo struct {
		idx int
		tok token.Token
	}
	type tcase struct {
		Name           string
		LoadedTokens   []token.Token
		ExpectedTokens []token.Token
		Checkpoints    []int
		Commits        []int
		Rollbacks      []rollbackInfo
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
			Rollbacks: []rollbackInfo{
				{
					idx: 1,
					tok: token.Token{},
				},
			},
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
			Rollbacks: []rollbackInfo{
				{
					idx: 2,
					tok: token.Token{Type: token.StringType},
				},
			},
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
			Rollbacks: []rollbackInfo{
				{
					idx: 2,
					tok: token.Token{Type: token.StringType},
				},
				{
					idx: 3,
					tok: token.Token{},
				},
			},
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
			Rollbacks: []rollbackInfo{
				{
					idx: 2,
					tok: token.Token{},
				},
				{
					idx: 9,
					tok: token.Token{Type: token.TagType},
				},
			},
		},
		{
			Name: "double nested rollback",
			LoadedTokens: []token.Token{
				{
					Type: token.SpaceType,
				},
				{
					Type: token.StringType,
				},
				{
					Type: token.SequenceEntryType,
				},
				{
					Type: token.CommentType,
				},
			},
			ExpectedTokens: []token.Token{
				{
					Type: token.SpaceType,
				},
				{
					Type: token.StringType,
				},
				{
					Type: token.SequenceEntryType,
				},
				{
					Type: token.StringType,
				},
				{
					Type: token.SequenceEntryType,
				},
				{
					Type: token.StringType,
				},
				{
					Type: token.SequenceEntryType,
				},
				{
					Type: token.CommentType,
				},
			},
			Checkpoints: []int{1, 3},
			Commits:     nil,
			Rollbacks: []rollbackInfo{
				{
					idx: 3,
					tok: token.Token{Type: token.SpaceType},
				},
				{
					idx: 5,
					tok: token.Token{Type: token.SpaceType},
				},
			},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.Name, func(t *testing.T) {
			stream := &testStream[token.Token]{
				values:       tc.LoadedTokens,
				defaultValue: token.Token{Type: token.EOFType},
				index:        0,
			}
			result := make([]token.Token, 0, len(tc.ExpectedTokens))
			tokenAccessor := cpaccessor.NewCheckpointingAccessor[token.Token](stream)

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
				if rollbackIdx < len(tc.Rollbacks) && i == tc.Rollbacks[rollbackIdx].idx {
					restored := tokenAccessor.Rollback()
					if restored != tc.Rollbacks[rollbackIdx].tok {
						t.Errorf("wrong restored token: expected %v but got %v at position %d",
							tc.Rollbacks[rollbackIdx].tok.Type, restored.Type, i)
					}
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

type testStream[T any] struct {
	values       []T
	defaultValue T
	index        int
}

func (t *testStream[T]) Next() T {
	if t.index >= len(t.values) {
		return t.defaultValue
	}
	tok := t.values[t.index]
	t.index++
	return tok
}

// Accessor must preserve all values rollbacked with nested rollback
// after committing the last existing checkpoint
func TestCommitWithNestedRollback(t *testing.T) {
	stream := &testStream[int]{
		values: []int{1, 2, 3},
	}

	accessor := cpaccessor.NewCheckpointingAccessor[int](stream)

	accessor.SetCheckpoint()
	accessor.Next()
	accessor.Next()
	accessor.SetCheckpoint()
	accessor.Next()
	accessor.Rollback()
	accessor.Commit()

	value := accessor.Next()
	if value != 3 {
		t.Fatalf("expected %d but got %d", 3, value)
	}

}
