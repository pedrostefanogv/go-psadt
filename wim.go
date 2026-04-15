//go:build windows

package psadt

import (
	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// MountWimFile mounts a WIM file.
func (s *Session) MountWimFile(opts types.MountWimOptions) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Mount-ADTWimFile", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// DismountWimFile dismounts a WIM file.
func (s *Session) DismountWimFile(opts types.DismountWimOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Dismount-ADTWimFile", opts)
	return s.executeVoid(ctx, cmd)
}

// NewZipFile creates a new ZIP archive.
func (s *Session) NewZipFile(opts types.NewZipOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("New-ADTZipFile", opts)
	return s.executeVoid(ctx, cmd)
}
