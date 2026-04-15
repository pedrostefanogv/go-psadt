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
	"time"

	"github.com/pedrostefanogv/go-psadt/internal/runner"
)

const (
	defaultModuleName = "PSAppDeployToolkit"
	defaultMinVersion = "4.1.0"
)

// Client is the main entry point for interacting with PSADT.
// It manages a persistent PowerShell process and module lifecycle.
type Client struct {
	runner     *runner.Runner
	logger     *slog.Logger
	moduleName string
	minVersion string
	timeout    time.Duration
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	psPath         string
	moduleName     string
	minVersion     string
	timeout        time.Duration
	logger         *slog.Logger
	usePowerShell7 bool
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

// NewClient creates a new PSADT client, starting a PowerShell process,
// importing the module, and validating the version.
func NewClient(opts ...Option) (*Client, error) {
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
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start PowerShell runner: %w", err)
	}

	client := &Client{
		runner:     r,
		logger:     cfg.logger,
		moduleName: cfg.moduleName,
		minVersion: cfg.minVersion,
		timeout:    cfg.timeout,
	}

	ctx := context.Background()

	// Import the PSADT module
	cfg.logger.Debug("importing PSADT module", "module", cfg.moduleName)
	if err := r.ImportModule(ctx, cfg.moduleName); err != nil {
		r.Stop()
		return nil, fmt.Errorf("failed to import module %s: %w", cfg.moduleName, err)
	}

	// Check module version
	cfg.logger.Debug("checking module version", "minVersion", cfg.minVersion)
	version, err := r.CheckModuleVersion(ctx, cfg.moduleName, cfg.minVersion)
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

// defaultContext returns a context with the client's default timeout.
func (c *Client) defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}
