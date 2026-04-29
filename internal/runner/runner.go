//go:build windows

package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout    = 30 * time.Second
	defaultModuleName = "PSAppDeployToolkit"
)

// OutputLineCallback is called for every stdout/stderr line that PowerShell emits
// outside of the JSON response markers. Useful for streaming PSADT logs to an
// RMM console or file in real time.
type OutputLineCallback func(line string)

// Config holds configuration for the PowerShell runner.
type Config struct {
	// PSPath is the path to the PowerShell executable.
	// Defaults to auto-detection of powershell.exe or pwsh.exe.
	PSPath string

	// Timeout is the default timeout for each command execution.
	Timeout time.Duration

	// UsePowerShell7 forces use of pwsh.exe instead of powershell.exe.
	UsePowerShell7 bool

	// OnOutput is called synchronously for each stdout/stderr line emitted
	// by PowerShell outside of JSON response markers. Set this to stream
	// PSADT log output to the caller in real time during long operations.
	OnOutput OutputLineCallback
}

// Runner manages a persistent PowerShell process.
type Runner struct {
	cmd           *exec.Cmd
	stdin         io.Writer
	stdoutScanner *bufio.Scanner
	stderr        io.Reader
	mu            sync.Mutex
	running       bool
	timeout       time.Duration
	psPath        string

	// liveOutputCh receives every non-marker stdout/stderr line in real time.
	// Callers can read from this channel to stream PSADT logs during long operations.
	// The channel is closed when the runner stops.
	liveOutputCh chan string

	// onOutput is an optional synchronous callback for each output line.
	onOutput func(line string)
}

// New creates and starts a new PowerShell runner.
func New(cfg Config) (*Runner, error) {
	psPath := cfg.PSPath
	if psPath == "" {
		psPath = detectPowerShell(cfg.UsePowerShell7)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	r := &Runner{
		timeout:      timeout,
		psPath:       psPath,
		onOutput:     cfg.OnOutput,
		liveOutputCh: make(chan string, 256),
	}

	if err := r.start(); err != nil {
		return nil, err
	}

	return r, nil
}

// start launches the PowerShell process.
func (r *Runner) start() error {
	r.cmd = exec.Command(r.psPath,
		"-NoProfile",
		"-NoLogo",
		"-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-OutputFormat", "Text",
		"-Command", "-",
	)

	// Set up encoding for UTF-8
	r.cmd.Env = append(r.cmd.Environ(),
		"POWERSHELL_TELEMETRY_OPTOUT=1",
	)

	stdinPipe, err := r.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	r.stdin = stdinPipe

	stdoutPipe, err := r.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := r.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	r.stderr = stderrPipe

	// Create scanner with large buffer for JSON responses
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max
	r.stdoutScanner = scanner

	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start PowerShell process: %w", err)
	}

	r.running = true

	// Start background stderr reader that feeds into liveOutputCh
	go r.drainStderr()

	// Set UTF-8 output encoding
	_, err = fmt.Fprintln(r.stdin, "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8")
	if err != nil {
		r.Stop()
		return fmt.Errorf("failed to set UTF-8 encoding: %w", err)
	}

	// Set error action preference
	_, err = fmt.Fprintln(r.stdin, "$ErrorActionPreference = 'Stop'")
	if err != nil {
		r.Stop()
		return fmt.Errorf("failed to set error action: %w", err)
	}

	return nil
}

// drainStderr reads stderr continuously and forwards lines to liveOutputCh.
func (r *Runner) drainStderr() {
	if r.stderr == nil {
		return
	}
	scanner := bufio.NewScanner(r.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case r.liveOutputCh <- line:
		default:
			// drop if channel full to avoid blocking the runner
		}
		if r.onOutput != nil {
			r.onOutput(line)
		}
	}
}

// Stop gracefully stops the PowerShell process.
func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil
	}

	r.running = false

	// Close the live output channel
	close(r.liveOutputCh)

	// Send exit command
	if r.stdin != nil {
		fmt.Fprintln(r.stdin, "exit")
	}

	// Wait for process to finish
	if r.cmd != nil && r.cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- r.cmd.Wait()
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			r.cmd.Process.Kill()
		}
	}

	return nil
}

// IsAlive checks if the PowerShell process is still running.
func (r *Runner) IsAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// LiveOutput returns a channel that receives every stdout/stderr line
// emitted by PowerShell outside of JSON response markers. The channel is
// closed when the runner stops. Read from this channel to stream PSADT
// log output in real time during long operations.
func (r *Runner) LiveOutput() <-chan string {
	return r.liveOutputCh
}

// Heartbeat sends a simple command to verify the process is responsive.
func (r *Runner) Heartbeat(ctx context.Context) error {
	data, err := r.Execute(ctx, HeartbeatCommand())
	if err != nil {
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	result := strings.TrimSpace(string(data))
	if !strings.Contains(result, "true") && !strings.Contains(result, "True") {
		return fmt.Errorf("unexpected heartbeat response: %s", result)
	}

	return nil
}

// ImportModule imports the PSADT module into the PowerShell session.
func (r *Runner) ImportModule(ctx context.Context, moduleName string) error {
	if moduleName == "" {
		moduleName = defaultModuleName
	}

	cmd := ImportModuleCommand(moduleName)
	_, err := r.ExecuteVoid(ctx, cmd)
	return err
}

// CheckModuleVersion verifies the PSADT module version meets the minimum requirement.
func (r *Runner) CheckModuleVersion(ctx context.Context, moduleName, minVersion string) (string, error) {
	if moduleName == "" {
		moduleName = defaultModuleName
	}

	cmd := CheckModuleVersionCommand(moduleName, minVersion)
	data, err := r.Execute(ctx, cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// detectPowerShell finds the PowerShell executable.
func detectPowerShell(preferPS7 bool) string {
	if preferPS7 {
		// Try pwsh.exe first
		if path, err := exec.LookPath("pwsh.exe"); err == nil {
			return path
		}
	}

	// Try Windows PowerShell
	if path, err := exec.LookPath("powershell.exe"); err == nil {
		return path
	}

	// Fallback to pwsh
	if path, err := exec.LookPath("pwsh.exe"); err == nil {
		return path
	}

	// Default
	return "powershell.exe"
}
