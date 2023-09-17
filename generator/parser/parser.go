// Package parser contains parser for package name and path.
package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
)

// Parser is used to parse package name and path from provided directory.
type Parser struct {
	PkgName string
	PkgPath string
}

// Parse parses files in dirPath and sets PkgName and PkgPath.
func (p *Parser) Parse(dirPath string) error {
	var err error
	if p.PkgPath, err = findPkgPath(dirPath); err != nil {
		return err
	}
	packages, err := parser.ParseDir(
		token.NewFileSet(),
		dirPath,
		func(info fs.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		},
		parser.PackageClauseOnly,
	)
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		ast.Walk((*pkgNameVisitor)(&p.PkgName), pkg)
	}
	return nil
}

type pkgNameVisitor string

func (v *pkgNameVisitor) Visit(n ast.Node) (w ast.Visitor) {
	switch n := n.(type) {
	case *ast.Package:
		return v
	case *ast.File:
		*v = pkgNameVisitor(n.Name.String())
		return v
	}
	return nil
}
