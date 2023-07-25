package token_test

import (
	"github.com/KSpaceer/yayamls/token"
	"testing"
)

func TestCharSetConformation(t *testing.T) {
	type tcase struct {
		name         string
		str          string
		charset      token.CharSetType
		wantedResult bool
	}

	tcases := []tcase{
		{
			name:         "valid decimal",
			str:          "1230049583920",
			charset:      token.DecimalCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid decimal",
			str:          "2121490141519ww",
			charset:      token.DecimalCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid word",
			str:          "word-123",
			charset:      token.WordCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid word",
			str:          "word-123/\\",
			charset:      token.WordCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid URI",
			str:          "foo://example.com:8042/path/here?name=yaml&d=%AF%FE#meme",
			charset:      token.URICharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid URI",
			str:          "foo://example.com:8042/path/here?name=yaml&d=%AF%AE#meme%l",
			charset:      token.URICharSetType,
			wantedResult: false,
		},
		{
			name:         "valid tag",
			str:          "%aa++ia23='",
			charset:      token.TagCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid tag",
			str:          "%aa++ia23='}",
			charset:      token.TagCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid anchor",
			str:          "13-33_12anchor",
			charset:      token.AnchorCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid anchor",
			str:          "13-33_12anchor,",
			charset:      token.AnchorCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid plain-safe",
			str:          "p1a|n-sAf3",
			charset:      token.PlainSafeCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid plain-safe",
			str:          "{p1a|n-uИsAf3}",
			charset:      token.PlainSafeCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid single quoted",
			str:          "''hehe\uAAAA",
			charset:      token.SingleQuotedCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid single quoted",
			str:          "''hehe\uAAAA'",
			charset:      token.SingleQuotedCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid double quoted",
			str:          "\\\\escaping\\n\\\"\\xFF\\uABCDcharacters\\UAAAAAAAA",
			charset:      token.DoubleQuotedCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid double quoted",
			str:          "\\\\escaping\\n\\\"characters\\g",
			charset:      token.DoubleQuotedCharSetType,
			wantedResult: false,
		},
		{
			name:         "invalid double quoted 2",
			str:          "\\\\escaping\\n\\\"characters\"",
			charset:      token.DoubleQuotedCharSetType,
			wantedResult: false,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			if result := token.ConformsCharSet(tc.str, tc.charset); result != tc.wantedResult {
				t.Errorf("mismatch results: expected %t but got %t", tc.wantedResult, result)
			}
		})
	}
}
