//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
)

// TestUserInFocusMode checks if the current user has Focus Assist / Quiet Hours enabled.
// This is critical for RMM agents to avoid interrupting users during presentations or focus time.
func (s *Session) TestUserInFocusMode() (bool, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTUserInFocusMode")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// TestSessionActive checks if there is an active deployment session.
// Useful for RMM agents to verify session state before operations.
func (s *Session) TestSessionActive() (bool, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	data, err := s.execute(ctx, "Test-ADTSessionActive")
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// GetUserToastNotificationMode retrieves the current toast notification mode.
func (s *Session) GetUserToastNotificationMode() (string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTUserToastNotificationMode")
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// RemoveDesktopShortcut removes a shortcut from the common/user desktop.
func (s *Session) RemoveDesktopShortcut(name string, allUsers bool) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTDesktopShortcut -Name %s", cmdbuilder.EscapeString(name))
	if allUsers {
		cmd += " -AllUsers"
	}
	return s.executeVoid(ctx, cmd)
}
