package token

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
	MinusType
	PlusType
)

const (
	StripChompingType = MinusType
	KeepChompingType  = PlusType
)

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

type Token struct {
	Type            Type
	Start           Position
	End             Position
	Origin          string
	conformationMap conformationBitmap
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
