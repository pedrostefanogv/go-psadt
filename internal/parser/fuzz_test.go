//go:build windows

package parser

import "testing"

// FuzzParse ensures the response parser never panics on arbitrary bytes.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"{\"Success\":true,\"Data\":123,\"Error\":null}",
		"{\"Success\":false,\"Data\":null,\"Error\":{\"Message\":\"m\",\"Type\":\"T\",\"StackTrace\":\"s\"}}",
		"{}",
		"null",
		"",
		"not json at all",
		"{\"Success\":true,\"Data\":[1,2,3]}",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		resp, err := Parse(data)
		if err != nil {
			return // malformed is fine; panics are not
		}
		var target map[string]interface{}
		_ = ParseInto(resp, &target)
		_, _ = ParseBool(data)
		_, _ = ParseString(data)
		_, _ = ParseUint32(data)
		_, _ = ParseUint64(data)
		_ = CheckSuccess(data)
	})
}
