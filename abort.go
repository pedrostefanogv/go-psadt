//go:build windows

package psadt

import (
	"context"
	"os/exec"
	"strconv"
	"time"
)

// Abort force-terminates the underlying PowerShell process tree, including
// any installer or child process it has spawned. Use it when a deployment is
// stuck (e.g. an installer waiting on a UI element) and the RMM agent needs
// to cancel the task instead of waiting for the command timeout.
//
// It uses "taskkill /T /F" so the whole tree dies (PowerShell + setup.exe +
// winget children), then stops the runner gracefully. After a successful
// abort the client is no longer usable until Reconnect is called.
//
// Safe to call concurrently with other operations: it does not wait for the
// in-flight command to finish.
func (c *Client) Abort() error {
	if c == nil || c.runner == nil {
		return nil
	}
	pid := c.runner.PID()
	if pid == 0 {
		// Nothing to kill by PID; still stop the runner.
		return c.runner.Stop()
	}

	// /T kills the tree, /F forces termination.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := executeTaskKill(ctx, pid); err != nil {
		c.logger.Warn("taskkill failed during abort; falling back to graceful stop", "pid", pid, "error", err)
	}

	c.logger.Info("PSADT client aborted (process tree killed)", "pid", pid)
	return c.runner.Stop()
}

// executeTaskKill force-terminates the process tree rooted at pid.
func executeTaskKill(ctx context.Context, pid int) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "taskkill",
		"/PID", strconv.Itoa(pid), "/T", "/F")
	return cmd.CombinedOutput()
}
