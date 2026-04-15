//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// GetRegistryKey gets a registry key value.
func (s *Session) GetRegistryKey(opts types.GetRegistryKeyOptions) (interface{}, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Get-ADTRegistryKey", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result interface{}
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetRegistryKey sets a registry key value.
func (s *Session) SetRegistryKey(opts types.SetRegistryKeyOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTRegistryKey", opts)
	return s.executeVoid(ctx, cmd)
}

// RemoveRegistryKey removes a registry key or value.
func (s *Session) RemoveRegistryKey(opts types.RemoveRegistryKeyOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Remove-ADTRegistryKey", opts)
	return s.executeVoid(ctx, cmd)
}

// TestRegistryValue tests if a registry value exists.
func (s *Session) TestRegistryValue(opts types.TestRegistryValueOptions) (bool, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Test-ADTRegistryValue", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	return parser.ParseBool(data)
}

// InvokeAllUsersRegistryAction executes a script block against all user registry hives.
func (s *Session) InvokeAllUsersRegistryAction(scriptBlock string, opts ...types.AllUsersRegistryOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("Invoke-ADTAllUsersRegistryAction -ScriptBlock %s", cmdbuilder.FormatScriptBlock(scriptBlock))
	return s.executeVoid(ctx, cmd)
}
