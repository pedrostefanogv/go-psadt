//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/types"
)

// RegisterDll registers a DLL file.
func (s *Session) RegisterDll(filePath string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Register-ADTDll -FilePath %s", cmdbuilder.EscapeString(filePath))
	return s.executeVoid(ctx, cmd)
}

// UnregisterDll unregisters a DLL file.
func (s *Session) UnregisterDll(filePath string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Unregister-ADTDll -FilePath %s", cmdbuilder.EscapeString(filePath))
	return s.executeVoid(ctx, cmd)
}

// InvokeRegSvr32 invokes regsvr32.exe with the specified options.
func (s *Session) InvokeRegSvr32(opts types.RegSvr32Options) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Invoke-ADTRegSvr32", opts)
	return s.executeVoid(ctx, cmd)
}
