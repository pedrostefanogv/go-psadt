//go:build windows

package psadt

import (
	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/types"
)

// SetActiveSetup creates or modifies an Active Setup registry entry.
func (s *Session) SetActiveSetup(opts types.ActiveSetupOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTActiveSetup", opts)
	return s.executeVoid(ctx, cmd)
}
