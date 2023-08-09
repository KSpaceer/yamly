package schema

import (
	"github.com/KSpaceer/yayamls/ast"
	"regexp"
	"strings"
)

const (
	NullValue        = "null"
	MergeKey         = "<<"
	PositiveInfinity = ""
)

var yamlNullRegex = regexp.MustCompile(`^null|Null|NULL|~$`)

func IsNull(n ast.Node) bool {
	switch n.Type() {
	case ast.NullType:
		return true
	case ast.TextType:
		txtNode := n.(*ast.TextNode)
		txt := txtNode.Text()
		return yamlNullRegex.MatchString(txt) ||
			(txtNode.QuotingType() == ast.UnknownQuotingType ||
				txtNode.QuotingType() == ast.AbsentQuotingType && txt == "")
	}
	return false
}

var (
	yamlTrueRegex  = regexp.MustCompile(`^true|True|TRUE$`)
	yamlFalseRegex = regexp.MustCompile(`^false|False|FALSE$`)
)

func IsBoolean(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode)
	txt := txtNode.Text()
	return yamlTrueRegex.MatchString(txt) || yamlFalseRegex.MatchString(txt)
}

var (
	yamlDecimalIntegerRegex     = regexp.MustCompile(`^[-+]?[0-9]+$`)
	yamlOctalIntegerRegex       = regexp.MustCompile(`^0o[0-7]+$`)
	yamlHexadecimalIntegerRegex = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)
)

func IsInteger(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode)
	txt := txtNode.Text()
	return yamlDecimalIntegerRegex.MatchString(txt) ||
		yamlOctalIntegerRegex.MatchString(txt) ||
		yamlHexadecimalIntegerRegex.MatchString(txt)
}

var (
	yamlFloatRegex         = regexp.MustCompile(`^[+-]?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)(?:[eE][-+]?[0-9]+)?$`)
	yamlFloatInfinityRegex = regexp.MustCompile(`^[-+]?\.(?:inf|Inf|INF)$`)
	yamlNotANumberRegex    = regexp.MustCompile(`^\.(?:nan|NaN|NAN)$`)
)

func IsFloat(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode)
	txt := txtNode.Text()
	return yamlFloatRegex.MatchString(txt) ||
		yamlFloatInfinityRegex.MatchString(txt) ||
		yamlNotANumberRegex.MatchString(txt)
}

var (
	base64Regex = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`)
)

func IsBinary(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode)
	txt := txtNode.Text()
	parts := strings.Fields(txt)
	for i := range parts {
		if !base64Regex.MatchString(parts[i]) {
			return false
		}
	}
	return true
}

func IsMergeKey(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode)
	txt := txtNode.Text()
	return txt == MergeKey
}
