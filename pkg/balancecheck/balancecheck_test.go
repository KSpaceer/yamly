package balancecheck_test

import (
	"github.com/KSpaceer/yayamls/pkg/balancecheck"
	"testing"
)

func TestParenthesesBalance(t *testing.T) {
	type tcase struct {
		name string
		src  string
		want bool
	}

	tcases := []tcase{
		{
			name: "empty",
			src:  "",
			want: true,
		},
		{
			name: "balanced",
			src:  "{([])}",
			want: true,
		},
		{
			name: "unbalanced",
			src:  "{([])}(",
			want: false,
		},
		{
			name: "intersection",
			src:  "{(})",
			want: false,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := balancecheck.NewBalanceChecker([][2]rune{
				{'{', '}'},
				{'[', ']'},
				{'(', ')'},
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
	type tcase struct {
		name string
		src  string
		want bool
	}

	tcases := []tcase{
		{
			name: "empty",
			src:  "",
			want: true,
		},
		{
			name: "balanced",
			src:  `""`,
			want: true,
		},
		{
			name: "unbalanced",
			src:  `"""`,
			want: false,
		},
	}
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			b := balancecheck.NewBalanceChecker([][2]rune{
				{'"', '"'},
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
