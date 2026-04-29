//go:build windows

package psadt

import (
	"context"
	"fmt"
	"strings"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/internal/runner"
	"github.com/pedrostefanogv/go-psadt/types"
)

// Session represents an open ADT deployment session.
// All PSADT function calls are made through a Session.
//
// Context propagation: every method uses the session's embedded context
// (if set via WithContext) or the Client's default timeout. This means
// callers can control deadlines/cancellation for an entire sequence of
// operations without passing ctx to each method individually.
type Session struct {
	client *Client
	runner *runner.Runner
	config types.SessionConfig
	ctx    context.Context // embedded context; nil means use default
	closed bool
}

// WithContext returns a shallow copy of the session that uses the given
// context for all subsequent method calls. The original session is not
// modified. This allows callers to set a deadline or cancellation for
// a group of operations without modifying every call site.
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	results, err := session.WithContext(ctx).GetApplication(opts)
func (s *Session) WithContext(ctx context.Context) *Session {
	return &Session{
		client: s.client,
		runner: s.runner,
		config: s.config,
		ctx:    ctx,
		closed: s.closed,
	}
}

// OpenSession opens a new ADT deployment session with the given configuration.
// This calls Open-ADTSession in PowerShell.
func (c *Client) OpenSession(cfg types.SessionConfig) (*Session, error) {
	ctx, cancel := c.defaultContext()
	defer cancel()
	return c.OpenSessionWithContext(ctx, cfg)
}

// OpenSessionWithContext opens a new ADT session with an explicit context.
func (c *Client) OpenSessionWithContext(ctx context.Context, cfg types.SessionConfig) (*Session, error) {
	cmd := cmdbuilder.Build("Open-ADTSession", cfg)
	c.logger.Debug("opening ADT session", "command", cmd)

	_, err := c.runner.ExecuteVoid(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to open ADT session: %w", err)
	}

	session := &Session{
		client: c,
		runner: c.runner,
		config: cfg,
	}

	c.logger.Info("ADT session opened",
		"appName", cfg.AppName,
		"deploymentType", cfg.DeploymentType,
		"deployMode", cfg.DeployMode,
	)

	return session, nil
}

// Close closes the ADT session with the specified exit code.
// This calls Close-ADTSession in PowerShell.
func (s *Session) Close(exitCode int) error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.CloseWithContext(ctx, exitCode)
}

// CloseWithContext closes the ADT session with an explicit context.
func (s *Session) CloseWithContext(ctx context.Context, exitCode int) error {
	if s.closed {
		return nil
	}

	cmd := fmt.Sprintf("Close-ADTSession -ExitCode %d", exitCode)
	s.client.logger.Debug("closing ADT session", "exitCode", exitCode)

	_, err := s.runner.ExecuteVoid(ctx, cmd)
	s.closed = true

	if err != nil {
		if isExpectedSessionCloseRunnerTermination(err) {
			s.client.logger.Info("ADT session closed and PowerShell runner exited", "exitCode", exitCode)
			return nil
		}
		return fmt.Errorf("failed to close ADT session: %w", err)
	}

	s.client.logger.Info("ADT session closed", "exitCode", exitCode)
	return nil
}

func isExpectedSessionCloseRunnerTermination(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "PowerShell process ended before completing response") ||
		strings.Contains(msg, "PowerShell runner is not running") ||
		strings.Contains(msg, "failed to write command to PowerShell")
}

// GetProperties returns the current session properties.
// This calls Get-ADTSession in PowerShell.
func (s *Session) GetProperties() (*types.SessionProperties, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.GetPropertiesWithContext(ctx)
}

// GetPropertiesWithContext returns session properties with an explicit context.
func (s *Session) GetPropertiesWithContext(ctx context.Context) (*types.SessionProperties, error) {
	cmd := "Get-ADTSession | Select-Object CurrentDate,CurrentDateTime,CurrentTime,InstallPhase,LogPath,UseDefaultMsi,DeployAppScriptFriendlyName,DeployAppScriptParameters,DeployAppScriptVersion | ConvertTo-Json -Depth 5"

	data, err := s.runner.Execute(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get session properties: %w", err)
	}

	var props types.SessionProperties
	if err := parser.ParseResponse(data, &props); err != nil {
		return nil, err
	}

	return &props, nil
}

// execute is a helper that executes a command and returns raw bytes.
func (s *Session) execute(ctx context.Context, cmd string) ([]byte, error) {
	s.client.logger.Debug("executing command", "command", cmd)
	return s.runner.Execute(ctx, cmd)
}

// executeVoid is a helper that executes a void command.
func (s *Session) executeVoid(ctx context.Context, cmd string) error {
	s.client.logger.Debug("executing void command", "command", cmd)
	data, err := s.runner.ExecuteVoid(ctx, cmd)
	if err != nil {
		return err
	}
	return parser.CheckSuccess(data)
}

// getContext returns the session's embedded context if set, otherwise the
// client's default context. This is the central context resolution point
// used by all session methods.
func (s *Session) getContext() (context.Context, context.CancelFunc) {
	if s.ctx != nil {
		return context.WithCancel(s.ctx)
	}
	return s.getContext()
}

// LiveOutput returns a channel that receives live stdout/stderr lines from
// the PowerShell process in real time. Use this to stream PSADT log output
// during long operations (e.g., installation progress).
func (s *Session) LiveOutput() <-chan string {
	return s.runner.LiveOutput()
}
