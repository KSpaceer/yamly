package ast

import "github.com/KSpaceer/fastyaml/token"

type NodeType int8

const (
	InvalidType NodeType = iota
	DocumentType
	ScalarType
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
	tk token.Token
}

func (t TagNode) Start() token.Position {
	return t.tk.Start
}

func (t TagNode) End() token.Position {
	return t.tk.End
}

func (TagNode) Type() NodeType {
	return TagType
}

func NewTagNode(tagToken token.Token) Node {
	return TagNode{
		tk: tagToken,
	}
}

type AnchorNode struct {
	tk token.Token
}

func (a AnchorNode) Start() token.Position {
	return a.tk.Start
}

func (a AnchorNode) End() token.Position {
	return a.tk.End
}

func (AnchorNode) Type() NodeType {
	return AnchorType
}

func NewAnchorNode(anchorToken token.Token) Node {
	return AnchorNode{
		tk: anchorToken,
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

func NewTextNode(start, end token.Position, text []byte) Node {
	return TextNode{
		start: start,
		end:   end,
		text:  text,
	}
}
