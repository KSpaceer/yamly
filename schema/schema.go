package schema

import (
	"fmt"
	"github.com/KSpaceer/yayamls/ast"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	MergeKey = "<<"
)

var yamlNullRegex = regexp.MustCompile(`^(?:null|Null|NULL|~|)$`)

func IsNull(n ast.Node) bool {
	switch n.Type() {
	case ast.NullType:
		return true
	case ast.TextType:
		txtNode := n.(*ast.TextNode)
		txt := txtNode.Text()
		return yamlNullRegex.MatchString(txt) &&
			(txtNode.QuotingType() == ast.UnknownQuotingType ||
				txtNode.QuotingType() == ast.AbsentQuotingType)
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

func FromBoolean(val bool) string {
	return strconv.FormatBool(val)
}

func ToBoolean(src string) (bool, error) {
	isTrue, isFalse := yamlTrueRegex.MatchString(src), yamlFalseRegex.MatchString(src)
	if !isTrue && !isFalse {
		return false, fmt.Errorf("value %q is not boolean", src)
	}
	return isTrue, nil
}

var (
	yamlDecimalIntegerRegex     = regexp.MustCompile(`^[-+]?[0-9]+$`)
	yamlDecimalUnsignedRegex    = regexp.MustCompile(`^\+?[0-9]+$`)
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

func IsUnsignedInteger(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode)
	txt := txtNode.Text()
	return yamlDecimalUnsignedRegex.MatchString(txt) ||
		yamlOctalIntegerRegex.MatchString(txt) ||
		yamlHexadecimalIntegerRegex.MatchString(txt)
}

func FromInteger(val int64) string {
	return strconv.FormatInt(val, 10)
}

func ToInteger(src string) (int64, error) {
	return strconv.ParseInt(src, 0, 64)
}

func FromUnsignedInteger(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func ToUnsignedInteger(src string) (uint64, error) {
	return strconv.ParseUint(src, 0, 64)
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

func FromFloat(val float64) string {
	switch {
	case math.IsInf(val, 1):
		return ".inf"
	case math.IsInf(val, -1):
		return "-.inf"
	case math.IsNaN(val):
		return ".nan"
	default:
		return strconv.FormatFloat(val, 'e', -1, 64)
	}
}

func ToFloat(src string) (float64, error) {
	switch {
	case yamlFloatInfinityRegex.MatchString(src):
		sign := 1
		if src[0] == '-' {
			sign = -1
		}
		return math.Inf(sign), nil
	case yamlNotANumberRegex.MatchString(src):
		return math.NaN(), nil
	default:
		return strconv.ParseFloat(src, 64)
	}
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

var (
	timestampLayouts = []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.DateOnly,
		"2006-1-2T15:4:5.999999999Z07:00", // short RFC339Nano
		"2001-1-2t15:4:5.999999999-07:00", // lower t + time zone without 'Z'
		"2001-1-2 15:4:5.999999999",       // space separated
		"2001-1-2",                        // date only short form
	}
)

func IsTimestamp(n ast.Node) bool {
	if n.Type() != ast.TextType {
		return false
	}
	txtNode := n.(*ast.TextNode)
	txt := txtNode.Text()
	for i := range timestampLayouts {
		_, err := time.Parse(timestampLayouts[i], txt)
		if err == nil {
			return true
		}
	}
	return false
}

func FromTimestamp(val time.Time) string {
	return val.Format(time.RFC3339)
}

func ToTimestamp(src string) (t time.Time, err error) {
	for i := range timestampLayouts {
		t, err = time.Parse(timestampLayouts[i], src)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}
