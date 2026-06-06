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
	ctx, cancel := s.getContext()
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
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Set-ADTRegistryKey", opts)
	return s.executeVoid(ctx, cmd)
}

// RemoveRegistryKey removes a registry key or value.
func (s *Session) RemoveRegistryKey(opts types.RemoveRegistryKeyOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Remove-ADTRegistryKey", opts)
	return s.executeVoid(ctx, cmd)
}

// TestRegistryValue tests if a registry value exists.
func (s *Session) TestRegistryValue(opts types.TestRegistryValueOptions) (bool, error) {
	ctx, cancel := s.getContext()
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
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Invoke-ADTAllUsersRegistryAction -ScriptBlock %s", cmdbuilder.FormatScriptBlock(scriptBlock))
	return s.executeVoid(ctx, cmd)
}

// ConvertRegistryPath converts a registry path between long and short forms.
func (s *Session) ConvertRegistryPath(registryPath string, toShort bool) (*types.RegistryPathInfo, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Convert-ADTRegistryPath -RegistryPath %s", cmdbuilder.EscapeString(registryPath))
	if toShort {
		cmd += " -ToShortPath"
	}
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result types.RegistryPathInfo
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRegistryKeyMultiString is a typed version of GetRegistryKey that returns a
// MultiString ([]string) value.
func (s *Session) GetRegistryKeyMultiString(key, name string) ([]string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("(Get-ADTRegistryKey -Key %s -Name %s).%s",
		escapeArg(key), escapeArg(name), escapeArg(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result []string
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetRegistryKeyBinary is a typed version of GetRegistryKey that returns a
// Binary ([]byte) value.
func (s *Session) GetRegistryKeyBinary(key, name string) ([]byte, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("(Get-ADTRegistryKey -Key %s -Name %s).%s",
		escapeArg(key), escapeArg(name), escapeArg(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result []byte
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetRegistryKeyQWord is a typed version of GetRegistryKey that returns a
// QWord (uint64) value.
func (s *Session) GetRegistryKeyQWord(key, name string) (uint64, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("(Get-ADTRegistryKey -Key %s -Name %s).%s",
		escapeArg(key), escapeArg(name), escapeArg(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return 0, err
	}
	return parser.ParseUint64(data)
}
