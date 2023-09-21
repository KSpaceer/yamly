package parser_test

import (
	"testing"

	"github.com/KSpaceer/yamly/generator/parser"
)

func TestParser_Parse(t *testing.T) {
	t.Parallel()

	type tcase struct {
		name            string
		dirPath         string
		expectedPkgPath string
		expectedPkgName string
	}

	tcases := []tcase{
		{
			name:            "current directory",
			dirPath:         ".",
			expectedPkgPath: "github.com/KSpaceer/yamly/generator/parser",
			expectedPkgName: "parser",
		},
		{
			name:            "module root",
			dirPath:         "../..",
			expectedPkgPath: "github.com/KSpaceer/yamly",
			expectedPkgName: "yamly",
		},
	}

	for _, tc := range tcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := parser.Parser{}
			err := p.Parse(tc.dirPath)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if p.PkgPath != tc.expectedPkgPath {
				t.Errorf("wrong package path:\n\texpected: %s\n\tgot: %s", tc.expectedPkgPath, p.PkgPath)
			}
			if p.PkgName != tc.expectedPkgName {
				t.Errorf("wrong package name:\n\texpected: %s\n\tgot: %s", tc.expectedPkgName, p.PkgName)
			}
		})
	}
}
