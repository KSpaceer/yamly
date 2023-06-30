package token

const (
	YAMLDirective = "YAML"
	TagDirective  = "TAG"
)

type Character = byte

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
)

const byteOrderMark = 0xFEFF

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
	var result bool
	switch cst {
	case DecimalCharSetType:
		result = isDecimal(t)
	case WordCharSetType:
		result = isWord(t)
	case URICharSetType:
		result = isURI(t)
	case TagCharSetType:
		result = isTagString(t)
	case AnchorCharSetType:
		result = isAnchorString(t)
	case PlainSafeCharSetType:
		result = isPlainSafeString(t)
	case SingleQuotedCharSetType:
		result = isSingleQuotedString(t)
	case DoubleQuotedCharSetType:
		result = isDoubleQuotedString(t)
	}
	t.conformationMap.Set(cst, result)
	return false
}

type Position struct {
	Row    int
	Column int
}
