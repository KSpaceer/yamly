package yamlchar_test

import (
	"testing"

	"github.com/KSpaceer/yamly/engines/yayamls/yamlchar"
)

func TestCharSetConformation(t *testing.T) {
	t.Parallel()

	type tcase struct {
		name         string
		str          string
		yamlcharet   yamlchar.CharSetType
		wantedResult bool
	}

	tcases := []tcase{
		{
			name:         "valid decimal",
			str:          "1230049583920",
			yamlcharet:   yamlchar.DecimalCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid decimal",
			str:          "2121490141519ww",
			yamlcharet:   yamlchar.DecimalCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid word",
			str:          "word-123",
			yamlcharet:   yamlchar.WordCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid word",
			str:          "word-123/\\",
			yamlcharet:   yamlchar.WordCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid URI",
			str:          "foo://example.com:8042/path/here?name=yaml&d=%AF%FE#meme",
			yamlcharet:   yamlchar.URICharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid URI",
			str:          "foo://example.com:8042/path/here?name=yaml&d=%AF%AE#meme%l",
			yamlcharet:   yamlchar.URICharSetType,
			wantedResult: false,
		},
		{
			name:         "valid tag",
			str:          "%aa++ia23='",
			yamlcharet:   yamlchar.TagCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid tag",
			str:          "%aa++ia23='}",
			yamlcharet:   yamlchar.TagCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid anchor",
			str:          "13-33_12anchor",
			yamlcharet:   yamlchar.AnchorCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid anchor",
			str:          "13-33_12anchor,",
			yamlcharet:   yamlchar.AnchorCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid plain-safe",
			str:          "p1a|n-sAf3",
			yamlcharet:   yamlchar.PlainSafeCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid plain-safe",
			str:          "{p1a|n-uИsAf3}",
			yamlcharet:   yamlchar.PlainSafeCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid single quoted",
			str:          "''hehe\uAAAA",
			yamlcharet:   yamlchar.SingleQuotedCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid single quoted",
			str:          "''hehe\uAAAA'",
			yamlcharet:   yamlchar.SingleQuotedCharSetType,
			wantedResult: false,
		},
		{
			name:         "valid double quoted",
			str:          "\\\\escaping\\n\\\"\\xFF\\uABCDcharacters\\UAAAAAAAA",
			yamlcharet:   yamlchar.DoubleQuotedCharSetType,
			wantedResult: true,
		},
		{
			name:         "invalid double quoted",
			str:          "\\\\escaping\\n\\\"characters\\g",
			yamlcharet:   yamlchar.DoubleQuotedCharSetType,
			wantedResult: false,
		},
		{
			name:         "invalid double quoted 2",
			str:          "\\\\escaping\\n\\\"characters\"",
			yamlcharet:   yamlchar.DoubleQuotedCharSetType,
			wantedResult: false,
		},
	}

	for _, tc := range tcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if result := yamlchar.ConformsCharSet(tc.str, tc.yamlcharet); result != tc.wantedResult {
				t.Errorf("mismatch results: expected %t but got %t", tc.wantedResult, result)
			}
		})
	}
}
