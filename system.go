//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/types"
)

// UpdateDesktop refreshes the desktop (F5 equivalent).
func (s *Session) UpdateDesktop() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Update-ADTDesktop")
}

// UpdateGroupPolicy forces a Group Policy update.
func (s *Session) UpdateGroupPolicy() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Update-ADTGroupPolicy")
}

// InstallMSUpdates installs Microsoft Updates from a specified directory.
func (s *Session) InstallMSUpdates(directory string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Install-ADTMSUpdates -Directory %s", cmdbuilder.EscapeString(directory))
	return s.executeVoid(ctx, cmd)
}

// InstallSCCMSoftwareUpdates triggers SCCM software updates installation.
func (s *Session) InstallSCCMSoftwareUpdates() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Install-ADTSCCMSoftwareUpdates")
}

// InvokeSCCMTask invokes an SCCM client task.
func (s *Session) InvokeSCCMTask(opts types.SCCMTaskOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Invoke-ADTSCCMTask", opts)
	return s.executeVoid(ctx, cmd)
}

// EnableTerminalServerInstallMode enables Terminal Server install mode.
func (s *Session) EnableTerminalServerInstallMode() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Enable-ADTTerminalServerInstallMode")
}

// DisableTerminalServerInstallMode disables Terminal Server install mode.
func (s *Session) DisableTerminalServerInstallMode() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Disable-ADTTerminalServerInstallMode")
}
