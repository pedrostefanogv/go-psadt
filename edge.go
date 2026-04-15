//go:build windows

package psadt

import (
	"fmt"

	"github.com/peterondra/go-psadt/internal/cmdbuilder"
	"github.com/peterondra/go-psadt/types"
)

// AddEdgeExtension adds a Microsoft Edge extension via registry policy.
func (s *Session) AddEdgeExtension(opts types.EdgeExtensionOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Add-ADTEdgeExtension", opts)
	return s.executeVoid(ctx, cmd)
}

// RemoveEdgeExtension removes a Microsoft Edge extension policy.
func (s *Session) RemoveEdgeExtension(extensionID string) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTEdgeExtension -ExtensionID %s", cmdbuilder.EscapeString(extensionID))
	return s.executeVoid(ctx, cmd)
}
