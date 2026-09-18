//go:build windows

package runner

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Integration tests run against a real powershell.exe. They are skipped with
// `go test -short` (CI environments without Windows may still run unit tests).

func TestIntegration_ExecuteRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r, err := New(Config{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	ctx := context.Background()

	data, err := r.Execute(ctx, "1 + 1")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	// Runner-level Execute returns the raw {Success,Data,Error} envelope.
	if !strings.Contains(string(data), "\"Data\":2") {
		t.Errorf("1 + 1 envelope = %q, want Data=2", string(data))
	}

	if err := r.Heartbeat(ctx); err != nil {
		t.Errorf("heartbeat failed: %v", err)
	}

	// Sequential commands must each see fresh state (persistent process).
	// NOTE: the wrapper executes commands inside `& { ... }` script blocks,
	// where plain assignments are block-local. Use the global scope to test
	// state persistence across commands.
	if _, err := r.Execute(ctx, "$global:__go_padt_test_var = 42"); err != nil {
		t.Fatalf("set var failed: %v", err)
	}
	data, err = r.Execute(ctx, "$global:__go_padt_test_var + 1")
	if err != nil {
		t.Fatalf("get var failed: %v", err)
	}
	if !strings.Contains(string(data), "\"Data\":43") {
		t.Errorf("$var + 1 envelope = %q, want Data=43 (state must persist)", string(data))
	}

	if !r.IsAlive() {
		t.Error("runner should still be alive")
	}
}

func TestIntegration_ExecuteBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r, err := New(Config{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	data, err := r.ExecuteBatch(context.Background(), []string{
		"$a = 2",
		"$b = 3",
		"$a * $b",
	})
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}
	if !strings.Contains(string(data), "\"Data\":6") {
		t.Errorf("batch envelope = %q, want Data=6", string(data))
	}
}

func TestIntegration_ErrorEnvelope(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r, err := New(Config{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	data, err := r.Execute(context.Background(), "throw 'boom-go-padt'")
	if err != nil {
		t.Fatalf("transport error not expected; envelope should carry the PSADT error: %v", err)
	}
	if !strings.Contains(string(data), "\"Success\":false") {
		t.Errorf("envelope = %q, want Success=false", string(data))
	}
	if !strings.Contains(string(data), "boom-go-padt") {
		t.Errorf("envelope = %q, want the PSADT error message", string(data))
	}

	// The runner must remain usable after a failed command.
	if !r.IsAlive() {
		t.Fatal("runner should survive a command error")
	}
	if err := r.Heartbeat(context.Background()); err != nil {
		t.Errorf("heartbeat after error failed: %v", err)
	}
}

func TestIntegration_TimeoutResync(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r, err := New(Config{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	ctx := context.Background()

	// First command hangs longer than its 2s deadline.
	ctx1, cancel1 := context.WithTimeout(ctx, 2*time.Second)
	if _, err := r.Execute(ctx1, "Start-Sleep -Seconds 5; 'slow-done'"); err == nil {
		cancel1()
		t.Fatal("expected timeout error")
	}
	cancel1()

	// The next command must not be answered with the leftover response of
	// the first one (desync protection). It uses the runner's 30s default
	// timeout because the previous command is still occupying PowerShell.
	data, err := r.Execute(ctx, "'fresh-response'")
	if err != nil {
		t.Fatalf("post-timeout command failed: %v", err)
	}
	if !strings.Contains(string(data), "fresh-response") {
		t.Errorf("post-timeout response = %q, want fresh-response (protocol resync failed)", string(data))
	}
}

func TestIntegration_OnCommandMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	calls := 0
	var gotErr error
	r, err := New(Config{
		Timeout: 30 * time.Second,
		OnCommand: func(cmd string, d time.Duration, err error) {
			calls++
			gotErr = err
		},
	})
	if err != nil {
		t.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	if _, err := r.Execute(context.Background(), "$true"); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if calls != 1 {
		t.Errorf("OnCommand called %d times, want 1", calls)
	}
	if gotErr != nil {
		t.Errorf("OnCommand error should be nil, got %v", gotErr)
	}
}

func TestIntegration_SetOnCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r, err := New(Config{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	calls := make(chan int, 4)
	r.SetOnCommand(func(_ string, _ time.Duration, _ error) { calls <- 1 })
	if _, err := r.Execute(context.Background(), "$true"); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	select {
	case <-calls:
	case <-time.After(5 * time.Second):
		t.Error("OnCommand set after start was not invoked")
	}
}
