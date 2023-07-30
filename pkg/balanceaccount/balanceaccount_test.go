package balanceaccount_test

import (
	"github.com/KSpaceer/yayamls/pkg/balanceaccount"
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
			b := balanceaccount.NewBalancer([][2]rune{
				{'{', '}'},
				{'[', ']'},
				{'(', ')'},
			})
			for _, r := range tc.src {
				if !b.AccountRune(r) {
					break
				}
			}
			if b.IsBalanced() != tc.want {
				t.Errorf("expected %t but got %t", tc.want, !tc.want)
			}
		})
	}

}
