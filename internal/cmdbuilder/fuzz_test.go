//go:build windows

package cmdbuilder

import (
	"strings"
	"testing"
)

// FuzzEscapeString checks the security invariants of EscapeString:
//   - output never breaks out of a single-quoted literal unless the input is
//     a recognized literal ($true, numbers, simple $variables)
//   - sub-expression injection attempts stay quoted
func FuzzEscapeString(f *testing.F) {
	seeds := []string{
		"",
		"firefox",
		"C:\\Program Files\\App 'quoted' \"double\"",
		"$true",
		"$env:ProgramFiles",
		"$(Remove-Item -Path C:/ -Recurse)",
		"$({ evil })",
		"$a; Start-Process calc",
		"123",
		"-42.5",
		"$(Invoke-Expression $x)",
		"$" + "{complex}",
		"multiline\nvalue",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		out := EscapeString(s)

		if !strings.HasPrefix(out, "'") {
			// Unquoted output is only allowed for known-safe literals.
			if !isLiteral(s) {
				t.Errorf("EscapeString(%q) = unquoted %q but input is not a literal", s, out)
			}
			if out != s {
				t.Errorf("literal pass-through changed the value: %q -> %q", s, out)
			}
		}

		// Unquoted output must never contain separators that could turn into
		// code boundaries.
		if out != "" && out[0] != '\'' {
			for _, r := range out {
				if r == '\'' || r == ';' || r == '(' || r == ')' || r == '{' || r == '}' {
					t.Errorf("unquoted output %q contains code boundary char %q (input %q)", out, string(r), s)
				}
			}
		}
	})
}

// FuzzBuild ensures Build never panics on arbitrary values and always
// returns output that starts with the command name.
func FuzzBuild(f *testing.F) {
	f.Add("hello", 5, true, "a;b", 1.5)
	f.Add("", 0, false, "", 0.0)
	f.Add("'; drop table --", -1, false, "junk\u0000\u00ff", 1e308)

	f.Fuzz(func(t *testing.T, s string, i int, b bool, item string, fl float64) {
		_ = i
		out := Build("Test-Command", fuzzOpts{
			Title: s,
			Flag:  b,
			Items: []string{item},
			Ratio: fl,
		})
		if out != "" && !strings.HasPrefix(out, "Test-Command") {
			t.Errorf("Build output = %q, want prefix Test-Command", out)
		}
	})
}

type fuzzOpts struct {
	Title string   `ps:"Title"`
	Flag  bool     `ps:"Flag,switch"`
	Items []string `ps:"Items"`
	Ratio float64  `ps:"Ratio"`
}
