//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// GetConfig retrieves the current PSADT configuration.
func (s *Session) GetConfig() (*types.ADTConfig, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTConfig")
	if err != nil {
		return nil, err
	}
	var result types.ADTConfig
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetStringTable retrieves the current PSADT string table.
func (s *Session) GetStringTable() (*types.ADTStringTable, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTStringTable")
	if err != nil {
		return nil, err
	}
	var result types.ADTStringTable
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDeferHistory retrieves the defer history for the current deployment.
func (s *Session) GetDeferHistory() (*types.DeferHistory, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	data, err := s.execute(ctx, "Get-ADTDeferHistory")
	if err != nil {
		return nil, err
	}
	var result types.DeferHistory
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetDeferHistory sets the defer history for the current deployment.
func (s *Session) SetDeferHistory(opts types.DeferHistory) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTDeferHistory", opts)
	return s.executeVoid(ctx, cmd)
}

// ResetDeferHistory resets the defer history for the current deployment.
func (s *Session) ResetDeferHistory() error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	return s.executeVoid(ctx, "Reset-ADTDeferHistory")
}

// SetPowerShellCulture sets the PowerShell culture for the current session.
func (s *Session) SetPowerShellCulture(cultureName string) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Set-ADTPowerShellCulture -CultureName %s", cmdbuilder.EscapeString(cultureName))
	return s.executeVoid(ctx, cmd)
}
