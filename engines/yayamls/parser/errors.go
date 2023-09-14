package parser

import (
	"fmt"
	"github.com/KSpaceer/yamly/engines/yayamls/token"
)

type ParenthesesType int8

const (
	UnknownParenthesesType ParenthesesType = iota
	CurlyParenthesesType
	SquareParenthesesType
)

func (pt ParenthesesType) String() string {
	var s string
	switch pt {
	case UnknownParenthesesType:
		s = "unknown"
	case CurlyParenthesesType:
		s = "}"
	case SquareParenthesesType:
		s = "]"
	}
	return s
}

func tokenTypeToParenthesesType(t token.Type) ParenthesesType {
	switch t {
	case token.SequenceStartType:
		return SquareParenthesesType
	case token.MappingStartType:
		return CurlyParenthesesType
	default:
		return UnknownParenthesesType
	}
}

type UnbalancedClosingParenthesisError struct {
	Tok token.Token
}

func (u UnbalancedClosingParenthesisError) Error() string {
	return fmt.Sprintf("parenthesis %q at position %s does not have preceding opening equivalent",
		u.Tok.Origin, u.Tok.Start)
}

type UnbalancedOpeningParenthesisError struct {
	Type        ParenthesesType
	ExpectedPos token.Position
}

func (u UnbalancedOpeningParenthesisError) Error() string {
	return fmt.Sprintf("parentheses are not balanced: expected to have parentheses %q at position %s",
		u.Type, u.ExpectedPos)
}

type QuoteType int8

const (
	UnknownQuoteType QuoteType = iota
	SingleQuoteType
	DoubleQuoteType
)

func (qt QuoteType) String() string {
	var s string
	switch qt {
	case UnknownQuoteType:
		s = "unknown"
	case SingleQuoteType:
		s = "single (`'`)"
	case DoubleQuoteType:
		s = "double (`\"`)"
	}
	return s
}

type UnbalancedQuotesError struct {
	Type        QuoteType
	ExpectedPos token.Position
}

func (u UnbalancedQuotesError) Error() string {
	return fmt.Sprintf("quotes are not balanced: expected to have quote %q at position %s",
		u.Type, u.ExpectedPos)
}

type TagError struct {
	Src string
	Pos token.Position
}

func (t TagError) Error() string {
	return fmt.Sprintf("cannot use string %q at position %s as tag",
		t.Src, t.Pos)
}

type DeadEndError struct {
	Pos token.Position
}

func (d DeadEndError) Error() string {
	return fmt.Sprintf("failed to parse data: meeting a 'dead end' token at position %s", d.Pos)
}
