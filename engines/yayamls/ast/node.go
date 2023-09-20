// Package ast contains types for YAML AST representation.
package ast

import (
	"github.com/KSpaceer/yamly/engines/yayamls/token"
)

// NodeType represents the type of element in AST
type NodeType int8

const (
	// InvalidType embodies an erroneous element of AST. They usually appear when an error occures during parsing.
	InvalidType NodeType = iota
	// DocumentType represents a YAML document.
	DocumentType
	// ContentType represents a content element (scalar, sequence or mapping) with properties (tag, anchor).
	ContentType
	// MappingType represents a YAML mapping.
	MappingType
	// MappingEntryType represents a single mapping entry (key-value pair).
	MappingEntryType
	// SequenceType represents a YAML sequence.
	SequenceType
	// CommentType represents a YAML comment.
	CommentType
	// DirectiveType represents a YAML directive.
	DirectiveType
	// TagType represents a YAML tag.
	TagType
	// AnchorType represents a YAML anchor.
	AnchorType
	// AliasType represents a YAML alias.
	AliasType
	// StreamType represents a YAML stream
	StreamType
	// DocumentPrefixType represents a YAML document prefix (BOM + comments)
	DocumentPrefixType
	// DocumentSuffixType represents a YAML document suffix (document end + comments)
	DocumentSuffixType
	// IndentType represents a YAML indentation.
	IndentType
	// PropertiesType represents element's properties (tag, anchor).
	PropertiesType
	// BlockHeaderType represents literal/folded properties (chomping, explicit indent indicator)
	BlockHeaderType
	// TextType represents a text, i.e. single scalar
	TextType
	// NullType represents a null scalar.
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

// ChompingType corresponds to YAML chomping type and defines
// what to do with trailing newlines in literal/folded style.
type ChompingType int8

const (
	UnknownChompingType ChompingType = iota
	// ClipChompingType prescribes removing of trailing newlines except the first one
	ClipChompingType
	// StripChompingType prescribes removing of all trailing newlines
	StripChompingType
	// KeepChompingType prescribes keeping of all trailing newlines
	KeepChompingType
)

// TokenChompingType derives chomping type from given token
func TokenChompingType(tok token.Token) ChompingType {
	switch tok.Type {
	case token.StripChompingType:
		return StripChompingType
	case token.KeepChompingType:
		return KeepChompingType
	}
	return UnknownChompingType
}

// QuotingType represents a quoting type of string.
type QuotingType int8

const (
	UnknownQuotingType QuotingType = iota
	// AbsentQuotingType means that string has no quotes
	AbsentQuotingType
	// SingleQuotingType means that string is enclosed in single quotes
	SingleQuotingType
	// DoubleQuotingType means that string is enclosed in double quotes
	DoubleQuotingType
)

// Node is a single element of YAML AST
type Node interface {
	// Type returns the node's type
	Type() NodeType
	// Accept implements "Visitor" pattern for AST.
	Accept(v Visitor)
}

// Texter is a more specific kind of Node that has some meaningful string data
type Texter interface {
	// Text returns string data associated with Texter node
	Text() string
}

// TexterNode is a composite of Node and Texter interfaces
type TexterNode interface {
	Texter
	Node
}

// ValidNode checks if Node is valid and can be processed.
func ValidNode(n Node) bool {
	return n != nil && n.Type() != InvalidType
}

// BasicNode is a simple container for given type and used during parsing.
type BasicNode struct {
	NodeType NodeType
}

func (b *BasicNode) Type() NodeType {
	return b.NodeType
}

func (*BasicNode) Accept(Visitor) {}

var invalidNode = &BasicNode{NodeType: InvalidType}

func NewInvalidNode() Node {
	return invalidNode
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

// IndentationIndicator returns an indentation value that was explicitly set in block header.
func (b *BlockHeaderNode) IndentationIndicator() int {
	return b.indentation
}

// ChompingIndicator returns a chomping type defines in block header.
func (b *BlockHeaderNode) ChompingIndicator() ChompingType {
	return b.chompingType
}

func NewBlockHeaderNode(chomping ChompingType, indentation int) Node {
	return &BlockHeaderNode{
		indentation:  indentation,
		chompingType: chomping,
	}
}

// TextNodeOption allows to modify YAML element associated with TextNode during creation.
type TextNodeOption interface {
	apply(*TextNode)
}

type textNodeOptionFunc func(*TextNode)

func (f textNodeOptionFunc) apply(o *TextNode) {
	f(o)
}

// WithQuotingType sets given QuotingType for TextNode string.
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
		text: text,
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

func (s *SequenceNode) AppendEntry(n Node) {
	s.entries = append(s.entries, n)
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

func (m *MappingNode) AppendEntry(n Node) {
	m.entries = append(m.entries, n)
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

func (m *MappingEntryNode) SetKey(n Node) {
	m.key = n
}

func (m *MappingEntryNode) SetValue(n Node) {
	m.value = n
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

var nullNode = &NullNode{}

func NewNullNode() Node {
	return nullNode
}
