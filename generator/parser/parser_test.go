package parser_test

import (
	"github.com/KSpaceer/yayamls/generator/parser"
	"testing"
)

func TestParser_Parse(t *testing.T) {
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
			expectedPkgPath: "github.com/KSpaceer/yayamls/generator/parser",
			expectedPkgName: "parser",
		},
		{
			name:            "module root",
			dirPath:         "../..",
			expectedPkgPath: "github.com/KSpaceer/yayamls",
			expectedPkgName: "yayamls",
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
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
