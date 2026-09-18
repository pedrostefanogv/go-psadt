//go:build windows

// Package psadt provides a Go wrapper for PSAppDeployToolkit v4.1.x.
//
// It allows Go applications to orchestrate Windows deployments, display UI dialogs,
// manage registry/services/filesystem and invoke installers — all through an
// idiomatic Go API with type-safety.
package psadt

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pedrostefanogv/go-psadt/internal/runner"
	"github.com/pedrostefanogv/go-psadt/types"
)

const (
	defaultModuleName  = "PSAppDeployToolkit"
	defaultMinVersion  = "4.1.0"
	defaultEnvCacheTTL = 5 * time.Minute
)

// Client is the main entry point for interacting with PSADT.
// It manages a persistent PowerShell process and module lifecycle.
type Client struct {
	runner     *runner.Runner
	logger     *slog.Logger
	moduleName string
	minVersion string
	timeout    time.Duration

	// Preserved from original options so Reconnect can restore the same config.
	psPath         string
	usePowerShell7 bool
	initTimeout    time.Duration

	envMu       sync.Mutex
	envCache    *types.EnvironmentInfo
	envCachedAt time.Time
	envCacheTTL time.Duration

	// autoPreferPS7 tries pwsh.exe before powershell.exe (fallback included).
	autoPreferPS7 bool

	// autoReconnect re-starts the PowerShell runner transparently when a
	// command finds it dead. Recommended for long-running RMM agents.
	autoReconnect bool

	// Observability
	createdAt    time.Time
	commandCount atomic.Uint64
	lastErrMu    sync.Mutex
	lastErr      error
	onCommand    func(cmd string, duration time.Duration, err error)
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	psPath         string
	moduleName     string
	minVersion     string
	timeout        time.Duration
	initTimeout    time.Duration
	logger         *slog.Logger
	usePowerShell7 bool
	envCacheTTL    time.Duration
	autoPreferPS7  bool
	autoReconnect  bool
	onCommand      func(cmd string, duration time.Duration, err error)
}

// WithPSPath sets the path to the PowerShell executable.
func WithPSPath(path string) Option {
	return func(c *clientConfig) {
		c.psPath = path
	}
}

// WithMinModuleVersion sets the minimum required PSADT module version.
func WithMinModuleVersion(version string) Option {
	return func(c *clientConfig) {
		c.minVersion = version
	}
}

// WithTimeout sets the default timeout for command execution.
func WithTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *clientConfig) {
		c.logger = logger
	}
}

// WithPowerShell7 forces use of PowerShell 7 (pwsh.exe).
func WithPowerShell7() Option {
	return func(c *clientConfig) {
		c.usePowerShell7 = true
	}
}

// WithEnvCacheTTL sets the TTL for the GetEnvironment cache.
// Set to 0 to disable caching.
func WithEnvCacheTTL(ttl time.Duration) Option {
	return func(c *clientConfig) {
		c.envCacheTTL = ttl
	}
}

// WithAutoPreferPS7 selects pwsh.exe (PowerShell 7) when available and
// falls back to powershell.exe (5.1). PSADT v4.1 works best on PS 7.
// Ignored when WithPSPath or WithPowerShell7 is also used.
func WithAutoPreferPS7() Option {
	return func(c *clientConfig) {
		c.autoPreferPS7 = true
	}
}

// WithAutoReconnect makes every session/client operation transparently
// restart the PowerShell runner (and re-import the PSADT module) when the
// process is found dead. Recommended for long-running RMM agents.
func WithAutoReconnect() Option {
	return func(c *clientConfig) {
		c.autoReconnect = true
	}
}

// OnCommand registers a per-command observability hook invoked after every
// PSADT command with the wrapped command, its duration and error (nil on
// success). The hook must be fast and must not call back into the client.
func OnCommand(hook func(cmd string, duration time.Duration, err error)) Option {
	return func(c *clientConfig) {
		c.onCommand = hook
	}
}

// WithInitTimeout sets the timeout for the initialization phase
// (Import-Module + CheckModuleVersion). Default: 2 minutes.
// Use this when Import-Module PSAppDeployToolkit may take longer
// (e.g., first run with JIT compilation, antivirus scanning).
func WithInitTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) {
		c.initTimeout = timeout
	}
}

// defaultInitTimeout is the fallback timeout for Import-Module + CheckModuleVersion
// when WithInitTimeout is not specified and the ctx has no deadline.
const defaultInitTimeout = 2 * time.Minute

// NewClient creates a new PSADT client, starting a PowerShell process,
// importing the module, and validating the version.
// Uses context.Background() for initialization — prefer NewClientWithContext
// when you need a deadline on Import-Module.
func NewClient(opts ...Option) (*Client, error) {
	return NewClientWithContext(context.Background(), opts...)
}

// NewClientWithContext creates a new PSADT client with an explicit context.
// The ctx is used for the initialization phase (Import-Module + CheckModuleVersion).
// If ctx has a deadline, Import-Module will respect it and return an error
// instead of blocking indefinitely when PowerShell is unresponsive.
//
// Use WithInitTimeout to set a dedicated timeout for initialization;
// otherwise the ctx deadline is used directly. If neither is set,
// defaultInitTimeout (2 minutes) is used as fallback.
func NewClientWithContext(ctx context.Context, opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		moduleName: defaultModuleName,
		minVersion: defaultMinVersion,
		timeout:    30 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.logger == nil {
		cfg.logger = slog.Default()
	}

	// Start PowerShell runner
	r, err := runner.New(runner.Config{
		PSPath:         cfg.psPath,
		Timeout:        cfg.timeout,
		UsePowerShell7: cfg.usePowerShell7,
		PreferPS7:      cfg.autoPreferPS7,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start PowerShell runner: %w", err)
	}

	client := &Client{
		runner:         r,
		logger:         cfg.logger,
		moduleName:     cfg.moduleName,
		minVersion:     cfg.minVersion,
		timeout:        cfg.timeout,
		psPath:         cfg.psPath,
		usePowerShell7: cfg.usePowerShell7,
		initTimeout:    cfg.initTimeout,
		envCacheTTL:    cfg.envCacheTTL,
		autoPreferPS7:  cfg.autoPreferPS7,
		autoReconnect:  cfg.autoReconnect,
		onCommand:      cfg.onCommand,
		createdAt:      time.Now(),
	}
	if client.envCacheTTL == 0 {
		client.envCacheTTL = defaultEnvCacheTTL
	}

	// The runner hook feeds client observability and forwards to the user hook.
	r.SetOnCommand(client.onRunnerCommand)

	// ── Init phase with explicit context ──
	// Priority: WithInitTimeout > ctx deadline > defaultInitTimeout (2min).
	// This guarantees Import-Module never blocks indefinitely even if the
	// caller passes a context without deadline and doesn't set WithInitTimeout.
	initCtx := ctx
	var initCancel context.CancelFunc
	if cfg.initTimeout > 0 {
		initCtx, initCancel = context.WithTimeout(ctx, cfg.initTimeout)
		defer initCancel()
	} else if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		// Fallback: ctx sem deadline e sem WithInitTimeout → usa defaultInitTimeout
		initCtx, initCancel = context.WithTimeout(ctx, defaultInitTimeout)
		defer initCancel()
	}

	// Import the PSADT module
	cfg.logger.Debug("importing PSADT module", "module", cfg.moduleName)
	if err := r.ImportModule(initCtx, cfg.moduleName); err != nil {
		r.Stop()
		return nil, fmt.Errorf("failed to import module %s: %w", cfg.moduleName, err)
	}

	// Check module version
	cfg.logger.Debug("checking module version", "minVersion", cfg.minVersion)
	version, err := r.CheckModuleVersion(initCtx, cfg.moduleName, cfg.minVersion)
	if err != nil {
		r.Stop()
		return nil, fmt.Errorf("module version check failed: %w", err)
	}
	cfg.logger.Info("PSADT module loaded", "version", version)

	return client, nil
}

// Close shuts down the PowerShell process and releases resources.
func (c *Client) Close() error {
	if c.runner != nil {
		c.logger.Debug("closing PSADT client")
		return c.runner.Stop()
	}
	return nil
}

// IsAlive checks if the underlying PowerShell process is responsive.
func (c *Client) IsAlive() bool {
	if c.runner == nil {
		return false
	}
	return c.runner.IsAlive()
}

// Reconnect attempts to restart the PowerShell runner if it has died.
// Returns the new client state after reconnection attempt.
func (c *Client) Reconnect(ctx context.Context) error {
	if c.runner != nil {
		c.runner.Stop()
	}

	cfg := runner.Config{
		PSPath:         c.psPath,
		Timeout:        c.timeout,
		UsePowerShell7: c.usePowerShell7,
	}

	r, err := runner.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to restart PowerShell runner: %w", err)
	}

	c.runner = r

	// Use initTimeout if configured, otherwise use ctx as-is with a fallback.
	// Without a fallback, Import-Module could hang indefinitely if the caller
	// passes a context.Background() or ctx without deadline.
	initCtx := ctx
	var initCancel context.CancelFunc
	if c.initTimeout > 0 {
		initCtx, initCancel = context.WithTimeout(ctx, c.initTimeout)
		defer initCancel()
	} else if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		initCtx, initCancel = context.WithTimeout(ctx, defaultInitTimeout)
		defer initCancel()
	}

	// Re-import module
	if err := r.ImportModule(initCtx, c.moduleName); err != nil {
		c.runner = nil
		return fmt.Errorf("failed to re-import module %s: %w", c.moduleName, err)
	}

	c.logger.Info("PSADT client reconnected")
	return nil
}

// Runner returns the underlying runner (useful for diagnostics).
func (c *Client) Runner() *runner.Runner {
	return c.runner
}

// Heartbeat verifies the underlying PowerShell process is responsive with a
// full command round-trip.
func (c *Client) Heartbeat(ctx context.Context) error {
	rr, err := c.ensureAlive(ctx)
	if err != nil {
		return err
	}
	return rr.Heartbeat(ctx)
}

// CommandCount returns the total number of PSADT commands executed by this
// client (including failed ones). Useful for observability in RMM agents.
func (c *Client) CommandCount() uint64 {
	return c.commandCount.Load()
}

// LastError returns the most recent error returned by a command, or nil.
func (c *Client) LastError() error {
	c.lastErrMu.Lock()
	defer c.lastErrMu.Unlock()
	return c.lastErr
}

// Uptime returns how long ago the client was created.
func (c *Client) Uptime() time.Duration {
	return time.Since(c.createdAt)
}

// ensureAlive returns a usable runner, transparently reconnecting when the
// process died and WithAutoReconnect was enabled.
func (c *Client) ensureAlive(ctx context.Context) (*runner.Runner, error) {
	if c.runner != nil && c.runner.IsAlive() {
		return c.runner, nil
	}
	if c.autoReconnect {
		if err := c.Reconnect(ctx); err != nil {
			return nil, err
		}
		return c.runner, nil
	}
	if c.runner == nil {
		return nil, fmt.Errorf("PSADT client has no runner")
	}
	return c.runner, nil
}

// onRunnerCommand is installed as the runner's OnCommand hook: it updates
// client observability counters and forwards to the user-provided hook.
func (c *Client) onRunnerCommand(cmd string, duration time.Duration, err error) {
	c.commandCount.Add(1)
	if err != nil {
		c.lastErrMu.Lock()
		c.lastErr = err
		c.lastErrMu.Unlock()
	}
	if c.onCommand != nil {
		c.onCommand(cmd, duration, err)
	}
}
// defaultContext returns a context with the client's default timeout.
func (c *Client) defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

// InvalidateEnvCache clears the cached environment data so the next call to
// GetEnvironment will fetch fresh data from the PowerShell runner.
func (c *Client) InvalidateEnvCache() {
	c.envMu.Lock()
	c.envCache = nil
	c.envCachedAt = time.Time{}
	c.envMu.Unlock()
}
