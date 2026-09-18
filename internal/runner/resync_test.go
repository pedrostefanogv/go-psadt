//go:build windows

package runner

import (
	"bufio"
	"context"
	"strings"
	"testing"
	"time"
)

// TestReadResponse_ResyncAfterTimeout verifies that leftover lines from a
// previous (timed-out) response are discarded instead of being misattributed
// to the next command's response or leaked into the live output stream.
func TestReadResponse_ResyncAfterTimeout(t *testing.T) {
	input := strings.Join([]string{
		"{\"Success\":false,\"Data\":null,\"Error\":{\"Message\":\"old\"}}", // leftover JSON
		EndMarker,                    // leftover end marker
		BeginMarker,                  // current command's begin
		"{\"Success\":true,\"Data\":123,\"Error\":null}",
		EndMarker,
		"",
	}, "\n")

	r := &Runner{
		stdoutScanner: bufio.NewScanner(strings.NewReader(input)),
		running:       true,
		timeout:       time.Second,
		desynced:      true, // simulate a previous response that timed out
		liveOutputCh:  make(chan string, 8),
		outCh:         make(chan outLine, 8),
		dead:          make(chan struct{}),
	}

	data, err := r.readResponse(context.Background(), BeginMarker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.desynced {
		t.Error("desynced flag should be cleared after resync")
	}

	got := strings.TrimSpace(string(data))
	want := "{\"Success\":true,\"Data\":123,\"Error\":null}"
	if got != want {
		t.Errorf("resynced response = %q, want %q", got, want)
	}
}

// TestReadResponse_NormalResponse verifies the normal path still works when
// there is no desynchronization.
func TestReadResponse_NormalResponse(t *testing.T) {
	input := strings.Join([]string{
		BeginMarker,
		"{\"Success\":true,\"Data\":\"ok\",\"Error\":null}",
		EndMarker,
		"",
	}, "\n")

	r := &Runner{
		stdoutScanner: bufio.NewScanner(strings.NewReader(input)),
		running:       true,
		timeout:       time.Second,
		liveOutputCh:  make(chan string, 8),
		outCh:         make(chan outLine, 8),
		dead:          make(chan struct{}),
	}

	data, err := r.readResponse(context.Background(), BeginMarker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "\"Success\":true") {
		t.Errorf("unexpected response: %q", string(data))
	}
}
