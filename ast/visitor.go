package ast

type Visitor interface {
	VisitStreamNode(n *StreamNode)
	VisitTagNode(n *TagNode)
	VisitAnchorNode(n *AnchorNode)
	VisitAliasNode(n *AliasNode)
	VisitTextNode(n *TextNode)
	VisitSequenceNode(n *SequenceNode)
	VisitMappingNode(n *MappingNode)
	VisitMappingEntryNode(n *MappingEntryNode)
	VisitNullNode(n *NullNode)
	VisitPropertiesNode(n *PropertiesNode)
	VisitContentNode(n *ContentNode)
}
