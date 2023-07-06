package astutils

import "github.com/KSpaceer/fastyaml/ast"

type Comparator struct{}

func NewComparator() *Comparator {
	return &Comparator{}
}

func (c *Comparator) Equal(first, second ast.Node) bool {
	if first == nil || second == nil {
		return first == nil && second == nil
	}
	firstCh, secondCh := make(chan ast.Node), make(chan ast.Node)
	firstVisitor := &comparingVisitor{firstCh}
	secondVisitor := &comparingVisitor{secondCh}
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

func (c *Comparator) compareNodes(first, second ast.Node) bool {
	if first == nil || second == nil {
		return false
	}
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

type comparingVisitor struct {
	cmpChan chan<- ast.Node
}

func (t *comparingVisitor) VisitStreamNode(n *ast.StreamNode) {
	t.cmpChan <- n
	for _, doc := range n.Documents() {
		doc.Accept(t)
	}
}

func (t *comparingVisitor) VisitTagNode(n *ast.TagNode) {
	t.cmpChan <- n
}

func (t *comparingVisitor) VisitAnchorNode(n *ast.AnchorNode) {
	t.cmpChan <- n
}

func (t *comparingVisitor) VisitAliasNode(n *ast.AliasNode) {
	t.cmpChan <- n
}

func (t *comparingVisitor) VisitTextNode(n *ast.TextNode) {
	t.cmpChan <- n
}

func (t *comparingVisitor) VisitScalarNode(n *ast.ScalarNode) {
	t.cmpChan <- n
	properties, content := n.Properties(), n.Content()
	if properties != nil {
		properties.Accept(t)
	}
	if content != nil {
		content.Accept(t)
	}
}

func (t *comparingVisitor) VisitCollectionNode(n *ast.CollectionNode) {
	t.cmpChan <- n
	properties, collection := n.Properties(), n.Collection()
	if properties != nil {
		properties.Accept(t)
	}
	if collection != nil {
		collection.Accept(t)
	}
}

func (t *comparingVisitor) VisitSequenceNode(n *ast.SequenceNode) {
	t.cmpChan <- n
	for _, entry := range n.Entries() {
		entry.Accept(t)
	}
}

func (t *comparingVisitor) VisitMappingNode(n *ast.MappingNode) {
	t.cmpChan <- n
	for _, entry := range n.Entries() {
		entry.Accept(t)
	}
}

func (t *comparingVisitor) VisitMappingEntryNode(n *ast.MappingEntryNode) {
	t.cmpChan <- n
	key, value := n.Key(), n.Value()
	if key != nil {
		key.Accept(t)
	}
	if value != nil {
		value.Accept(t)
	}
}

func (t *comparingVisitor) VisitNullNode(n *ast.NullNode) {
	t.cmpChan <- n
}

func (t *comparingVisitor) VisitPropertiesNode(n *ast.PropertiesNode) {
	t.cmpChan <- n
	anchor, tag := n.Anchor(), n.Tag()
	if anchor != nil {
		anchor.Accept(t)
	}
	if tag != nil {
		tag.Accept(t)
	}
}
