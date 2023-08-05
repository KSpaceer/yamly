package ast

import (
	"github.com/KSpaceer/yayamls/token"
)

type NodeType int8

const (
	InvalidType NodeType = iota
	DocumentType
	ContentType
	MappingType
	MappingEntryType
	SequenceType
	CommentType
	DirectiveType
	TagType
	AnchorType
	AliasType
	StreamType
	DocumentPrefixType
	DocumentSuffixType
	IndentType
	PropertiesType
	BlockHeaderType
	TextType
	NullType
)

func (t NodeType) String() string {
	switch t {
	case InvalidType:
		return "invalid"
	case DocumentType:
		return "document"
	case ContentType:
		return "content"
	case MappingType:
		return "mapping"
	case MappingEntryType:
		return "mapping_entry"
	case SequenceType:
		return "sequence"
	case CommentType:
		return "comment"
	case DirectiveType:
		return "directive"
	case TagType:
		return "tag"
	case AnchorType:
		return "anchor"
	case AliasType:
		return "alias"
	case StreamType:
		return "stream"
	case DocumentPrefixType:
		return "document_prefix"
	case DocumentSuffixType:
		return "document_suffix"
	case IndentType:
		return "indent"
	case PropertiesType:
		return "properties"
	case BlockHeaderType:
		return "block_header"
	case TextType:
		return "text"
	case NullType:
		return "null"
	}
	return ""
}

type ChompingType int8

const (
	UnknownChompingType ChompingType = iota
	ClipChompingType
	StripChompingType
	KeepChompingType
)

func TokenChompingType(tok token.Token) ChompingType {
	switch tok.Type {
	case token.StripChompingType:
		return StripChompingType
	case token.KeepChompingType:
		return KeepChompingType
	}
	return UnknownChompingType
}

type QuotingType int8

const (
	UnknownQuotingType QuotingType = iota
	AbsentQuotingType
	SingleQuotingType
	DoubleQuotingType
)

type Node interface {
	Type() NodeType
	Accept(v Visitor)
}

type Texter interface {
	Text() string
}

func ValidNode(n Node) bool {
	return n != nil && n.Type() != InvalidType
}

type BasicNode struct {
	NodeType NodeType
}

func (b *BasicNode) Type() NodeType {
	return b.NodeType
}

func (*BasicNode) Accept(Visitor) {}

func NewInvalidNode() Node {
	return &BasicNode{
		NodeType: InvalidType,
	}
}

func NewBasicNode(tp NodeType) Node {
	return &BasicNode{
		NodeType: tp,
	}
}

type StreamNode struct {
	documents []Node
}

func (*StreamNode) Type() NodeType {
	return StreamType
}

func (s *StreamNode) Accept(v Visitor) {
	v.VisitStreamNode(s)
}

func (s *StreamNode) Documents() []Node {
	return s.documents
}

func NewStreamNode(documents []Node) Node {
	return &StreamNode{
		documents: documents,
	}
}

type PropertiesNode struct {
	tag    Node
	anchor Node
}

func (*PropertiesNode) Type() NodeType {
	return PropertiesType
}

func (p *PropertiesNode) Accept(v Visitor) {
	v.VisitPropertiesNode(p)
}

func (p *PropertiesNode) Tag() Node {
	return p.tag
}

func (p *PropertiesNode) Anchor() Node {
	return p.anchor
}

func NewPropertiesNode(tag, anchor Node) Node {
	return &PropertiesNode{
		tag:    tag,
		anchor: anchor,
	}
}

type TagNode struct {
	text string
}

func (*TagNode) Type() NodeType {
	return TagType
}

func (t *TagNode) Accept(v Visitor) {
	v.VisitTagNode(t)
}

func (t *TagNode) Text() string {
	return t.text
}

func NewTagNode(text string) Node {
	return &TagNode{
		text: text,
	}
}

type AnchorNode struct {
	text string
}

func (*AnchorNode) Type() NodeType {
	return AnchorType
}

func (a *AnchorNode) Accept(v Visitor) {
	v.VisitAnchorNode(a)
}

func (a *AnchorNode) Text() string {
	return a.text
}

func NewAnchorNode(text string) Node {
	return &AnchorNode{
		text: text,
	}
}

type AliasNode struct {
	text string
}

func (*AliasNode) Type() NodeType {
	return AliasType
}

func (a *AliasNode) Accept(v Visitor) {
	v.VisitAliasNode(a)
}

func (a *AliasNode) Text() string {
	return a.text
}

func NewAliasNode(text string) Node {
	return &AliasNode{
		text: text,
	}
}

type BlockHeaderNode struct {
	indentation  int
	chompingType ChompingType
}

func (*BlockHeaderNode) Type() NodeType {
	return BlockHeaderType
}

func (*BlockHeaderNode) Accept(Visitor) {}

func (b *BlockHeaderNode) IndentationIndicator() int {
	return b.indentation
}

func (b *BlockHeaderNode) ChompingIndicator() ChompingType {
	return b.chompingType
}

func NewBlockHeaderNode(chomping ChompingType, indentation int) Node {
	return &BlockHeaderNode{
		indentation:  indentation,
		chompingType: chomping,
	}
}

type TextNodeOption interface {
	apply(*TextNode)
}

type textNodeOptionFunc func(*TextNode)

func (f textNodeOptionFunc) apply(o *TextNode) {
	f(o)
}

func WithQuotingType(t QuotingType) TextNodeOption {
	return textNodeOptionFunc(func(node *TextNode) {
		node.quotingType = t
	})
}

type TextNode struct {
	quotingType QuotingType
	text        string
}

func (*TextNode) Type() NodeType {
	return TextType
}

func (t *TextNode) QuotingType() QuotingType {
	return t.quotingType
}

func (t *TextNode) Accept(v Visitor) {
	v.VisitTextNode(t)
}

func (t *TextNode) Text() string {
	return t.text
}

func NewTextNode(text string, opts ...TextNodeOption) Node {
	node := TextNode{
		quotingType: AbsentQuotingType,
		text:        text,
	}
	for _, opt := range opts {
		opt.apply(&node)
	}
	return &node
}

type ContentNode struct {
	properties Node
	content    Node
}

func (*ContentNode) Type() NodeType {
	return ContentType
}

func (c *ContentNode) Accept(v Visitor) {
	v.VisitContentNode(c)
}

func (c *ContentNode) Properties() Node {
	return c.properties
}

func (c *ContentNode) Content() Node {
	return c.content
}

func NewContentNode(properties, content Node) Node {
	return &ContentNode{
		properties: properties,
		content:    content,
	}
}

type IndentNode struct {
	indent int
}

func (*IndentNode) Type() NodeType {
	return IndentType
}

func (i *IndentNode) Indent() int {
	return i.indent
}

func (*IndentNode) Accept(Visitor) {}

func NewIndentNode(indent int) Node {
	return &IndentNode{
		indent: indent,
	}
}

type SequenceNode struct {
	entries []Node
}

func (*SequenceNode) Type() NodeType {
	return SequenceType
}

func (s *SequenceNode) Accept(v Visitor) {
	v.VisitSequenceNode(s)
}

func (s *SequenceNode) Entries() []Node {
	return s.entries
}

func NewSequenceNode(entries []Node) Node {
	return &SequenceNode{entries: entries}
}

type MappingNode struct {
	entries []Node
}

func (*MappingNode) Type() NodeType {
	return MappingType
}

func (m *MappingNode) Accept(v Visitor) {
	v.VisitMappingNode(m)
}

func (m *MappingNode) Entries() []Node {
	return m.entries
}

func NewMappingNode(entries []Node) Node {
	return &MappingNode{entries: entries}
}

type MappingEntryNode struct {
	key, value Node
}

func (*MappingEntryNode) Type() NodeType {
	return MappingEntryType
}

func (m *MappingEntryNode) Accept(v Visitor) {
	v.VisitMappingEntryNode(m)
}

func (m *MappingEntryNode) Key() Node {
	return m.key
}

func (m *MappingEntryNode) Value() Node {
	return m.value
}

func NewMappingEntryNode(key, value Node) Node {
	return &MappingEntryNode{key: key, value: value}
}

type NullNode struct{}

func (*NullNode) Type() NodeType {
	return NullType
}

func (n *NullNode) Accept(v Visitor) {
	v.VisitNullNode(n)
}

func NewNullNode() Node {
	return &NullNode{}
}
