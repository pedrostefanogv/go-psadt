//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// SendKeys sends keystrokes to a window.
func (s *Session) SendKeys(opts types.SendKeysOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Send-ADTKeys", opts)
	return s.executeVoid(ctx, cmd)
}

// ConvertToNTAccountOrSID converts between NTAccount and SID formats.
func (s *Session) ConvertToNTAccountOrSID(identity string, toSID bool) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Convert-ADTToNTAccountOrSID -Identity %s", cmdbuilder.EscapeString(identity))
	if toSID {
		cmd += " -ToSID"
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// SetItemPermission sets permissions on a file or folder.
func (s *Session) SetItemPermission(opts types.ItemPermissionOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTItemPermission", opts)
	return s.executeVoid(ctx, cmd)
}

// InvokeCommandWithRetries invokes a command with retry logic.
func (s *Session) InvokeCommandWithRetries(scriptBlock string, opts ...types.RetryOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Invoke-ADTCommandWithRetries -ScriptBlock %s", cmdbuilder.FormatScriptBlock(scriptBlock))
	if len(opts) > 0 {
		extras := cmdbuilder.Build("", opts[0])
		if extras != "" {
			cmd += " " + extras
		}
	}
	return s.executeVoid(ctx, cmd)
}

// GetUniversalDate gets a date/time string formatted for the current culture.
func (s *Session) GetUniversalDate(dateTime ...string) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := "Get-ADTUniversalDate"
	if len(dateTime) > 0 && dateTime[0] != "" {
		cmd += fmt.Sprintf(" -DateTime %s", cmdbuilder.EscapeString(dateTime[0]))
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// RemoveInvalidFileNameChars removes invalid characters from a filename.
func (s *Session) RemoveInvalidFileNameChars(name string) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTInvalidFileNameChars -Name %s", cmdbuilder.EscapeString(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// OutPowerShellEncodedCommand encodes a PowerShell command to Base64.
func (s *Session) OutPowerShellEncodedCommand(scriptBlock string) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Out-ADTPowerShellEncodedCommand -ScriptBlock %s", cmdbuilder.FormatScriptBlock(scriptBlock))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// NewTemplate creates a new PSADT template in a target directory.
func (s *Session) NewTemplate(destination string) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("New-ADTTemplate -Destination %s", cmdbuilder.EscapeString(destination))
	return s.executeVoid(ctx, cmd)
}
