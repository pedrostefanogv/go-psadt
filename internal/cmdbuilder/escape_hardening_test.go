//go:build windows

package cmdbuilder

import (
	"reflect"
	"strings"
	"testing"
)

// TestEscapeString_BlockInjection ensures strings containing sub-expressions
// are quoted as literals instead of being interpreted as PowerShell code.
func TestEscapeString_BlockInjection(t *testing.T) {
	cases := []string{
		"$(Remove-Item -Path C:/ -Recurse)",
		"$({ evil })",
		"$a; Start-Process calc",
		"$env:PATH; calc.exe",
	}
	for _, s := range cases {
		got := EscapeString(s)
		if !strings.HasPrefix(got, "'") {
			t.Errorf("EscapeString(%q) = %q; expected a quoted literal", s, got)
		}
	}
}

// TestEscapeString_AllowsSimpleVariables ensures benign PS variable
// references and booleans are still passed unquoted.
func TestEscapeString_AllowsSimpleVariables(t *testing.T) {
	ok := map[string]string{
		"$true":             "$true",
		"$false":            "$false",
		"$null":             "$null",
		"$env:ProgramFiles": "$env:ProgramFiles",
		"$env:TEMP":         "$env:TEMP",
		"$adtSession":       "$adtSession",
		"123":               "123",
		"-42.5":             "-42.5",
	}
	for in, want := range ok {
		if got := EscapeString(in); got != want {
			t.Errorf("EscapeString(%q) = %q; want %q", in, got, want)
		}
	}
}

// TestFormatMapParam_BoolLiterals verifies boolean values inside map
// parameters are emitted as $true/$false rather than quoted strings.
func TestFormatMapParam_BoolLiterals(t *testing.T) {
	v := map[string]interface{}{
		"Enabled": true,
		"Hidden":  false,
		"Name":    "x",
	}
	got := formatMapParam("Opts", reflect.ValueOf(v))
	if !strings.Contains(got, "'Enabled'=$true") {
		t.Errorf("bool true should be $true literal, got: %s", got)
	}
	if !strings.Contains(got, "'Hidden'=$false") {
		t.Errorf("bool false should be $false literal, got: %s", got)
	}
	if !strings.Contains(got, "'Name'='x'") {
		t.Errorf("string values should remain quoted, got: %s", got)
	}
}
