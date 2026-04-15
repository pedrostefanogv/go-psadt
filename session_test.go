//go:build windows

package psadt

import (
	"errors"
	"testing"
)

func TestIsExpectedSessionCloseRunnerTermination(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "eof after close", err: errors.New("PowerShell process ended before completing response"), want: true},
		{name: "runner already stopped", err: errors.New("PowerShell runner is not running"), want: true},
		{name: "broken pipe on close", err: errors.New("failed to write command to PowerShell: write /dev/stdin: broken pipe"), want: true},
		{name: "timeout remains actionable", err: errors.New("timeout waiting for PowerShell response after 30s"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpectedSessionCloseRunnerTermination(tt.err)
			if got != tt.want {
				t.Fatalf("isExpectedSessionCloseRunnerTermination(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
