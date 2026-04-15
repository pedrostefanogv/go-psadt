//go:build windows

package cmdbuilder

import (
	"strings"
	"testing"
)

type testOpts struct {
	Name    string   `ps:"Name"`
	Path    string   `ps:"Path"`
	Recurse bool     `ps:"Recurse,switch"`
	Count   int      `ps:"Count"`
	Items   []string `ps:"Items"`
	Verbose bool     `ps:"Verbose,switch"`
}

type testProcess struct {
	Name        string `ps:"Name"`
	Description string `ps:"Description"`
}

type testWithProcess struct {
	Processes []testProcess `ps:"Processes"`
}

func TestBuild_SimpleCommand(t *testing.T) {
	cmd := Build("Get-ADTFreeDiskSpace", nil)
	if cmd != "Get-ADTFreeDiskSpace" {
		t.Errorf("expected 'Get-ADTFreeDiskSpace', got '%s'", cmd)
	}
}

func TestBuild_WithStringParams(t *testing.T) {
	opts := testOpts{
		Name: "test app",
		Path: `C:\Program Files\App`,
	}
	cmd := Build("Copy-ADTFile", opts)
	if !strings.Contains(cmd, "Copy-ADTFile") {
		t.Error("missing command name")
	}
	if !strings.Contains(cmd, "-Name 'test app'") {
		t.Errorf("missing -Name param in: %s", cmd)
	}
	if !strings.Contains(cmd, `-Path 'C:\Program Files\App'`) {
		t.Errorf("missing -Path param in: %s", cmd)
	}
}

func TestBuild_WithSwitchParams(t *testing.T) {
	opts := testOpts{
		Name:    "test",
		Recurse: true,
		Verbose: false,
	}
	cmd := Build("Remove-ADTFile", opts)
	if !strings.Contains(cmd, "-Recurse") {
		t.Errorf("missing -Recurse switch in: %s", cmd)
	}
	if strings.Contains(cmd, "-Verbose") {
		t.Error("should not include -Verbose when false")
	}
}

func TestBuild_WithIntParam(t *testing.T) {
	opts := testOpts{
		Name:  "test",
		Count: 5,
	}
	cmd := Build("Test-Command", opts)
	if !strings.Contains(cmd, "-Count 5") {
		t.Errorf("missing -Count 5 in: %s", cmd)
	}
}

func TestBuild_ZeroValuesExcluded(t *testing.T) {
	opts := testOpts{} // all zero values
	cmd := Build("Test-Command", opts)
	if cmd != "Test-Command" {
		t.Errorf("expected only command name for zero values, got: %s", cmd)
	}
}

func TestBuild_WithStringSlice(t *testing.T) {
	opts := testOpts{
		Items: []string{"one", "two", "three"},
	}
	cmd := Build("Test-Command", opts)
	if !strings.Contains(cmd, "-Items @('one', 'two', 'three')") {
		t.Errorf("unexpected slice format in: %s", cmd)
	}
}

func TestBuild_WithStructSlice(t *testing.T) {
	opts := testWithProcess{
		Processes: []testProcess{
			{Name: "notepad", Description: "Notepad"},
		},
	}
	cmd := Build("Show-ADTInstallationWelcome", opts)
	if !strings.Contains(cmd, "-Processes @{Name='notepad'; Description='Notepad'}") {
		t.Errorf("unexpected struct slice format in: %s", cmd)
	}
}

func TestEscapeString_SingleQuotes(t *testing.T) {
	result := EscapeString("it's a test")
	expected := "'it''s a test'"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestEscapeString_Empty(t *testing.T) {
	result := EscapeString("")
	if result != "''" {
		t.Errorf("expected empty quotes, got %s", result)
	}
}

func TestEscapeString_PSVariable(t *testing.T) {
	result := EscapeString("$true")
	if result != "$true" {
		t.Errorf("PS variables should not be quoted, got %s", result)
	}
}

func TestEscapeString_Number(t *testing.T) {
	result := EscapeString("42")
	if result != "42" {
		t.Errorf("numbers should not be quoted, got %s", result)
	}
}

func TestEscapeArray_Multiple(t *testing.T) {
	result := EscapeArray([]string{"a", "b", "c"})
	expected := "@('a', 'b', 'c')"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestEscapeArray_Single(t *testing.T) {
	result := EscapeArray([]string{"only"})
	expected := "'only'"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestEscapeArray_Empty(t *testing.T) {
	result := EscapeArray(nil)
	if result != "@()" {
		t.Errorf("expected @(), got %s", result)
	}
}
