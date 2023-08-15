package reader

import "github.com/KSpaceer/yayamls/ast"

type streamIterator struct {
	i    int
	docs []ast.Node
}

func newStreamIterator(s *ast.StreamNode) nodeIterator {
	return &streamIterator{
		i:    -1,
		docs: s.Documents(),
	}
}

func (s *streamIterator) next() bool {
	s.i++
	return s.i < len(s.docs)
}

func (s *streamIterator) node() ast.Node {
	return s.docs[s.i]
}

func (s *streamIterator) empty() bool {
	return s.i >= len(s.docs)
}
