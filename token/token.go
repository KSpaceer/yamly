package token

import "strconv"

const (
	YAMLDirective = "YAML"
	TagDirective  = "TAG"
)

type Character = rune

const (
	SequenceEntryCharacter     Character = '-'
	MappingKeyCharacter        Character = '?'
	MappingValueCharacter      Character = ':'
	CollectEntryCharacter      Character = ','
	SequenceStartCharacter     Character = '['
	SequenceEndCharacter       Character = ']'
	MappingStartCharacter      Character = '{'
	MappingEndCharacter        Character = '}'
	CommentCharacter           Character = '#'
	AnchorCharacter            Character = '&'
	AliasCharacter             Character = '*'
	TagCharacter               Character = '!'
	LiteralCharacter           Character = '|'
	FoldedCharacter            Character = '>'
	SingleQuoteCharacter       Character = '\''
	DoubleQuoteCharacter       Character = '"'
	DirectiveCharacter         Character = '%'
	ReservedAtCharacter        Character = '@'
	ReservedBackquoteCharacter Character = '`'
	LineFeedCharacter          Character = '\n'
	CarriageReturnCharacter    Character = '\r'
	SpaceCharacter             Character = ' '
	TabCharacter               Character = '\t'
	EscapeCharacter            Character = '\\'
	DotCharacter               Character = '.'
	ByteOrderMarkCharacter     Character = 0xFEFF
	DirectiveEndCharacter      Character = '-'
	StripChompingCharacter     Character = '-'
	KeepChompingCharacter      Character = '+'
	DocumentEndCharacter       Character = '.'
)

type Type uint8

const (
	UnknownType Type = iota
	SequenceEntryType
	MappingKeyType
	MappingValueType
	CollectEntryType
	SequenceStartType
	SequenceEndType
	MappingStartType
	MappingEndType
	CommentType
	AnchorType
	AliasType
	TagType
	LiteralType
	FoldedType
	SingleQuoteType
	DoubleQuoteType
	DirectiveType
	LineBreakType
	SpaceType
	TabType
	BOMType
	EOFType
	DocumentEndType
	DirectiveEndType
	StringType
	StripChompingType
	KeepChompingType
)

func (t Type) String() string {
	switch t {
	case UnknownType:
		return "unknown"
	case SequenceEntryType:
		return "sequence-entry"
	case MappingKeyType:
		return "mapping-key"
	case MappingValueType:
		return "mapping-value"
	case CollectEntryType:
		return "collect-entry"
	case SequenceStartType:
		return "sequence-start"
	case SequenceEndType:
		return "sequence-end"
	case MappingStartType:
		return "mapping-start"
	case MappingEndType:
		return "mapping-end"
	case CommentType:
		return "comment"
	case AnchorType:
		return "anchor"
	case AliasType:
		return "alias"
	case TagType:
		return "tag"
	case LiteralType:
		return "literal"
	case FoldedType:
		return "folded"
	case SingleQuoteType:
		return "single-quote"
	case DoubleQuoteType:
		return "double-quote"
	case DirectiveType:
		return "directive"
	case LineBreakType:
		return "line-break"
	case SpaceType:
		return "space"
	case TabType:
		return "tab"
	case BOMType:
		return "byte-order-mark"
	case EOFType:
		return "end-of-file"
	case DocumentEndType:
		return "document-end"
	case DirectiveEndType:
		return "directive-end"
	case StringType:
		return "string"
	case KeepChompingType:
		return "keep-chomping"
	case StripChompingType:
		return "strip-chomping"
	default:
		return ""
	}
}

type CharSetType int16

const (
	UnknownCharSetType CharSetType = 0
)

const (
	DecimalCharSetType CharSetType = 1 << iota
	WordCharSetType
	URICharSetType
	TagCharSetType
	AnchorCharSetType
	PlainSafeCharSetType
	SingleQuotedCharSetType
	DoubleQuotedCharSetType
)

func IsWhiteSpace(tok Token) bool {
	switch tok.Type {
	case SpaceType, TabType:
		return true
	default:
		return false
	}
}

func IsNonBreak(tok Token) bool {
	switch tok.Type {
	case BOMType, LineBreakType, EOFType:
		return false
	default:
		return true
	}
}

func MayPrecedeWord(tok Token) bool {
	switch tok.Type {
	case SpaceType, TabType, LineBreakType, UnknownType:
		return true
	default:
		return false
	}
}

func IsOpeningFlowIndicator(tok Token) bool {
	switch tok.Type {
	case MappingStartType, SequenceStartType, CollectEntryType:
		return true
	default:
		return false
	}
}

func IsClosingFlowIndicator(tok Token) bool {
	switch tok.Type {
	case MappingEndType, SequenceEndType, CollectEntryType:
		return true
	default:
		return false
	}
}

func IsFlowIndicator(tok Token) bool {
	return IsOpeningFlowIndicator(tok) || IsClosingFlowIndicator(tok)
}

type Token struct {
	Type            Type
	Start           Position
	End             Position
	Origin          string
	conformationMap conformationBitmap
}

func (t Token) String() string {
	return t.Type.String() + "[" + t.Origin + "]" + " Start:" +
		t.Start.String() + " End:" + t.End.String() + "\n"
}

func (t *Token) ConformsCharSet(cst CharSetType) bool {
	result, ok := t.conformationMap.Get(cst)
	if ok {
		return result
	}
	return t.slowConformation(cst)
}

func (t *Token) slowConformation(cst CharSetType) bool {
	result := ConformsCharSet(t.Origin, cst)
	t.conformationMap = t.conformationMap.Set(cst, result)
	return result
}

type Position struct {
	Row    int
	Column int
}

func (p Position) String() string {
	return "{{Row: " + strconv.Itoa(p.Row) + ", Column: " + strconv.Itoa(p.Column) + "}}"
}
