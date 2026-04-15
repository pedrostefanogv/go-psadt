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
func isLiteral(s string) bool {
	// PS booleans
	if s == "$true" || s == "$false" || s == "$null" {
		return true
	}
	// PS variables
	if strings.HasPrefix(s, "$") {
		return true
	}
	// Pure numeric
	if isNumeric(s) {
		return true
	}
	return false
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
