package token

const (
	YAMLDirective = "YAML"
	TagDirective  = "TAG"
)

type Character byte

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
	ReservedType
	LineBreakType
	SpaceType
	TabType
	BOMType
	EOFType
	DocumentEndType
	DirectiveEndType
	DotType
	StringType
	MinusType
	PlusType
)

const (
	StripChompingType = MinusType
	KeepChompingType  = PlusType
)

type CharSetType byte

const (
	UnknownCharSetType CharSetType = 0
	DecimalCharSetType CharSetType = 1 << iota
	HexadecimalCharSetTYpe
	WordCharSetType
	URICharSetType
	TagCharSetType
	AnchorCharSetType
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
	Type             Type
	Start            Position
	End              Position
	Origin           string
	conformationBits byte
}

func (t Token) ConformsCharSet(cst CharSetType) bool {
	return t.conformationBits&byte(cst) == 1
}

type Position struct {
	Offset int
	Row    int
	Column int
}
