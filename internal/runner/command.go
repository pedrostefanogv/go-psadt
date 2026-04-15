//go:build windows

package runner

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"
)

// Execute sends a PowerShell command and returns the JSON response bytes.
// The command is automatically wrapped with try/catch and delimiters.
func (r *Runner) Execute(ctx context.Context, psCommand string) ([]byte, error) {
	return r.executeWrapped(ctx, WrapCommand(psCommand))
}

// ExecuteVoid sends a PowerShell command that returns no data.
func (r *Runner) ExecuteVoid(ctx context.Context, psCommand string) ([]byte, error) {
	return r.executeWrapped(ctx, WrapVoidCommand(psCommand))
}

// ExecuteRaw sends a raw PowerShell command string (already wrapped) and reads the response.
func (r *Runner) executeWrapped(ctx context.Context, wrappedCmd string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil, fmt.Errorf("PowerShell runner is not running")
	}

	// Write command to stdin
	_, err := fmt.Fprintln(r.stdin, wrappedCmd)
	if err != nil {
		r.running = false
		return nil, fmt.Errorf("failed to write command to PowerShell: %w", err)
	}

	// Read response between markers
	return r.readResponse(ctx)
}

// readResponse reads stdout until it finds the begin/end markers, extracting the JSON between them.
func (r *Runner) readResponse(ctx context.Context) ([]byte, error) {
	scanner := r.stdoutScanner
	if scanner == nil {
		return nil, fmt.Errorf("stdout scanner not initialized")
	}

	// Apply timeout from context or default
	timeout := r.timeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	type scanResult struct {
		line string
		err  error
	}

	// Channel for receiving scanned lines
	lineCh := make(chan scanResult, 1)

	var jsonLines []string
	inResponse := false

	for {
		// Read next line with timeout
		go func() {
			if scanner.Scan() {
				lineCh <- scanResult{line: scanner.Text()}
			} else {
				lineCh <- scanResult{err: scanner.Err()}
			}
		}()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(timeout):
			return nil, fmt.Errorf("timeout waiting for PowerShell response after %v", timeout)
		case result := <-lineCh:
			if result.err != nil {
				r.running = false
				if result.err == bufio.ErrTooLong {
					return nil, fmt.Errorf("PowerShell response too large")
				}
				return nil, fmt.Errorf("error reading PowerShell output: %w", result.err)
			}

			line := strings.TrimSpace(result.line)

			if line == BeginMarker {
				inResponse = true
				jsonLines = nil
				continue
			}

			if line == EndMarker {
				if !inResponse {
					continue
				}
				return []byte(strings.Join(jsonLines, "\n")), nil
			}

			if inResponse {
				jsonLines = append(jsonLines, line)
			}
		}
	}
}
