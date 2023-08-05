package chars

type CharSetType int16

const (
	UnknownCharSetType CharSetType = 0
)

const (
	DecimalCharSetType CharSetType = 1 << iota
	WordCharSetType
	URICharSetType
	TagCharSetType
	AnchorCharSetType
	PlainSafeCharSetType
	SingleQuotedCharSetType
	DoubleQuotedCharSetType
)

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
