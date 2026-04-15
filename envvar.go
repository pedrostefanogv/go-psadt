//go:build windows

package psadt

import (
	"fmt"

	"github.com/peterondra/go-psadt/internal/cmdbuilder"
	"github.com/peterondra/go-psadt/internal/parser"
	"github.com/peterondra/go-psadt/types"
)

// GetEnvironmentVariable gets an environment variable value.
func (s *Session) GetEnvironmentVariable(variable string, target ...types.EnvironmentVariableTarget) (string, error) {
	ctx, cancel := s.client.defaultContext()
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
	ctx, cancel := s.client.defaultContext()
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
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTEnvironmentVariable -Variable %s", cmdbuilder.EscapeString(variable))
	if len(target) > 0 && target[0] != "" {
		cmd += fmt.Sprintf(" -Target %s", cmdbuilder.EscapeString(string(target[0])))
	}
	return s.executeVoid(ctx, cmd)
}
