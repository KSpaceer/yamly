package chars

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func ConvertFromYAMLSingleQuotedString(s string) (string, error) {
	var sb strings.Builder
	sb.Grow(len(s))
	var hasPrecedingQuote bool

	for _, r := range s {
		switch r {
		case '\'':
			if hasPrecedingQuote {
				hasPrecedingQuote = false
				sb.WriteByte('\'')
			} else {
				hasPrecedingQuote = true
			}
		default:
			if hasPrecedingQuote {
				return "", fmt.Errorf("string contains unquoted single quote")
			}
			if !IsJSONChar(r) {
				return "", fmt.Errorf("expected to have JSON char (see YAML spec [2] nb-json), but got %c",
					r)
			}
			sb.WriteRune(r)
		}
	}
	if hasPrecedingQuote {
		return "", fmt.Errorf("string contains unquoted single quote")
	}
	return sb.String(), nil
}

func ConvertToYAMLSingleQuotedString(s string) (string, error) {
	return strings.ReplaceAll(s, "'", "''"), nil
}

func ConvertFromYAMLDoubleQuotedString(s string) (string, error) {
	var sb strings.Builder
	sb.Grow(len(s))
	runes := []rune(s)
	n := len(runes)
	var (
		i   int
		err error
	)
	for i = 0; i < n; {
		switch runes[i] {
		case '\\':
			var escaped rune
			i++
			escaped, i, err = unescapeCharacter(runes, i)
			if err != nil {
				return "", err
			}
			sb.WriteRune(escaped)
		case '"':
			return "", fmt.Errorf("unexpected unescaped double quote in a double quoted string")
		default:
			if !IsJSONChar(runes[i]) {
				return "", fmt.Errorf("expected to have JSON char (see YAML spec [2] nb-json), but got %c",
					runes[i])
			}
			sb.WriteRune(runes[i])
			i++
		}
	}
	return sb.String(), nil
}

func ConvertToYAMLDoubleQuotedString(s string) (string, error) {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		if escaped, ok := escapeCharacter(r); ok {
			sb.WriteString(escaped)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

func unescapeCharacter(runes []rune, i int) (rune, int, error) {
	if i == len(runes) {
		return 0, 0, fmt.Errorf("unexpected end of string after escaping with '\\'")
	}
	switch runes[i] {
	case 'x', 'u', 'U':
		return unescapeUnicodeCharacter(runes, i)
	default:
		return unescapeASCIICharacter(runes, i)
	}
}

func escapeCharacter(r rune) (string, bool) {
	if result, ok := escapeASCIICharacter(r); ok {
		return result, ok
	}
	return escapeUnicodeCharacter(r)
}

func escapeUnicodeCharacter(r rune) (string, bool) {
	if IsJSONChar(r) {
		return "", false
	}
	escaped := strconv.QuoteRuneToASCII(r)
	return escaped[1 : len(escaped)-1], true
}

func escapeASCIICharacter(r rune) (string, bool) {
	var result string
	switch r {
	case 0:
		result = "\\0"
	case '\a':
		result = `\a`
	case '\b':
		result = `\b`
	case '\t':
		result = `\t`
	case '\n':
		result = `\n`
	case '\v':
		result = `\v`
	case '\f':
		result = `\f`
	case '\r':
		result = `\r`
	case '\x1B':
		result = `\e`
	// no need to escape space characters, because writer doesn't do multiline double quoted strings
	// case ' ':
	//     result = `\ `
	case '"':
		result = `\"`
	// no need to escape forward slashes, because it is implemented just for JSON compatibility
	// and escaping while writing reduces readability
	// case '/':
	// result = `\/`
	case '\\':
		result = `\\`
	case '\x85':
		result = `\N`
	case '\xA0':
		result = `\_`
	case 0x2028:
		result = `\L`
	case 0x2029:
		result = `\P`
	default:
		return "", false
	}
	return result, true
}

func unescapeASCIICharacter(runes []rune, i int) (rune, int, error) {
	var result rune
	switch runes[i] {
	case '0':
		result = 0
	case 'a':
		result = '\a'
	case 'b':
		result = '\b'
	case 't':
		result = '\t'
	case 'n':
		result = '\n'
	case 'v':
		result = '\v'
	case 'f':
		result = '\f'
	case 'r':
		result = '\r'
	case 'e':
		result = '\x1B' // \e
	case ' ':
		result = ' '
	case '"':
		result = '"'
	case '/':
		result = '/'
	case '\\':
		result = '\\'
	case 'N':
		result = '\x85' // \N
	case '_':
		result = '\xA0' // \_
	case 'L':
		result = 0x2028 // \L
	case 'P':
		result = 0x2029 // \P
	default:
		return 0, 0, fmt.Errorf("escaping character %q is not supported", runes[i])
	}
	i++
	return result, i, nil
}

func unescapeUnicodeCharacter(runes []rune, i int) (rune, int, error) {
	var escaped []rune
	switch runes[i] {
	case 'x':
		i++
		if len(runes)-i < 2 {
			return 0, 0, fmt.Errorf("unexpected end of string after escaping with \\x")
		}
		if !AreHexDigits(runes[i], runes[i+1]) {
			return 0, 0, fmt.Errorf("expected to have 2 hexadecimal digits after \\x, but have %s",
				string(runes[i:i+2]))
		}
		i += 2
		escaped = runes[i : i+2]
	case 'u':
		i++
		if len(runes)-i < 4 {
			return 0, 0, fmt.Errorf("unexpected end of string after escaping with \\u")
		}
		if !AreHexDigits(runes[i : i+4]...) {
			return 0, 0, fmt.Errorf("expected to have 4 hexadecimal digits after \\u, but have %s",
				string(runes[i:i+4]))
		}
		i += 4
		escaped = runes[i : i+4]
	case 'U':
		i++
		if len(runes)-i < 8 {
			return 0, 0, fmt.Errorf("unexpected end of string after escaping with \\U")
		}
		if !AreHexDigits(runes[i : i+8]...) {
			return 0, 0, fmt.Errorf("expected to have 8 hexadecimal digits after \\U, but have %s",
				string(runes[i:i+8]))
		}
		i += 8
		escaped = runes[i : i+8]
	}
	result, err := parseEscapedHexDigits(escaped)
	return result, i, err
}

func parseEscapedHexDigits(runes []rune) (rune, error) {
	result, err := strconv.ParseInt(string(runes), 16, 32)
	if err != nil {
		return 0, err
	}
	r := rune(result)
	if !utf8.ValidRune(r) {
		r = utf8.RuneError
	}
	return r, nil
}
