package decode

import (
	"gopkg.in/yaml.v3"
)

type nodeIteratorImpl struct {
	i     int
	nodes []*yaml.Node
}

func (s *nodeIteratorImpl) node() *yaml.Node {
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

func newNodeIterator(n *yaml.Node) nodeIterator {
	return &nodeIteratorImpl{
		i:     0,
		nodes: n.Content,
	}
}
