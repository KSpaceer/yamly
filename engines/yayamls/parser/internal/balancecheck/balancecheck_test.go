package balancecheck_test

import (
	"testing"

	"github.com/KSpaceer/yamly/engines/yayamls/parser/internal/balancecheck"
	"github.com/KSpaceer/yamly/engines/yayamls/token"
)

func TestParenthesesBalance(t *testing.T) {
	t.Parallel()

	type tcase struct {
		name string
		src  []token.Type
		want bool
	}

	tcases := []tcase{
		{
			name: "empty",
			src:  nil,
			want: true,
		},
		{
			name: "balanced",
			src: []token.Type{
				token.MappingStartType,
				token.SequenceStartType,
				token.SequenceEndType,
				token.MappingEndType,
			},
			want: true,
		},
		{
			name: "unbalanced",
			src: []token.Type{
				token.MappingStartType,
				token.SequenceStartType,
				token.SequenceEndType,
				token.MappingEndType,
				token.MappingStartType,
			},
			want: false,
		},
		{
			name: "intersection",
			src: []token.Type{
				token.MappingStartType,
				token.SequenceStartType,
				token.MappingEndType,
				token.SequenceEndType,
			},
			want: false,
		},
	}

	for _, tc := range tcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := balancecheck.NewBalanceChecker([][2]token.Type{
				{token.MappingStartType, token.MappingEndType},
				{token.SequenceStartType, token.SequenceEndType},
			})
			for _, r := range tc.src {
				if !b.Add(r) {
					break
				}
			}
			if b.IsBalanced() != tc.want {
				t.Errorf("expected %t but got %t", tc.want, !tc.want)
			}
		})
	}
}

func TestQuotesBalance(t *testing.T) {
	t.Parallel()

	type tcase struct {
		name string
		src  []token.Type
		want bool
	}

	tcases := []tcase{
		{
			name: "empty",
			src:  nil,
			want: true,
		},
		{
			name: "balanced",
			src: []token.Type{
				token.DoubleQuoteType,
				token.DoubleQuoteType,
			},
			want: true,
		},
		{
			name: "unbalanced",
			src: []token.Type{
				token.DoubleQuoteType,
				token.DoubleQuoteType,
				token.DoubleQuoteType,
			},
			want: false,
		},
	}
	for _, tc := range tcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := balancecheck.NewBalanceChecker([][2]token.Type{
				{token.DoubleQuoteType, token.DoubleQuoteType},
			})
			for _, r := range tc.src {
				if !b.Add(r) {
					break
				}
			}
			if b.IsBalanced() != tc.want {
				t.Errorf("expected %t but got %t", tc.want, !tc.want)
			}
		})
	}
}
