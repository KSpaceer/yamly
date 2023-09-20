// Package schema contains function to derive and convert types from YAML source.
package schema

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// MergeKey is a special mapping key used to merge parent and child mappings entries.
	MergeKey = "<<"
)

// IsNull shows if string can represent a null value.
func IsNull(s string) bool {
	switch s {
	case "null", "Null", "NULL", "~", "":
		return true
	default:
		return false
	}
}

// IsBoolean shows if string can represent a boolean value.
func IsBoolean(s string) bool {
	_, ok := tryGetBoolean(s)
	return ok
}

// FromBoolean converts Go boolean into YAML boolean.
func FromBoolean(val bool) string {
	return strconv.FormatBool(val)
}

// ToBoolean tries to convert YAML text into Go boolean.
func ToBoolean(src string) (bool, error) {
	val, ok := tryGetBoolean(src)
	if !ok {
		return false, fmt.Errorf("value %q is not boolean", src)
	}
	return val, nil
}

func tryGetBoolean(s string) (v bool, isBoolean bool) {
	switch s {
	case "true", "True", "TRUE":
		return true, true
	case "false", "False", "FALSE":
		return false, true
	default:
		return false, false
	}
}

var (
	yamlDecimalIntegerRegex     = regexp.MustCompile(`^[-+]?[0-9]+$`)
	yamlDecimalUnsignedRegex    = regexp.MustCompile(`^\+?[0-9]+$`)
	yamlOctalIntegerRegex       = regexp.MustCompile(`^0o[0-7]+$`)
	yamlHexadecimalIntegerRegex = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)
)

// IsInteger shows if string can represent a signed integer value.
func IsInteger(s string) bool {
	return yamlDecimalIntegerRegex.MatchString(s) ||
		yamlOctalIntegerRegex.MatchString(s) ||
		yamlHexadecimalIntegerRegex.MatchString(s)
}

// IsUnsignedInteger shows if string can represent an unsigned integer value.
func IsUnsignedInteger(s string) bool {
	return yamlDecimalUnsignedRegex.MatchString(s) ||
		yamlOctalIntegerRegex.MatchString(s) ||
		yamlHexadecimalIntegerRegex.MatchString(s)
}

// FromInteger converts Go integer value into YAML integer.
func FromInteger(val int64) string {
	return strconv.FormatInt(val, 10)
}

// ToInteger tries to convert YAML into Go integer with given bit size.
func ToInteger(src string, bitSize int) (int64, error) {
	return strconv.ParseInt(src, 0, bitSize)
}

// FromUnsignedInteger converts Go unsigned integer value into YAML integer.
func FromUnsignedInteger(val uint64) string {
	return strconv.FormatUint(val, 10)
}

// ToUnsignedInteger tries to convert YAML into Go unsigned integer with given bit size.
func ToUnsignedInteger(src string, bitSize int) (uint64, error) {
	return strconv.ParseUint(src, 0, bitSize)
}

var (
	yamlFloatRegex         = regexp.MustCompile(`^[+-]?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)(?:[eE][-+]?[0-9]+)?$`)
	yamlFloatInfinityRegex = regexp.MustCompile(`^[-+]?\.(?:inf|Inf|INF)$`)
	yamlNotANumberRegex    = regexp.MustCompile(`^\.(?:nan|NaN|NAN)$`)
)

// IsFloat shows if string can represent a floating point number value.
func IsFloat(s string) bool {
	return yamlFloatRegex.MatchString(s) ||
		yamlFloatInfinityRegex.MatchString(s) ||
		yamlNotANumberRegex.MatchString(s)
}

// FromFloat converts Go float value into YAML float.
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

// ToFloat tries to convert YAML into Go floating point number with given bit size.
func ToFloat(src string, bitSize int) (float64, error) {
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
		return strconv.ParseFloat(src, bitSize)
	}
}

var (
	base64Regex = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`)
)

// IsBinary shows if string represents a YAML binary scalar (!!binary)
func IsBinary(s string) bool {
	parts := strings.Fields(s)
	for i := range parts {
		if !base64Regex.MatchString(parts[i]) {
			return false
		}
	}
	return true
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

// IsTimestamp shows if string represents a YAML timedate.
func IsTimestamp(s string) bool {
	for i := range timestampLayouts {
		_, err := time.Parse(timestampLayouts[i], s)
		if err == nil {
			return true
		}
	}
	return false
}

// FromTimestamp converts Go time.Time value into YAML timedate.
func FromTimestamp(val time.Time) string {
	return val.Format(time.RFC3339)
}

// ToTimestamp tries to convert YAML string into Go time.Time value.
func ToTimestamp(src string) (t time.Time, err error) {
	for i := range timestampLayouts {
		t, err = time.Parse(timestampLayouts[i], src)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}
