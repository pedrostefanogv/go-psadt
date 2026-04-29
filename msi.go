//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// GetMsiExitCodeMessage gets the message for an MSI exit code.
func (s *Session) GetMsiExitCodeMessage(exitCode int) (string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTMsiExitCodeMessage -MsiExitCode %d", exitCode)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// GetMsiTableProperty gets properties from an MSI database table.
func (s *Session) GetMsiTableProperty(opts types.MsiTableOptions) (map[string]string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Get-ADTMsiTableProperty", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetMsiProperty sets a property in an MSI database.
func (s *Session) SetMsiProperty(opts types.SetMsiPropertyOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTMsiProperty", opts)
	return s.executeVoid(ctx, cmd)
}

// NewMsiTransform creates a new MSI transform file.
func (s *Session) NewMsiTransform(opts types.MsiTransformOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("New-ADTMsiTransform", opts)
	return s.executeVoid(ctx, cmd)
}
