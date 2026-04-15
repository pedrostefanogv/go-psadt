//go:build windows

package runner

import (
	"bufio"
	"context"
	"strings"
	"testing"
	"time"
)

func TestWrapCommand(t *testing.T) {
	cmd := WrapCommand("Get-ADTFreeDiskSpace")
	if !strings.Contains(cmd, "Get-ADTFreeDiskSpace") {
		t.Error("wrapped command should contain the original command")
	}
	if !strings.Contains(cmd, "$result = & {") {
		t.Error("wrapped command should execute inside a script block")
	}
	if !strings.Contains(cmd, BeginMarker) {
		t.Error("wrapped command should contain begin marker")
	}
	if !strings.Contains(cmd, EndMarker) {
		t.Error("wrapped command should contain end marker")
	}
	if !strings.Contains(cmd, "ConvertTo-Json") {
		t.Error("wrapped command should convert to JSON")
	}
	if !strings.Contains(cmd, "try") {
		t.Error("wrapped command should have try/catch")
	}
}

func TestWrapCommand_MultilineCommand(t *testing.T) {
	cmd := WrapCommand(CheckModuleVersionCommand("PSAppDeployToolkit", "4.1.0"))
	if !strings.Contains(cmd, "$result = & {") {
		t.Error("multiline commands should execute inside a script block")
	}
	if !strings.Contains(cmd, "$m = Get-Module") {
		t.Error("wrapped multiline command should preserve the original script body")
	}
}

func TestWrapVoidCommand(t *testing.T) {
	cmd := WrapVoidCommand("Close-ADTInstallationProgress")
	if !strings.Contains(cmd, "Close-ADTInstallationProgress") {
		t.Error("wrapped void command should contain the original command")
	}
	if !strings.Contains(cmd, `Data = $null`) {
		t.Error("void command should set Data to $null")
	}
}

func TestImportModuleCommand(t *testing.T) {
	cmd := ImportModuleCommand("PSAppDeployToolkit")
	if !strings.Contains(cmd, "Import-Module") {
		t.Error("should contain Import-Module")
	}
	if !strings.Contains(cmd, "PSAppDeployToolkit") {
		t.Error("should contain module name")
	}
	if !strings.Contains(cmd, "-Force") {
		t.Error("should use -Force")
	}
}

func TestCheckModuleVersionCommand(t *testing.T) {
	cmd := CheckModuleVersionCommand("PSAppDeployToolkit", "4.1.0")
	if !strings.Contains(cmd, "PSAppDeployToolkit") {
		t.Error("should contain module name")
	}
	if !strings.Contains(cmd, "4.1.0") {
		t.Error("should contain version")
	}
}

func TestDetectPowerShell(t *testing.T) {
	path := detectPowerShell(false)
	if path == "" {
		t.Error("should detect a PowerShell path")
	}
	if !strings.Contains(strings.ToLower(path), "powershell") && !strings.Contains(strings.ToLower(path), "pwsh") {
		t.Errorf("unexpected path: %s", path)
	}
}

func TestHeartbeatCommand(t *testing.T) {
	cmd := HeartbeatCommand()
	if cmd != "$true" {
		t.Errorf("expected '$true', got '%s'", cmd)
	}
}

func TestReadResponse_EOF(t *testing.T) {
	r := &Runner{
		stdoutScanner: bufio.NewScanner(strings.NewReader("")),
		running:       true,
		timeout:       time.Second,
	}

	_, err := r.readResponse(context.Background())
	if err == nil {
		t.Fatal("expected EOF error, got nil")
	}
	if !strings.Contains(err.Error(), "ended before completing response") {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.running {
		t.Fatal("runner should be marked as not running after EOF")
	}
}
