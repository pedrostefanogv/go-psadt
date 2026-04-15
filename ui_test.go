//go:build windows

package psadt

import (
	"testing"

	"github.com/pedrostefanogv/go-psadt/types"
)

func TestNormalizeDialogBoxIcon(t *testing.T) {
	tests := []struct {
		name    string
		input   types.DialogSystemIcon
		want    types.DialogSystemIcon
		wantErr bool
	}{
		{name: "question passthrough", input: types.IconQuestion, want: types.IconQuestion},
		{name: "none passthrough", input: types.DialogSystemIcon("None"), want: types.DialogSystemIcon("None")},
		{name: "error to stop", input: types.IconError, want: types.DialogSystemIcon("Stop")},
		{name: "hand to stop", input: types.IconHand, want: types.DialogSystemIcon("Stop")},
		{name: "warning to exclamation", input: types.IconWarning, want: types.IconExclamation},
		{name: "asterisk to information", input: types.IconAsterisk, want: types.IconInformation},
		{name: "shield rejected", input: types.IconShield, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDialogBoxIcon(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeDialogBoxIcon(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeProgressOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   types.ProgressOptions
		wantErr bool
	}{
		{name: "default valid", input: types.ProgressOptions{StatusMessage: "Installing..."}},
		{name: "topmost compatibility valid", input: types.ProgressOptions{TopMost: true}},
		{name: "conflicting topmost flags", input: types.ProgressOptions{TopMost: true, NotTopMost: true}, wantErr: true},
		{name: "inplace rejected", input: types.ProgressOptions{InPlace: true}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeProgressOptions(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNormalizePromptOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   types.PromptOptions
		wantErr bool
	}{
		{name: "default valid", input: types.PromptOptions{Message: "Continue?"}},
		{name: "topmost compatibility valid", input: types.PromptOptions{TopMost: true}},
		{name: "conflicting topmost flags", input: types.PromptOptions{TopMost: true, NotTopMost: true}, wantErr: true},
		{name: "conflicting timeout flags", input: types.PromptOptions{ExitOnTimeout: true, NoExitOnTimeout: true}, wantErr: true},
		{name: "secure input without request input", input: types.PromptOptions{SecureInput: true}, wantErr: true},
		{name: "request input with list items", input: types.PromptOptions{RequestInput: true, ListItems: []string{"A", "B"}}, wantErr: true},
		{name: "default index out of range", input: types.PromptOptions{ListItems: []string{"A"}, DefaultIndex: 1}, wantErr: true},
		{name: "list selection valid", input: types.PromptOptions{ListItems: []string{"A", "B"}, DefaultIndex: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizePromptOptions(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNormalizeWelcomeOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   types.WelcomeOptions
		wantErr bool
	}{
		{name: "default valid", input: types.WelcomeOptions{Title: "My App", Subtitle: "Installing"}},
		{name: "silent valid", input: types.WelcomeOptions{Title: "My App", Subtitle: "Installing", Silent: true}},
		{name: "allow defer valid", input: types.WelcomeOptions{Title: "My App", Subtitle: "Installing", AllowDefer: true, DeferTimes: 3}},
		{name: "silent with allow defer rejected", input: types.WelcomeOptions{Silent: true, AllowDefer: true}, wantErr: true},
		{name: "silent with allow defer close processes rejected", input: types.WelcomeOptions{Silent: true, AllowDeferCloseProcesses: true}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeWelcomeOptions(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
