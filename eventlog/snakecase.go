package eventlog

import (
	"unicode"
)

// camelToSnakeCase converts a string containing one or more camel
// cased words into their snake cased equivalents. For example,
// "CamelCase" results in "camel_case".
func camelToSnakeCase(s string) string {
	result := ""
	boundary := true // Are we on a word boundary?
	for _, r := range s {
		if unicode.IsUpper(r) && !boundary {
			result += "_" + string(unicode.ToLower(r))
		} else if unicode.IsUpper(r) {
			result += string(unicode.ToLower(r))
		} else {
			result += string(r)
		}
		boundary = !(unicode.IsLetter(r) || unicode.IsDigit(r))
	}
	return result
}
