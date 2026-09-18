//go:build windows

package psadt

import (
	"context"
	"io"
	"log/slog"
	"time"
	"strings"
	"testing"

	"github.com/pedrostefanogv/go-psadt/types"
)

// TestOpenSessionValidation ensures invalid session configs fail fast in Go
// (before any PowerShell round-trip).
func TestOpenSessionValidation(t *testing.T) {
	cases := []struct {
		name    string
		cfg     types.SessionConfig
		wantSub string
	}{
		{
			name:    "missing app name",
			cfg:     types.SessionConfig{DeploymentType: types.DeployInstall},
			wantSub: "AppName is required",
		},
		{
			name:    "bogus deployment type",
			cfg:     types.SessionConfig{AppName: "x", DeploymentType: types.DeploymentType("Upgrade")},
			wantSub: "DeploymentType",
		},
		{
			name:    "bogus deploy mode",
			cfg:     types.SessionConfig{AppName: "x", DeploymentType: types.DeployInstall, DeployMode: types.DeployMode("Noisy")},
			wantSub: "DeployMode",
		},
	}

	client := &Client{logger: discardLogger()}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.OpenSessionWithContext(testContext(), tc.cfg)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("error %q should mention %q", err.Error(), tc.wantSub)
			}
		})
	}
}
// discardLogger returns a logger that writes nothing (test helper).
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// testContext returns a short-lived context for validation-only tests.
func testContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() { <-ctx.Done(); cancel() }()
	return ctx
}

