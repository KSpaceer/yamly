package chars_test

import (
	"github.com/KSpaceer/yamly/engines/yayamls/chars"
	"testing"
)

func TestCharSetConformation(t *testing.T) {
	type tcase struct {
		name         string
		str          string
		charset      chars.CharSetType
		wantedResult bool
	}

	tcases := []tcase{
		{
			name:         "valid decimal",
			str:          "1230049583920",
			charset:      chars.DecimalCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid decimal",
			str:          "2121490141519ww",
			charset:      chars.DecimalCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid word",
			str:          "word-123",
			charset:      chars.WordCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid word",
			str:          "word-123/\\",
			charset:      chars.WordCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid URI",
			str:          "foo://example.com:8042/path/here?name=yaml&d=%AF%FE#meme",
			charset:      chars.URICharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid URI",
			str:          "foo://example.com:8042/path/here?name=yaml&d=%AF%AE#meme%l",
			charset:      chars.URICharSetType,
			wantedResult: false,
		},
		{
			name:         "valid tag",
			str:          "%aa++ia23='",
			charset:      chars.TagCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid tag",
			str:          "%aa++ia23='}",
			charset:      chars.TagCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid anchor",
			str:          "13-33_12anchor",
			charset:      chars.AnchorCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid anchor",
			str:          "13-33_12anchor,",
			charset:      chars.AnchorCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid plain-safe",
			str:          "p1a|n-sAf3",
			charset:      chars.PlainSafeCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid plain-safe",
			str:          "{p1a|n-uИsAf3}",
			charset:      chars.PlainSafeCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid single quoted",
			str:          "''hehe\uAAAA",
			charset:      chars.SingleQuotedCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid single quoted",
			str:          "''hehe\uAAAA'",
			charset:      chars.SingleQuotedCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid double quoted",
			str:          "\\\\escaping\\n\\\"\\xFF\\uABCDcharacters\\UAAAAAAAA",
			charset:      chars.DoubleQuotedCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid double quoted",
			str:          "\\\\escaping\\n\\\"characters\\g",
			charset:      chars.DoubleQuotedCharSetType,
			wantedResult: false,
		},
		{
			name:         "invalid double quoted 2",
			str:          "\\\\escaping\\n\\\"characters\"",
			charset:      chars.DoubleQuotedCharSetType,
			wantedResult: false,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			if result := chars.ConformsCharSet(tc.str, tc.charset); result != tc.wantedResult {
				t.Errorf("mismatch results: expected %t but got %t", tc.wantedResult, result)
			}
		})
	}
}
