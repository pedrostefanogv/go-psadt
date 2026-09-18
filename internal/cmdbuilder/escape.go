//go:build windows

package cmdbuilder

import (
	"fmt"
	"reflect"
	"strings"
)

// EscapeString escapes a string for safe use in PowerShell commands.
// It wraps the string in single quotes and escapes embedded single quotes.
func EscapeString(s string) string {
	if s == "" {
		return "''"
	}

	// If the string looks like a PS variable, number, or boolean, don't quote it
	if isLiteral(s) {
		return s
	}

	// Escape single quotes by doubling them, then wrap in single quotes
	escaped := strings.ReplaceAll(s, "'", "''")
	return fmt.Sprintf("'%s'", escaped)
}

// isLiteral checks if a string is a PowerShell literal that should not be quoted.
// SECURITY: strings starting with "$" are only treated as literals when they are
// simple variable references (letters, digits, "_", ":", "."). Anything that
// contains sub-expressions, separators or whitespace — e.g. "$(Remove-Item ...)",
// "${...}", "$a; b" — is quoted as a plain string, preventing code injection
// through user-controlled values (app names, messages, registry paths, etc.).
func isLiteral(s string) bool {
	// PS booleans
	if s == "$true" || s == "$false" || s == "$null" {
		return true
	}
	// Simple PS variable references only
	if strings.HasPrefix(s, "$") && isSimplePSVariable(s) {
		return true
	}
	// Pure numeric
	if isNumeric(s) {
		return true
	}
	return false
}

// isSimplePSVariable reports whether s matches $[A-Za-z_][A-Za-z0-9_.:]*.
func isSimplePSVariable(s string) bool {
	if len(s) < 2 {
		return false
	}
	c := s[1]
	if !(c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
		return false
	}
	for i := 2; i < len(s); i++ {
		c = s[i]
		if !(c == '_' || c == '.' || c == ':' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// isNumeric checks if a string is a valid number.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
	}
	if start >= len(s) {
		return false
	}
	hasDot := false
	for i := start; i < len(s); i++ {
		if s[i] == '.' && !hasDot {
			hasDot = true
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// EscapeArray formats a Go string slice as a PowerShell array expression.
func EscapeArray(items []string) string {
	if len(items) == 0 {
		return "@()"
	}
	escaped := make([]string, len(items))
	for i, s := range items {
		escaped[i] = EscapeString(s)
	}
	if len(escaped) == 1 {
		return escaped[0]
	}
	return fmt.Sprintf("@(%s)", strings.Join(escaped, ", "))
}

// FormatHashtable formats a struct as a PowerShell hashtable using `ps` tags or JSON tags.
func FormatHashtable(v reflect.Value) string {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	t := v.Type()
	pairs := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Get the key name from ps tag, then json tag, then field name
		key := ""
		if tag := field.Tag.Get("ps"); tag != "" && tag != "-" {
			key, _ = parseTag(tag)
		} else if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			key = strings.Split(tag, ",")[0]
		} else {
			key = field.Name
		}

		if key == "" {
			continue
		}

		if isZero(fieldVal) {
			continue
		}

		switch fieldVal.Kind() {
		case reflect.String:
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, EscapeString(fieldVal.String())))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			pairs = append(pairs, fmt.Sprintf("%s=%d", key, fieldVal.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			pairs = append(pairs, fmt.Sprintf("%s=%d", key, fieldVal.Uint()))
		case reflect.Float32, reflect.Float64:
			pairs = append(pairs, fmt.Sprintf("%s=%g", key, fieldVal.Float()))
		case reflect.Bool:
			if fieldVal.Bool() {
				pairs = append(pairs, fmt.Sprintf("%s=$true", key))
			}
		default:
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, EscapeString(fmt.Sprintf("%v", fieldVal.Interface()))))
		}
	}

	if len(pairs) == 0 {
		return ""
	}

	return fmt.Sprintf("@{%s}", strings.Join(pairs, "; "))
}

// FormatScriptBlock wraps a string as a PowerShell script block.
func FormatScriptBlock(script string) string {
	if script == "" {
		return ""
	}
	return fmt.Sprintf("{%s}", script)
}

// FormatHashtableGeneric formats a map[string]string as a PowerShell hashtable literal.
// Used for passthrough scenarios like Set-ADTIniSection -Values.
func FormatHashtableGeneric(values map[string]string) string {
	if len(values) == 0 {
		return "@{}"
	}
	pairs := make([]string, 0, len(values))
	for k, v := range values {
		pairs = append(pairs, fmt.Sprintf("%s=%s", EscapeString(k), EscapeString(v)))
	}
	return fmt.Sprintf("@{%s}", strings.Join(pairs, "; "))
}
