package decode

import "github.com/KSpaceer/yamly/ast"

type anchorsKeeper struct {
	anchors    map[string]ast.Node
	metAnchor  bool
	anchorName string
}

func newAnchorsKeeper() anchorsKeeper {
	return anchorsKeeper{
		anchors: map[string]ast.Node{},
	}
}

func (ak *anchorsKeeper) StoreAnchor(anchorName string) {
	ak.metAnchor = true
	ak.anchorName = anchorName
}

func (ak *anchorsKeeper) BindToLatestAnchor(n ast.Node) {
	if ak.metAnchor {
		ak.anchors[ak.anchorName] = n
		ak.metAnchor = false
	}
}

func (ak *anchorsKeeper) DereferenceAlias(alias string) (ast.Node, error) {
	anchored, ok := ak.anchors[alias]
	if !ok {
		return nil, AliasDereferenceError{name: alias}
	}
	return anchored, nil
}

func (ak *anchorsKeeper) clear() {
	clear(ak.anchors)
	ak.metAnchor = false
	ak.anchorName = ""
}
