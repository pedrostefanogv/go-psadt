//go:build windows

package psadt

import (
	"context"
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// TestModuleInitialized checks if the PSADT module has been initialized.
func (s *Session) TestModuleInitialized() (bool, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTModuleInitialized")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// GetSession retrieves read-only properties of the current ADT session.
func (s *Session) GetSession() (*types.SessionProperties, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTSession")
	if err != nil {
		return nil, err
	}
	var result types.SessionProperties
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPowerShellProcessPath retrieves the path and version of the PowerShell
// process used by the runner.
func (s *Session) GetPowerShellProcessPath() (*types.PowerShellProcessPathInfo, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTPowerShellProcessPath")
	if err != nil {
		return nil, err
	}
	var result types.PowerShellProcessPathInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// NewLogFileName generates a standardized PSADT log file name.
func (s *Session) NewLogFileName(opts ...types.LogFileNameOptions) (string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	var cmd string
	if len(opts) > 0 {
		cmd = cmdbuilder.Build("New-ADTLogFileName", opts[0])
	} else {
		cmd = "New-ADTLogFileName"
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// AddModuleCallback registers a callback script for a specific PSADT hook point.
func (s *Session) AddModuleCallback(opts types.ModuleCallbackOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Add-ADTModuleCallback", opts)
	return s.executeVoid(ctx, cmd)
}

// GetModuleCallback retrieves registered module callbacks, optionally filtered
// by hook point.
func (s *Session) GetModuleCallback(hookPoint ...string) ([]types.CallbackInfo, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := "Get-ADTModuleCallback"
	if len(hookPoint) > 0 && hookPoint[0] != "" {
		cmd += fmt.Sprintf(" -HookPoint %s", cmdbuilder.EscapeString(hookPoint[0]))
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result []types.CallbackInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RemoveModuleCallback removes a specific module callback by name and hook point.
func (s *Session) RemoveModuleCallback(name, hookPoint string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTModuleCallback -Name %s -HookPoint %s",
		cmdbuilder.EscapeString(name), cmdbuilder.EscapeString(hookPoint))
	return s.executeVoid(ctx, cmd)
}

// ClearModuleCallback clears all callbacks for a specific hook point.
func (s *Session) ClearModuleCallback(hookPoint string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Clear-ADTModuleCallback -HookPoint %s", cmdbuilder.EscapeString(hookPoint))
	return s.executeVoid(ctx, cmd)
}
// InitializeModule initializes the PSADT module lifecycle without opening a
// full deployment session. Useful when a Go host wants PSADT environment
// variables ($envProgramFiles, etc.) available for raw scripts before
// Open-ADTSession. The SessionConfig fields DeploymentType/DeployMode and
// the Log* fields are forwarded.
func (c *Client) InitializeModule(cfg types.SessionConfig) error {
	ctx, cancel := c.defaultContext()
	defer cancel()
	return c.InitializeModuleWithContext(ctx, cfg)
}

// InitializeModuleWithContext initializes the module with an explicit context.
func (c *Client) InitializeModuleWithContext(ctx context.Context, cfg types.SessionConfig) error {
	rr, err := c.ensureAlive(ctx)
	if err != nil {
		return err
	}
	cmd := cmdbuilder.Build("Initialize-ADTModule", cfg)
	_, err = rr.ExecuteVoid(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to initialize PSADT module: %w", err)
	}
	return nil
}

// InitializeModuleIfUninitialized initializes the module only when it is not
// already initialized in the PowerShell process (idempotent).
func (c *Client) InitializeModuleIfUninitialized(cfg types.SessionConfig) error {
	ctx, cancel := c.defaultContext()
	defer cancel()
	return c.InitializeModuleIfUninitializedWithContext(ctx, cfg)
}

// InitializeModuleIfUninitializedWithContext initializes the module with an
// explicit context, only when needed.
func (c *Client) InitializeModuleIfUninitializedWithContext(ctx context.Context, cfg types.SessionConfig) error {
	rr, err := c.ensureAlive(ctx)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("Initialize-ADTModuleIfUninitialized -SessionState $ExecutionContext.SessionState")
	_, err = rr.ExecuteVoid(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to initialize PSADT module (if uninitialized): %w", err)
	}
	_ = cfg // kept for API symmetry; the idempotent variant takes no config
	return nil
}