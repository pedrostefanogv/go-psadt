//go:build windows

package psadt

import (
	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/types"
)

// WriteLogEntry writes an entry to the PSADT log file.
func (s *Session) WriteLogEntry(opts types.LogEntryOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Write-ADTLogEntry", opts)
	return s.executeVoid(ctx, cmd)
}
