//go:build windows

package cmdbuilder

import (
	"fmt"
	"reflect"
	"strings"
)

// Build generates a PowerShell command string from a command name and an options struct.
// The struct fields are mapped to PS parameters using the `ps` tag.
// Tag format: `ps:"ParamName"` or `ps:"ParamName,switch"` for switch parameters.
func Build(cmdName string, opts interface{}) string {
	if opts == nil {
		return cmdName
	}

	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return cmdName
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return cmdName
	}

	var parts []string
	parts = append(parts, cmdName)

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		tag := field.Tag.Get("ps")
		if tag == "" || tag == "-" {
			continue
		}

		paramName, isSwitch := parseTag(tag)
		if paramName == "" {
			continue
		}

		param := formatParam(paramName, fieldVal, isSwitch)
		if param != "" {
			parts = append(parts, param)
		}
	}

	return strings.Join(parts, " ")
}

// parseTag parses a `ps:"Name,switch"` tag and returns the param name and whether it's a switch.
func parseTag(tag string) (string, bool) {
	parts := strings.Split(tag, ",")
	name := strings.TrimSpace(parts[0])
	isSwitch := false
	for _, opt := range parts[1:] {
		if strings.TrimSpace(opt) == "switch" {
			isSwitch = true
		}
	}
	return name, isSwitch
}

// formatParam formats a single PowerShell parameter from a reflect.Value.
func formatParam(name string, v reflect.Value, isSwitch bool) string {
	if isSwitch {
		if v.Kind() == reflect.Bool && v.Bool() {
			return fmt.Sprintf("-%s", name)
		}
		return ""
	}

	if isZero(v) {
		return ""
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if s == "" {
			return ""
		}
		return fmt.Sprintf("-%s %s", name, EscapeString(s))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := v.Int()
		if val == 0 {
			return ""
		}
		return fmt.Sprintf("-%s %d", name, val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := v.Uint()
		if val == 0 {
			return ""
		}
		return fmt.Sprintf("-%s %d", name, val)

	case reflect.Float32, reflect.Float64:
		val := v.Float()
		if val == 0 {
			return ""
		}
		return fmt.Sprintf("-%s %g", name, val)

	case reflect.Bool:
		if v.Bool() {
			return fmt.Sprintf("-%s $true", name)
		}
		return ""

	case reflect.Slice:
		return formatSliceParam(name, v)

	case reflect.Map:
		return formatMapParam(name, v)

	case reflect.Interface:
		if v.IsNil() {
			return ""
		}
		return formatParam(name, v.Elem(), false)

	case reflect.Struct:
		return formatStructParam(name, v)

	default:
		return fmt.Sprintf("-%s %s", name, EscapeString(fmt.Sprintf("%v", v.Interface())))
	}
}

// formatSliceParam formats a slice parameter as a PS array.
func formatSliceParam(name string, v reflect.Value) string {
	if v.IsNil() || v.Len() == 0 {
		return ""
	}

	// Check if it's a slice of structs (e.g., []ProcessDefinition)
	elemKind := v.Type().Elem().Kind()
	if elemKind == reflect.Struct {
		items := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = FormatHashtable(v.Index(i))
		}
		if len(items) == 1 {
			return fmt.Sprintf("-%s %s", name, items[0])
		}
		return fmt.Sprintf("-%s @(%s)", name, strings.Join(items, ", "))
	}

	// Slice of strings
	if elemKind == reflect.String {
		items := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = EscapeString(v.Index(i).String())
		}
		if len(items) == 1 {
			return fmt.Sprintf("-%s %s", name, items[0])
		}
		return fmt.Sprintf("-%s @(%s)", name, strings.Join(items, ", "))
	}

	// Slice of ints
	if elemKind == reflect.Int || elemKind == reflect.Int32 || elemKind == reflect.Int64 {
		items := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = fmt.Sprintf("%d", v.Index(i).Int())
		}
		if len(items) == 1 {
			return fmt.Sprintf("-%s %s", name, items[0])
		}
		return fmt.Sprintf("-%s @(%s)", name, strings.Join(items, ", "))
	}

	return ""
}

// formatMapParam formats a map parameter as a PS hashtable.
func formatMapParam(name string, v reflect.Value) string {
	if v.IsNil() || v.Len() == 0 {
		return ""
	}

	pairs := make([]string, 0, v.Len())
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		pairs = append(pairs, fmt.Sprintf("%s=%s", EscapeString(fmt.Sprintf("%v", key.Interface())), EscapeString(fmt.Sprintf("%v", val.Interface()))))
	}

	return fmt.Sprintf("-%s @{%s}", name, strings.Join(pairs, "; "))
}

// formatStructParam formats a struct as a single PS hashtable.
func formatStructParam(name string, v reflect.Value) string {
	ht := FormatHashtable(v)
	if ht == "" {
		return ""
	}
	return fmt.Sprintf("-%s %s", name, ht)
}

// isZero checks if a reflect.Value is its type's zero value.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}
