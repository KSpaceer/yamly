package parser_test

import (
	"github.com/KSpaceer/fastyaml/ast"
	"github.com/KSpaceer/fastyaml/parser"
	"github.com/KSpaceer/fastyaml/token"
	"testing"
)

func TestParser(t *testing.T) {
	type tcase struct {
		tokens      []token.Token
		expectedAST ast.Node
	}

	tcases := []tcase{
		{
			tokens: []token.Token{
				{
					Type:   token.EOFType,
					Start:  token.Position{},
					End:    token.Position{},
					Origin: "",
				},
			},
			expectedAST: nil,
		},
	}
	cmp := newASTComparator()
	for _, tc := range tcases {
		result := parser.Parse(&testTokenStream{
			tokens: tc.tokens,
			index:  0,
		})
		if !cmp.compare(tc.expectedAST, result) {
			t.Fail()
		}
	}
}

type comparator struct{}

func newASTComparator() *comparator {
	return &comparator{}
}

func (c *comparator) compare(first, second ast.Node) bool {
	firstCh, secondCh := make(chan ast.Node), make(chan ast.Node)
	firstVisitor := &testComparingVisitor{firstCh}
	secondVisitor := &testComparingVisitor{secondCh}
	go func() {
		first.Accept(firstVisitor)
		close(firstCh)
	}()
	go func() {
		second.Accept(secondVisitor)
		close(secondCh)
	}()

	result := true
	for {
		firstNode, firstOk := <-firstCh
		secondNode, secondOk := <-secondCh
		if !firstOk && !secondOk {
			break
		}
		if !c.compareNodes(firstNode, secondNode) {
			result = false
			break
		}
	}
	for range firstCh {
	}
	for range secondCh {
	}
	return result
}

func (c *comparator) compareNodes(first, second ast.Node) bool {
	if first.Type() != second.Type() {
		return false
	}

	switch f := first.(type) {
	case ast.Texter:
		return f.Text() == second.(ast.Texter).Text()
	default:
		return true
	}
}

type testComparingVisitor struct {
	cmpChan chan<- ast.Node
}

func (t *testComparingVisitor) VisitStreamNode(n *ast.StreamNode) {
	t.cmpChan <- n
	for _, doc := range n.Documents() {
		doc.Accept(t)
	}
}

func (t *testComparingVisitor) VisitTagNode(n *ast.TagNode) {
	t.cmpChan <- n
}

func (t *testComparingVisitor) VisitAnchorNode(n *ast.AnchorNode) {
	t.cmpChan <- n
}

func (t *testComparingVisitor) VisitAliasNode(n *ast.AliasNode) {
	t.cmpChan <- n
}

func (t *testComparingVisitor) VisitTextNode(n *ast.TextNode) {
	t.cmpChan <- n
}

func (t *testComparingVisitor) VisitScalarNode(n *ast.ScalarNode) {
	t.cmpChan <- n
}

func (t *testComparingVisitor) VisitCollectionNode(n *ast.CollectionNode) {
	t.cmpChan <- n
	n.Properties().Accept(t)
	n.Collection().Accept(t)
}

func (t *testComparingVisitor) VisitSequenceNode(n *ast.SequenceNode) {
	t.cmpChan <- n
	for _, entry := range n.Entries() {
		entry.Accept(t)
	}
}

func (t *testComparingVisitor) VisitMappingNode(n *ast.MappingNode) {
	t.cmpChan <- n
	for _, entry := range n.Entries() {
		entry.Accept(t)
	}
}

func (t *testComparingVisitor) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	t.cmpChan <- n
	n.Key().Accept(t)
	n.Value().Accept(t)
}

func (t *testComparingVisitor) VisitBlockNode(n *ast.BlockNode) {
	t.cmpChan <- n
	n.Content().Accept(t)
}

func (t *testComparingVisitor) VisitNullNode(n *ast.NullNode) {
	t.cmpChan <- n
}

func (t *testComparingVisitor) VisitPropertiesNode(n *ast.PropertiesNode) {
	t.cmpChan <- n
	n.Anchor().Accept(t)
	n.Tag().Accept(t)
}
