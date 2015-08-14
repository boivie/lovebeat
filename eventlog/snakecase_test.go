package eventlog

import (
	"testing"
)

func TestCamelToSnakeCase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"loweronly", "loweronly"},
		{"LeadingCapCamelCase", "leading_cap_camel_case"},
		{"javaStyleCamelCase", "java_style_camel_case"},
		{"TrailingCapX", "trailing_cap_x"},
		{"./$%()[]{}", "./$%()[]{}"},
		{"ÄckligRäka", "äcklig_räka"},
		{"_LeadingUnderscore", "_leading_underscore"},
		{"MultiWord CamelCase String", "multi_word camel_case string"},
		{"Test123Digits", "test123_digits"},
		{"", ""},
	}
	for _, c := range cases {
		got := camelToSnakeCase(c.in)
		if got != c.want {
			t.Errorf("CamelToSnakeCase(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
