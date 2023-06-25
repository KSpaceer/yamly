package ast

type Visitor interface {
	VisitStreamNode(n *StreamNode)
	VisitTagNode(n *TagNode)
	VisitAnchorNode(n *AnchorNode)
	VisitAliasNode(n *AliasNode)
	VisitTextNode(n *TextNode)
	VisitScalarNode(n *ScalarNode)
	VisitCollectionNode(n *CollectionNode)
	VisitSequenceNode(n *SequenceNode)
	VisitMappingNode(n *MappingNode)
	VisitMappingEntryNode(n *MappingEntryNode)
	VisitBlockNode(n *BlockNode)
	VisitNullNode(n *NullNode)
	VisitPropertiesNode(n *PropertiesNode)
}
