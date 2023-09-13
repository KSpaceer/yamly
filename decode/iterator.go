package decode

import "github.com/KSpaceer/yamly/ast"

type nodeIteratorImpl struct {
	i     int
	nodes []ast.Node
}

func (s *nodeIteratorImpl) node() ast.Node {
	if s.empty() {
		return nil
	}
	n := s.nodes[s.i]
	s.i++
	return n
}

func (s *nodeIteratorImpl) empty() bool {
	return s.i >= len(s.nodes)
}

func newStreamIterator(s *ast.StreamNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     0,
		nodes: s.Documents(),
	}
}

func newSequenceIterator(s *ast.SequenceNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     0,
		nodes: s.Entries(),
	}
}

func newMappingIterator(m *ast.MappingNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     0,
		nodes: m.Entries(),
	}
}

func newMappingEntryIterator(m *ast.MappingEntryNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     0,
		nodes: []ast.Node{m.Key(), m.Value()},
	}
}

func newPropertiesIterator(p *ast.PropertiesNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     0,
		nodes: []ast.Node{p.Anchor(), p.Tag()},
	}
}

func newContentIterator(c *ast.ContentNode) nodeIterator {
	return &nodeIteratorImpl{
		i:     0,
		nodes: []ast.Node{c.Properties(), c.Content()},
	}
}
