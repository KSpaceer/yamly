package ast

import (
	"github.com/KSpaceer/fastyaml/token"
)

type NodeType int8

const (
	InvalidType NodeType = iota
	DocumentType
	BlockType
	ScalarType
	CollectionType
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
	FloatNumberType
	IndentType
	PropertiesType
	BlockHeaderType
	TextType
	NullType
)

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

type Node interface {
	Start() token.Position
	End() token.Position
	Type() NodeType
}

func ValidNode(n Node) bool {
	return n != nil && n.Type() != InvalidType
}

type BasicNode struct {
	start, end token.Position
	NodeType   NodeType
}

func (b *BasicNode) Start() token.Position {
	return b.start
}

func (b *BasicNode) End() token.Position {
	return b.end
}

func (b *BasicNode) Type() NodeType {
	return b.NodeType
}

func NewInvalidNode(start, end token.Position) Node {
	return &BasicNode{
		start:    start,
		end:      end,
		NodeType: InvalidType,
	}
}

func NewBasicNode(start, end token.Position, tp NodeType) Node {
	return &BasicNode{
		start:    start,
		end:      end,
		NodeType: tp,
	}
}

type StreamNode struct {
	start, end token.Position
	documents  []Node
}

func (s StreamNode) Start() token.Position {
	return s.start
}

func (s StreamNode) End() token.Position {
	return s.end
}

func (StreamNode) Type() NodeType {
	return StreamType
}

func (s StreamNode) Documents() []Node {
	return s.documents
}

func NewStreamNode(start, end token.Position, documents []Node) Node {
	return StreamNode{
		start:     start,
		end:       end,
		documents: documents,
	}
}

type PropertiesNode struct {
	start, end token.Position
	tag        Node
	anchor     Node
}

func (p PropertiesNode) Start() token.Position {
	return p.start
}

func (p PropertiesNode) End() token.Position {
	return p.end
}

func (PropertiesNode) Type() NodeType {
	return PropertiesType
}

func (p PropertiesNode) Tag() Node {
	return p.tag
}

func (p PropertiesNode) Anchor() Node {
	return p.anchor
}

func NewPropertiesNode(start, end token.Position, tag, anchor Node) Node {
	return PropertiesNode{
		start:  start,
		end:    end,
		tag:    tag,
		anchor: anchor,
	}
}

type TagNode struct {
	start, end token.Position
	text       string
}

func (t TagNode) Start() token.Position {
	return t.start
}

func (t TagNode) End() token.Position {
	return t.end
}

func (TagNode) Type() NodeType {
	return TagType
}

func NewTagNode(start, end token.Position, text string) Node {
	return TagNode{
		start: start,
		end:   end,
		text:  text,
	}
}

type AnchorNode struct {
	start, end token.Position
	text       string
}

func (a AnchorNode) Start() token.Position {
	return a.start
}

func (a AnchorNode) End() token.Position {
	return a.end
}

func (AnchorNode) Type() NodeType {
	return AnchorType
}

func NewAnchorNode(start, end token.Position, text string) Node {
	return AnchorNode{
		start: start,
		end:   end,
		text:  text,
	}
}

type AliasNode struct {
	start, end token.Position
	text       string
}

func (a AliasNode) Start() token.Position {
	return a.start
}

func (a AliasNode) End() token.Position {
	return a.end
}

func (AliasNode) Type() NodeType {
	return AliasType
}

func NewAliasNode(start, end token.Position, text string) Node {
	return AliasNode{
		start: start,
		end:   end,
		text:  text,
	}
}

type BlockHeaderNode struct {
	start, end   token.Position
	indentation  int
	chompingType ChompingType
}

func (b BlockHeaderNode) Start() token.Position {
	return b.start
}

func (b BlockHeaderNode) End() token.Position {
	return b.end
}

func (BlockHeaderNode) Type() NodeType {
	return BlockHeaderType
}

func (b BlockHeaderNode) IndentationIndicator() int {
	return b.indentation
}

func (b BlockHeaderNode) ChompingIndicator() ChompingType {
	return b.chompingType
}

func NewBlockHeaderNode(start, end token.Position, chomping ChompingType, indentation int) BlockHeaderNode {
	return BlockHeaderNode{
		start:        start,
		end:          end,
		indentation:  indentation,
		chompingType: chomping,
	}
}

type TextNode struct {
	start, end token.Position
	text       string
}

func (t TextNode) Start() token.Position {
	return t.start
}

func (t TextNode) End() token.Position {
	return t.end
}

func (TextNode) Type() NodeType {
	return TextType
}

func (t TextNode) Text() string {
	return t.text
}

func NewTextNode(start, end token.Position, text string) Node {
	return TextNode{
		start: start,
		end:   end,
		text:  text,
	}
}

type ScalarNode struct {
	start, end token.Position
	properties Node
	content    Node
}

func (s ScalarNode) Start() token.Position {
	return s.start
}

func (s ScalarNode) End() token.Position {
	return s.end
}

func (ScalarNode) Type() NodeType {
	return ScalarType
}

func NewScalarNode(start, end token.Position, properties, content Node) Node {
	return ScalarNode{
		start:      start,
		end:        end,
		properties: properties,
		content:    content,
	}
}

type CollectionNode struct {
	start, end token.Position
	properties Node
	collection Node
}

func (c CollectionNode) Start() token.Position {
	return c.start
}

func (c CollectionNode) End() token.Position {
	return c.end
}

func (CollectionNode) Type() NodeType {
	return CollectionType
}

func NewCollectionNode(start, end token.Position, properties, collection Node) Node {
	return CollectionNode{
		start:      start,
		end:        end,
		properties: properties,
		collection: collection,
	}
}

type IndentNode struct {
	start, end token.Position
	indent     int
}

func (i IndentNode) Start() token.Position {
	return i.start
}

func (i IndentNode) End() token.Position {
	return i.end
}

func (IndentNode) Type() NodeType {
	return IndentType
}

func (i IndentNode) Indent() int {
	return i.indent
}

func NewIndentNode(start, end token.Position, indent int) Node {
	return IndentNode{
		start:  start,
		end:    end,
		indent: indent,
	}
}

type SequenceNode struct {
	start, end token.Position
	entries    []Node
}

func (s SequenceNode) Start() token.Position {
	return s.start
}

func (s SequenceNode) End() token.Position {
	return s.end
}

func (SequenceNode) Type() NodeType {
	return SequenceType
}

func (s SequenceNode) Entries() []Node {
	return s.entries
}

func NewSequenceNode(start token.Position, end token.Position, entries []Node) Node {
	return SequenceNode{start: start, end: end, entries: entries}
}

type MappingNode struct {
	start, end token.Position
	entries    []Node
}

func (m MappingNode) Start() token.Position {
	return m.start
}

func (m MappingNode) End() token.Position {
	return m.end
}

func (MappingNode) Type() NodeType {
	return MappingType
}

func (m MappingNode) Entries() []Node {
	return m.entries
}

func NewMappingNode(start token.Position, end token.Position, entries []Node) Node {
	return MappingNode{start: start, end: end, entries: entries}
}

type MappingEntryNode struct {
	start, end token.Position
	key, value Node
}

func (m MappingEntryNode) Start() token.Position {
	return m.start
}

func (m MappingEntryNode) End() token.Position {
	return m.end
}

func (MappingEntryNode) Type() NodeType {
	return MappingEntryType
}

func NewMappingEntryNode(start, end token.Position, key, value Node) Node {
	return MappingEntryNode{start: start, end: end, key: key, value: value}
}

type BlockNode struct {
	start, end token.Position
	content    Node
}

func (b BlockNode) Start() token.Position {
	return b.start
}

func (b BlockNode) End() token.Position {
	return b.end
}

func (BlockNode) Type() NodeType {
	return BlockType
}

func (b BlockNode) Content() Node {
	return b.content
}

func NewBlockNode(start, end token.Position, content Node) Node {
	return BlockNode{
		start:   start,
		end:     end,
		content: content,
	}
}

type NullNode struct {
	pos token.Position
}

func (n NullNode) Start() token.Position {
	return n.pos
}

func (n NullNode) End() token.Position {
	return n.pos
}

func (NullNode) Type() NodeType {
	return NullType
}

func NewNullNode(pos token.Position) Node {
	return NullNode{pos: pos}
}
