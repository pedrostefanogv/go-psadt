//go:build windows

package psadt

import (
	"github.com/peterondra/go-psadt/internal/cmdbuilder"
	"github.com/peterondra/go-psadt/internal/parser"
	"github.com/peterondra/go-psadt/types"
)

// GetApplication searches for installed applications matching the criteria.
func (s *Session) GetApplication(opts types.GetApplicationOptions) ([]types.InstalledApplication, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Get-ADTApplication", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result []types.InstalledApplication
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UninstallApplication uninstalls applications matching the criteria.
func (s *Session) UninstallApplication(opts types.UninstallApplicationOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Uninstall-ADTApplication", opts)
	return s.executeVoid(ctx, cmd)
}
