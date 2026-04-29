//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
)

// AddFont installs a font on the system.
// The name can be a font file name (e.g., 'MyFont.ttf') or font registry name.
func (s *Session) AddFont(name string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Add-ADTFont -Name %s", cmdbuilder.EscapeString(name))
	return s.executeVoid(ctx, cmd)
}

// RemoveFont removes a font from the system.
func (s *Session) RemoveFont(name string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTFont -Name %s", cmdbuilder.EscapeString(name))
	return s.executeVoid(ctx, cmd)
}
