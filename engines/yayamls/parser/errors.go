package parser

import (
	"fmt"

	"github.com/KSpaceer/yamly/engines/yayamls/token"
)

type parenthesesType int8

const (
	unknownParenthesesType parenthesesType = iota
	curlyParenthesesType
	squareParenthesesType
)

func (pt parenthesesType) String() string {
	var s string
	switch pt {
	case unknownParenthesesType:
		s = "unknown"
	case curlyParenthesesType:
		s = "}"
	case squareParenthesesType:
		s = "]"
	}
	return s
}

func tokenTypeToParenthesesType(t token.Type) parenthesesType {
	switch t {
	case token.SequenceStartType:
		return squareParenthesesType
	case token.MappingStartType:
		return curlyParenthesesType
	default:
		return unknownParenthesesType
	}
}

// UnbalancedClosingParenthesisError is used to indicate case when a closing parenthesis (or bracket) appears
// in source without corresponding opening counterpart. E.g. "[]]"
type UnbalancedClosingParenthesisError struct {
	Tok token.Token
}

func (u UnbalancedClosingParenthesisError) Error() string {
	return fmt.Sprintf("parenthesis %q at position %s does not have preceding opening equivalent",
		u.Tok.Origin, u.Tok.Start)
}

// UnbalancedOpeningParenthesisError is used to indicate case when an opening parenthesis (or bracket) appears
// in source without corresponding closing counterpart. E.g. "[[]"
type UnbalancedOpeningParenthesisError struct {
	ptype       parenthesesType
	ExpectedPos token.Position
}

func (u UnbalancedOpeningParenthesisError) Error() string {
	return fmt.Sprintf("parentheses are not balanced: expected to have parentheses %q at position %s",
		u.ptype, u.ExpectedPos)
}

type quoteType int8

const (
	unknownQuoteType quoteType = iota
	singleQuoteType
	doubleQuoteType
)

func (qt quoteType) String() string {
	var s string
	switch qt {
	case unknownQuoteType:
		s = "unknown"
	case singleQuoteType:
		s = "single (`'`)"
	case doubleQuoteType:
		s = "double (`\"`)"
	}
	return s
}

// UnbalancedQuotesError is used to indicate case when an opening quote appears
// in source without corresponding closing counterpart. E.g. `'` or `"text" "`
type UnbalancedQuotesError struct {
	qtype       quoteType
	ExpectedPos token.Position
}

func (u UnbalancedQuotesError) Error() string {
	return fmt.Sprintf("quotes are not balanced: expected to have quote %q at position %s",
		u.qtype, u.ExpectedPos)
}

// TagError indicates case when tag has a string after it which is not allowed in YAML.
type TagError struct {
	Src string
	Pos token.Position
}

func (t TagError) Error() string {
	return fmt.Sprintf("cannot use string %q at position %s as tag",
		t.Src, t.Pos)
}

// DeadEndError is used to indicate some sort of loops occured during parsing
// when the same token appears multiple times. When YAML document is correct,
// there will be no 'dead ends' because parsing will go lightly.
type DeadEndError struct {
	Pos token.Position
}

func (d DeadEndError) Error() string {
	return fmt.Sprintf("failed to parse data: meeting a 'dead end' token at position %s", d.Pos)
}
