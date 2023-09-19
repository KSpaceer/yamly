package chars

// CharSetType defines a limited set of characters used in YAML specification definitions.
type CharSetType int16

const (
	UnknownCharSetType CharSetType = 0
)

const (
	// DecimalCharSetType corresponds to [35] ns-dec-digit of YAML specification
	DecimalCharSetType CharSetType = 1 << iota
	// WordCharSetType corresponds to [38] ns-word-char of YAML specification
	WordCharSetType
	// URICharSetType corresponds to [39] ns-uri-char of YAML specification
	URICharSetType
	// TagCharSetType corresponds to [40] ns-tag-char of YAML specification
	TagCharSetType
	// AnchorCharSetType corresponds to [102] ns-anchor-char of YAML specification
	AnchorCharSetType
	// PlainSafeCharSetType corresponds to [129] ns-plain-safe-in of YAML specification
	PlainSafeCharSetType
	// SingleQuotedCharSetType corresponds to [119] ns-single-char of YAML specification
	SingleQuotedCharSetType
	// DoubleQuotedCharSetType corresponds to [108] ns-double-char of YAML specification
	DoubleQuotedCharSetType
)

// ConformsCharSet checks if given string consists of and follows the rules of provided CharSetType
func ConformsCharSet(s string, cst CharSetType) bool {
	var result bool
	switch cst {
	case DecimalCharSetType:
		result = IsDecimal(s)
	case WordCharSetType:
		result = IsWord(s)
	case URICharSetType:
		result = IsURI(s)
	case TagCharSetType:
		result = IsTagString(s)
	case AnchorCharSetType:
		result = IsAnchorString(s)
	case PlainSafeCharSetType:
		result = IsPlainSafeString(s)
	case SingleQuotedCharSetType:
		result = IsSingleQuotedString(s)
	case DoubleQuotedCharSetType:
		result = IsDoubleQuotedString(s)
	}
	return result
}
