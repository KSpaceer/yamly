package reader

import "github.com/KSpaceer/yayamls/ast"

type nodeIteratorImpl struct {
	i     int
	nodes []ast.Node
}

func (s *nodeIteratorImpl) next() bool {
	s.i++
	return s.i < len(s.nodes)
}

func (s *nodeIteratorImpl) node() ast.Node {
	return s.nodes[s.i]
}

func (s *nodeIteratorImpl) empty() bool {
	return s.i >= len(s.nodes)
}

func newStreamIterator(s *ast.StreamNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     -1,
		nodes: s.Documents(),
	}
}

func newSequenceIterator(s *ast.SequenceNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     -1,
		nodes: s.Entries(),
	}
}

func newMappingIterator(m *ast.MappingNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     -1,
		nodes: m.Entries(),
	}
}

func newMappingEntryIterator(m *ast.MappingEntryNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     -1,
		nodes: []ast.Node{m.Key(), m.Value()},
	}
}
