//go:build windows

package psadt

import (
	"github.com/peterondra/go-psadt/internal/cmdbuilder"
	"github.com/peterondra/go-psadt/types"
)

// SetActiveSetup creates or modifies an Active Setup registry entry.
func (s *Session) SetActiveSetup(opts types.ActiveSetupOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTActiveSetup", opts)
	return s.executeVoid(ctx, cmd)
}
