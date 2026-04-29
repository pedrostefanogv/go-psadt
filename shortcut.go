//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// NewShortcut creates a new shortcut.
func (s *Session) NewShortcut(opts types.NewShortcutOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("New-ADTShortcut", opts)
	return s.executeVoid(ctx, cmd)
}

// SetShortcut modifies an existing shortcut.
func (s *Session) SetShortcut(opts types.SetShortcutOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTShortcut", opts)
	return s.executeVoid(ctx, cmd)
}

// GetShortcut retrieves information about a shortcut.
func (s *Session) GetShortcut(path string) (*types.ShortcutInfo, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTShortcut -Path %s", cmdbuilder.EscapeString(path))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result types.ShortcutInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
