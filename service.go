//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// StartServiceAndDependencies starts a Windows service and its dependencies.
func (s *Session) StartServiceAndDependencies(opts types.ServiceOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Start-ADTServiceAndDependencies", opts)
	return s.executeVoid(ctx, cmd)
}

// StopServiceAndDependencies stops a Windows service and its dependencies.
func (s *Session) StopServiceAndDependencies(opts types.ServiceOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Stop-ADTServiceAndDependencies", opts)
	return s.executeVoid(ctx, cmd)
}

// GetServiceStartMode gets the start mode of a Windows service.
func (s *Session) GetServiceStartMode(name string) (types.ServiceStartMode, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTServiceStartMode -Name %s", cmdbuilder.EscapeString(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	str, err := parser.ParseString(data)
	if err != nil {
		return "", err
	}
	return types.ServiceStartMode(str), nil
}

// SetServiceStartMode sets the start mode of a Windows service.
func (s *Session) SetServiceStartMode(name string, mode types.ServiceStartMode) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Set-ADTServiceStartMode -Name %s -StartMode %s",
		cmdbuilder.EscapeString(name),
		cmdbuilder.EscapeString(string(mode)))
	return s.executeVoid(ctx, cmd)
}

// TestServiceExists checks if a Windows service exists.
func (s *Session) TestServiceExists(name string) (bool, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Test-ADTServiceExists -Name %s", cmdbuilder.EscapeString(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}
