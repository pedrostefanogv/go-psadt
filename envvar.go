//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// GetEnvironmentVariable gets an environment variable value.
func (s *Session) GetEnvironmentVariable(variable string, target ...types.EnvironmentVariableTarget) (string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTEnvironmentVariable -Variable %s", cmdbuilder.EscapeString(variable))
	if len(target) > 0 && target[0] != "" {
		cmd += fmt.Sprintf(" -Target %s", cmdbuilder.EscapeString(string(target[0])))
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// SetEnvironmentVariable sets an environment variable.
func (s *Session) SetEnvironmentVariable(variable, value string, target ...types.EnvironmentVariableTarget) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Set-ADTEnvironmentVariable -Variable %s -Value %s",
		cmdbuilder.EscapeString(variable),
		cmdbuilder.EscapeString(value))
	if len(target) > 0 && target[0] != "" {
		cmd += fmt.Sprintf(" -Target %s", cmdbuilder.EscapeString(string(target[0])))
	}
	return s.executeVoid(ctx, cmd)
}

// RemoveEnvironmentVariable removes an environment variable.
func (s *Session) RemoveEnvironmentVariable(variable string, target ...types.EnvironmentVariableTarget) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTEnvironmentVariable -Variable %s", cmdbuilder.EscapeString(variable))
	if len(target) > 0 && target[0] != "" {
		cmd += fmt.Sprintf(" -Target %s", cmdbuilder.EscapeString(string(target[0])))
	}
	return s.executeVoid(ctx, cmd)
}

// GetEnvironmentTable retrieves the complete PSADT environment table as a flat key→value map.
func (s *Session) GetEnvironmentTable() (types.EnvironmentTableInfo, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTEnvironmentTable")
	if err != nil {
		return nil, err
	}
	var result types.EnvironmentTableInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ExportEnvironmentTableToSessionState exports the environment table to the current
// PowerShell session state, making PSADT variables available as $envVarName.
func (s *Session) ExportEnvironmentTableToSessionState() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Export-ADTEnvironmentTableToSessionState")
}

// UpdateEnvironmentPsProvider updates the PowerShell Environment provider with
// the latest values from the PSADT environment table.
func (s *Session) UpdateEnvironmentPsProvider() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Update-ADTEnvironmentPsProvider")
}
