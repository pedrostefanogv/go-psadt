//go:build windows

package psadt

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pedrostefanogv/go-psadt/types"
)

// Client-level integration tests against the real PSAppDeployToolkit module
// (Windows only). Skipped with -short. UI functions are intentionally not
// invoked here to avoid popping real dialogs.

func TestIntegration_ClientLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	client, err := NewClientWithContext(ctx, WithTimeout(60*time.Second))
	if err != nil {
		t.Fatalf("NewClient failed (is PSAppDeployToolkit installed?): %v", err)
	}
	if !client.IsAlive() {
		t.Fatal("client should be alive after NewClient")
	}
	if client.Uptime() <= 0 {
		t.Error("uptime should be positive")
	}
	if err := client.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
	if client.IsAlive() {
		t.Error("client should not be alive after Close")
	}
}

func TestIntegration_Reconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	client, err := NewClientWithContext(ctx, WithTimeout(60*time.Second))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	// Kill the process; Reconnect must restore it (config preserved).
	if err := client.runner.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if client.IsAlive() {
		t.Fatal("client should be dead after runner.Stop")
	}

	if err := client.Reconnect(ctx); err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}
	if !client.IsAlive() {
		t.Fatal("client should be alive after Reconnect")
	}
	if err := client.Heartbeat(ctx); err != nil {
		t.Errorf("post-reconnect heartbeat failed: %v", err)
	}
}

func TestIntegration_MetricsAndEnvCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var hookCalls atomic.Uint64
	client, err := NewClientWithContext(ctx,
		WithTimeout(60*time.Second),
		WithEnvCacheTTL(50*time.Millisecond),
		OnCommand(func(_ string, _ time.Duration, err error) {
			if err != nil {
				t.Errorf("hook observed error: %v", err)
			}
			hookCalls.Add(1)
		}),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	before := client.CommandCount()

	env1, err := client.GetEnvironment()
	if err != nil {
		t.Fatalf("GetEnvironment failed: %v", err)
	}
	if env1.OS.Name == "" {
		t.Log("warning: OS name empty (unexpected for PSADT 4.1)")
	}

	// Second call within TTL must hit the cache (same pointer).
	env2, err := client.GetEnvironment()
	if err != nil {
		t.Fatalf("GetEnvironment (cached) failed: %v", err)
	}
	if env1 != env2 {
		t.Error("second GetEnvironment should return the cached snapshot")
	}

	if client.CommandCount() < before+1 {
		t.Errorf("command count should have increased: %d < %d", client.CommandCount(), before+1)
	}
	if client.LastError() != nil {
		t.Errorf("unexpected last error: %v", client.LastError())
	}
	if hookCalls.Load() == 0 {
		t.Error("OnCommand hook should have fired for the env command")
	}

	// TTL expiry forces a fresh fetch.
	time.Sleep(80 * time.Millisecond)
	env3, err := client.GetEnvironment()
	if err != nil {
		t.Fatalf("GetEnvironment after TTL failed: %v", err)
	}
	if env1 == env3 {
		t.Error("GetEnvironment after TTL expiry should fetch fresh data")
	}
}

func TestIntegration_SessionOpenClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	client, err := NewClientWithContext(ctx, WithTimeout(60*time.Second))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	session, err := client.OpenSessionWithContext(ctx,
		types.NewSessionConfig().
			App("Contoso", "go-padt-integration-test", "1.0.0").
			Silent().
			Install().
			Build(),
	)
	if err != nil {
		t.Fatalf("OpenSession failed: %v", err)
	}

	props, err := session.GetProperties()
	if err != nil {
		// Open-ADTSession internally swallows environment failures (e.g. WMI
		// access denied for Win32_ComputerSystem when not elevated) and still
		// reports success, so Get-ADTSession then finds no session. Treat that
		// as an environment limitation, not a library bug.
		t.Skipf("session not usable in this environment (Open-ADTSession internal failure): %v", err)
	}
	if props.AppName != "go-padt-integration-test" {
		t.Errorf("session AppName = %q, want go-padt-integration-test", props.AppName)
	}

	if err := session.CloseWithContext(ctx, 0); err != nil {
		t.Errorf("CloseWithContext failed: %v", err)
	}
}

func TestIntegration_ClientPool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := NewClientPool(ctx, 2, WithTimeout(60*time.Second))
	if err != nil {
		t.Fatalf("NewClientPool failed: %v", err)
	}
	defer pool.Close()

	c1, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire 1 failed: %v", err)
	}
	c2, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire 2 failed: %v", err)
	}
	if c1 == c2 {
		t.Error("Acquire must not return the same client twice while checked out")
	}
	pool.Release(c1)
	pool.Release(c2)

	c3, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire 3 failed: %v", err)
	}
	pool.Release(c3)

	if pool.Size() < 2 {
		t.Errorf("pool size = %d, want >= 2", pool.Size())
	}
}
