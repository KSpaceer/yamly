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
	SequenceType
	CommentType
	DirectiveType
	TagType
	AnchorType
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

type NodePosition struct {
	StartPos token.Position
	EndPos   token.Position
}

type BasicNode struct {
	NodePosition
	NodeType NodeType
}

func (b *BasicNode) Start() token.Position {
	return b.StartPos
}

func (b *BasicNode) End() token.Position {
	return b.EndPos
}

func (b *BasicNode) Type() NodeType {
	return b.NodeType
}

func NewInvalidNode(start, end token.Position) Node {
	return &BasicNode{
		NodePosition: NodePosition{
			StartPos: start,
			EndPos:   end,
		},
		NodeType: InvalidType,
	}
}

func NewBasicNode(start, end token.Position, tp NodeType) Node {
	return &BasicNode{
		NodePosition: NodePosition{
			StartPos: start,
			EndPos:   end,
		},
		NodeType: tp,
	}
}

type StreamNode struct {
	NodePosition
	Documents []Node
}

func (s StreamNode) Start() token.Position {
	return s.StartPos
}

func (s StreamNode) End() token.Position {
	return s.EndPos
}

func (StreamNode) Type() NodeType {
	return StreamType
}

type CommentNode struct {
	StartPos token.Position
	EndPos   token.Position
}

type PropertiesNode struct {
	NodePosition
	Tag    Node
	Anchor Node
}

func (p PropertiesNode) Start() token.Position {
	return p.StartPos
}

func (p PropertiesNode) End() token.Position {
	return p.EndPos
}

func (PropertiesNode) Type() NodeType {
	return PropertiesType
}

func NewPropertiesNode(start, end token.Position, tag, anchor Node) Node {
	return PropertiesNode{
		NodePosition: NodePosition{
			StartPos: start,
			EndPos:   end,
		},
		Tag:    tag,
		Anchor: anchor,
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
	text       []byte
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

func (t TextNode) Text() []byte {
	return t.text
}

func NewTextNode(start, end token.Position, text []byte) Node {
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
	start, end token.Position
}

func (n NullNode) Start() token.Position {
	return n.start
}

func (n NullNode) End() token.Position {
	return n.end
}

func (NullNode) Type() NodeType {
	return NullType
}

func NewNullNode(start token.Position, end token.Position) Node {
	return NullNode{start: start, end: end}
}
