package parser

import "testing"

func Test_findPkgPath(t *testing.T) {
	pkgPath, err := findPkgPath(".")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	const expectedPkgPath = "github.com/KSpaceer/yayamls/generator/parser"
	if pkgPath != expectedPkgPath {
		t.Errorf("expected %q\ngot %q", expectedPkgPath, pkgPath)
	}
}
