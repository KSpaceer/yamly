package reader

import "github.com/KSpaceer/yayamls/ast"

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

func (ak *anchorsKeeper) markAsLatestVisited(anchorName string) {
	ak.metAnchor = true
	ak.anchorName = anchorName
}

func (ak *anchorsKeeper) maybeBindToLatestVisited(n ast.Node) {
	if ak.metAnchor {
		ak.anchors[ak.anchorName] = n
		ak.metAnchor = false
	}
}

func (ak *anchorsKeeper) dereferenceAlias(alias string) (ast.Node, error) {
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
