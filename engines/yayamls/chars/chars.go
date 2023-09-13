package chars

import (
	"strings"
)

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

func IsDecimal(s string) bool {
	for _, c := range s {
		if !IsDigit(c) {
			return false
		}
	}
	return true
}

// YAML specification: [39] ns-word-char
func IsWord(s string) bool {
	for _, c := range s {
		if !(IsDigit(c) || IsASCIILetter(c) || c == '-') {
			return false
		}
	}
	return true
}

// YAML specification: [39] ns-uri-char
func IsURI(s string) bool {
	const URIOnlyChars = "-#;/?:@&=+$_.!~*'"

	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n; i++ {
		switch runes[i] {
		case '%':
			i++
			if n-i < 2 || !AreHexDigits(runes[i], runes[i+1]) {
				return false
			}
			i++
		default:
			r := runes[i]
			if !(IsDigit(r) || IsASCIILetter(r) || IsFlowIndicatorChar(r) || strings.IndexRune(URIOnlyChars, r) != -1) {
				return false
			}
		}
	}
	return true
}

// YAML specification: [40] ns-tag-char
func IsTagString(s string) bool {
	const TagOnlyChars = "#;/?:@&=+$_.~*'"

	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n; i++ {
		switch runes[i] {
		case '%':
			i++
			if n-i < 2 || !AreHexDigits(runes[i], runes[i+1]) {
				return false
			}
			i++
		default:
			r := runes[i]
			if !(IsDigit(r) || IsASCIILetter(r) || strings.IndexRune(TagOnlyChars, r) != -1) {
				return false
			}
		}
	}
	return true
}

// YAML specification: [102] ns-anchor-char
func IsAnchorString(s string) bool {
	// has same definition with plain safe chars
	return IsPlainSafeString(s)
}

// YAML specification: [129] ns-plain-safe-in
func IsPlainSafeString(s string) bool {
	for _, c := range s {
		if IsFlowIndicatorChar(c) {
			return false
		}
	}
	return true
}

// YAML specification: [118] nb-single-char
func IsSingleQuotedString(s string) bool {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n; i++ {
		switch runes[i] {
		case '\'':
			i++
			if i == n || runes[i] != '\'' {
				return false
			}
		default:
			if !IsJSONChar(runes[i]) {
				return false
			}
		}
	}
	return true
}

// YAML specification: [107] nb-double-char
func IsDoubleQuotedString(s string) bool {
	runes := []rune(s)
	n := len(runes)
	var (
		i  int
		ok bool
	)
	for i = 0; i < n; {
		switch runes[i] {
		case '\\':
			i++
			i, ok = IsEscapedCharacter(runes, i)
			if !ok {
				return false
			}
		case '"':
			return false
		default:
			if !IsJSONChar(runes[i]) {
				return false
			}
			i++
		}
	}
	return true
}

// YAML specification: [23] c-flow-indicator
func IsFlowIndicatorChar(r rune) bool {
	switch r {
	case ',', '[', ']', '{', '}':
		return true
	default:
		return false
	}
}

func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsHexDigit(r rune) bool {
	return IsDigit(r) || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

func AreHexDigits(runes ...rune) bool {
	for i := range runes {
		if !IsHexDigit(runes[i]) {
			return false
		}
	}
	return true
}

func IsASCIILetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// YAML specification: [2] nb-json
func IsJSONChar(r rune) bool {
	return r == 0x09 || (r >= 0x20 && r <= 0x10FFFF)
}

// YAML specification: [41-58] ns-esc-...
const singleEscapedCharacters = "0at\tnvfre \"/\\N_LP"

func IsEscapedCharacter(runes []rune, i int) (int, bool) {
	if i == len(runes) {
		return 0, false
	}
	switch runes[i] {
	// YAML specification: [59] ns-esc-8-bit
	case 'x':
		i++
		if len(runes)-i < 2 || !AreHexDigits(runes[i], runes[i+1]) {
			return 0, false
		}
		i += 2
	// YAML specification: [60] ns-esc-16-bit
	case 'u':
		i++
		if len(runes)-i < 4 || !AreHexDigits(runes[i:i+4]...) {
			return 0, false
		}
		i += 4
	// YAML specification: [61] ns-esc-32-bit
	case 'U':
		i++
		if len(runes)-i < 8 || !AreHexDigits(runes[i:i+8]...) {
			return 0, false
		}
		i += 8
	default:
		if strings.IndexRune(singleEscapedCharacters, runes[i]) == -1 {
			return 0, false
		}
		i++
	}
	return i, true
}

func IsWhitespaceChar(r rune) bool {
	return r == SpaceCharacter || r == TabCharacter
}

func IsLineBreakChar(r rune) bool {
	return r == LineFeedCharacter || r == CarriageReturnCharacter
}

// IsWhitespaceChar OR IsLineBreakChar
func IsBlankChar(r rune) bool {
	return IsWhitespaceChar(r) || IsLineBreakChar(r)
}
