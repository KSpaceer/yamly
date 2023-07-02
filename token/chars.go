package token

import "strings"

func isDecimal(t *Token) bool {
	for _, c := range t.Origin {
		if !isDigit(c) {
			return false
		}
	}
	return true
}

// YAML specification: [39] ns-word-char
func isWord(t *Token) bool {
	for _, c := range t.Origin {
		if !(isDigit(c) || isASCIILetter(c) || c == '-') {
			return false
		}
	}
	return true
}

// YAML specification: [39] ns-uri-char
func isURI(t *Token) bool {
	const URIOnlyChars = "#;/?:@&=+$_.!~*'"

	runes := []rune(t.Origin)
	n := len(runes)
	for i := 0; i < n; i++ {
		switch runes[i] {
		case '%':
			i++
			if n-i < 2 || !areHexDigits(runes[i], runes[i+1]) {
				return false
			}
			i++
		default:
			r := runes[i]
			if !(isDigit(r) || isASCIILetter(r) || isFlowIndicator(r) || strings.IndexRune(URIOnlyChars, r) != -1) {
				return false
			}
		}
	}
	return true
}

// YAML specification: [40] ns-tag-char
func isTagString(t *Token) bool {
	const TagOnlyChars = "#;/?:@&=+$_.~*'"

	runes := []rune(t.Origin)
	n := len(runes)
	for i := 0; i < n; i++ {
		switch runes[i] {
		case '%':
			i++
			if n-i < 2 || !areHexDigits(runes[i], runes[i+1]) {
				return false
			}
			i++
		default:
			r := runes[i]
			if !(isDigit(r) || isASCIILetter(r) || strings.IndexRune(TagOnlyChars, r) != -1) {
				return false
			}
		}
	}
	return true
}

// YAML specification: [102] ns-anchor-char
func isAnchorString(t *Token) bool {
	// has same definition with plain safe chars
	return isPlainSafeString(t)
}

// YAML specification: [129] ns-plain-safe-in
func isPlainSafeString(t *Token) bool {
	for _, c := range t.Origin {
		if isFlowIndicator(c) {
			return false
		}
	}
	return true
}

// YAML specification: [118] nb-single-char
func isSingleQuotedString(t *Token) bool {
	runes := []rune(t.Origin)
	n := len(runes)
	for i := 0; i < n; i++ {
		switch runes[i] {
		case '\'':
			i++
			if i == n || runes[i] != '\'' {
				return false
			}
		default:
			if !isJSONChar(runes[i]) {
				return false
			}
		}
	}
	return true
}

// YAML specification: [107] nb-double-char
func isDoubleQuotedString(t *Token) bool {
	runes := []rune(t.Origin)
	n := len(runes)
	var (
		i  int
		ok bool
	)
	for i = 0; i < n; {
		switch runes[i] {
		case '\\':
			i++
			i, ok = isEscapedCharacter(runes, i)
			if !ok {
				return false
			}
		case '"':
			return false
		default:
			if !isJSONChar(runes[i]) {
				return false
			}
			i++
		}
	}
	return true
}

// YAML specification: [23] c-flow-indicator
func isFlowIndicator(r rune) bool {
	switch r {
	case ',', '[', ']', '{', '}':
		return true
	default:
		return false
	}
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexDigit(r rune) bool {
	return isDigit(r) || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

func areHexDigits(runes ...rune) bool {
	for i := range runes {
		if !isHexDigit(runes[i]) {
			return false
		}
	}
	return true
}

func isASCIILetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// YAML specification: [2] nb-json
func isJSONChar(r rune) bool {
	return r == 0x09 || (r >= 0x20 && r <= 0x10FFFF)
}

func isEscapedCharacter(runes []rune, i int) (int, bool) {
	// YAML specification: [41-58] ns-esc-...
	const singleEscapedCharacters = "0at\tnvfre \"/\\N_LP"

	if i == len(runes) {
		return 0, false
	}
	switch runes[i] {
	// YAML specification: [59] ns-esc-8-bit
	case 'x':
		i++
		if len(runes)-i < 2 || !areHexDigits(runes[i], runes[i+1]) {
			return 0, false
		}
		i += 2
	// YAML specification: [60] ns-esc-16-bit
	case 'u':
		i++
		if len(runes)-i < 4 || !areHexDigits(runes[i:i+4]...) {
			return 0, false
		}
		i += 4
	// YAML specification: [61] ns-esc-32-bit
	case 'U':
		i++
		if len(runes)-i < 8 || !areHexDigits(runes[i:i+8]...) {
			return 0, false
		}
		i += 8
	default:
		if strings.IndexRune(singleEscapedCharacters, runes[i]) == -1 {
			return 0, false
		}
	}
	return i, true
}
