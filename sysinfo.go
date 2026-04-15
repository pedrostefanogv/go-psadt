//go:build windows

package psadt

import (
	"fmt"

	"github.com/peterondra/go-psadt/internal/cmdbuilder"
	"github.com/peterondra/go-psadt/internal/parser"
	"github.com/peterondra/go-psadt/types"
)

// GetLoggedOnUser gets the list of logged-on users.
func (s *Session) GetLoggedOnUser() ([]types.LoggedOnUser, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTLoggedOnUser")
	if err != nil {
		return nil, err
	}
	var result []types.LoggedOnUser
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetFreeDiskSpace gets the free disk space in MB.
func (s *Session) GetFreeDiskSpace(drive ...string) (uint64, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := "Get-ADTFreeDiskSpace"
	if len(drive) > 0 && drive[0] != "" {
		cmd += fmt.Sprintf(" -Drive %s", cmdbuilder.EscapeString(drive[0]))
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return 0, err
	}
	return parser.ParseUint64(data)
}

// GetPendingReboot checks for pending reboots.
func (s *Session) GetPendingReboot() (*types.PendingRebootInfo, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTPendingReboot")
	if err != nil {
		return nil, err
	}
	var result types.PendingRebootInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOperatingSystemInfo gets operating system information.
func (s *Session) GetOperatingSystemInfo() (*types.OSInfo, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTOperatingSystemInfo")
	if err != nil {
		return nil, err
	}
	var result types.OSInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserProfiles gets user profiles on the system.
func (s *Session) GetUserProfiles(opts ...types.UserProfileOptions) ([]types.UserProfile, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	var cmd string
	if len(opts) > 0 {
		cmd = cmdbuilder.Build("Get-ADTUserProfiles", opts[0])
	} else {
		cmd = "Get-ADTUserProfiles"
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result []types.UserProfile
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetFileVersion gets the version of a file.
func (s *Session) GetFileVersion(filePath string) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTFileVersion -File %s", cmdbuilder.EscapeString(filePath))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// GetExecutableInfo gets detailed information about an executable.
func (s *Session) GetExecutableInfo(filePath string) (*types.ExecutableInfo, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTExecutableInfo -Path %s", cmdbuilder.EscapeString(filePath))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result types.ExecutableInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPEFileArchitecture gets the architecture of a PE file (x86, x64, etc.).
func (s *Session) GetPEFileArchitecture(filePath string) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTPEFileArchitecture -FilePath %s", cmdbuilder.EscapeString(filePath))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// GetWindowTitle gets windows matching a title pattern.
func (s *Session) GetWindowTitle(opts types.GetWindowTitleOptions) ([]types.WindowTitle, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Get-ADTWindowTitle", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result []types.WindowTitle
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPresentationSettingsEnabledUsers gets users with presentation settings enabled.
func (s *Session) GetPresentationSettingsEnabledUsers() ([]string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTPresentationSettingsEnabledUsers")
	if err != nil {
		return nil, err
	}
	var result []string
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUserNotificationState gets the current user notification state.
func (s *Session) GetUserNotificationState() (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTUserNotificationState")
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}
