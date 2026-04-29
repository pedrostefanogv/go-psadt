//go:build windows

package psadt

import (
	"context"
	"fmt"
	"strings"

	"github.com/pedrostefanogv/go-psadt/internal/parser"
)

// ExecuteBatch runs multiple PowerShell commands in a single round-trip
// to the PowerShell process. All commands are joined and wrapped together,
// significantly reducing latency for multi-step operations. The returned
// value is the JSON output of the last command in the batch.
//
// Example:
//
//	data, err := session.ExecuteBatch(ctx, []string{
//	    "Get-ADTFreeDiskSpace",
//	    "Test-ADTCallerIsAdmin",
//	})
func (s *Session) ExecuteBatch(ctx context.Context, commands []string) ([]byte, error) {
	if len(commands) == 0 {
		return nil, nil
	}
	s.client.logger.Debug("executing batch", "count", len(commands))
	return s.runner.ExecuteBatch(ctx, commands)
}

// ExecuteRawScript runs an arbitrary PowerShell script block within the
// PSADT session. The script inherits the module context (session state,
// variables, etc.). Use this as an escape hatch for PSADT features not
// yet exposed by the Go API.
//
// Note: The script is wrapped with try/catch and markers automatically.
func (s *Session) ExecuteRawScript(ctx context.Context, script string) ([]byte, error) {
	s.client.logger.Debug("executing raw script", "length", len(script))
	return s.runner.Execute(ctx, script)
}

// ExecuteRawScriptU runs a raw PowerShell script that returns no data.
func (s *Session) ExecuteRawVoidScript(ctx context.Context, script string) error {
	s.client.logger.Debug("executing raw void script", "length", len(script))
	_, err := s.runner.ExecuteVoid(ctx, script)
	return err
}

// GetRegistryKeyString is a typed version of GetRegistryKey that returns a
// string value. It's a convenience wrapper for the common case of reading
// string registry values.
func (s *Session) GetRegistryKeyString(key, name string) (string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("(Get-ADTRegistryKey -Key %s -Name %s).%s",
		escapeArg(key), escapeArg(name), escapeArg(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// GetRegistryKeyDWord is a typed version of GetRegistryKey that returns a
// DWord (uint32) value.
func (s *Session) GetRegistryKeyDWord(key, name string) (uint32, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("(Get-ADTRegistryKey -Key %s -Name %s).%s",
		escapeArg(key), escapeArg(name), escapeArg(name))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return 0, err
	}
	return parser.ParseUint32(data)
}

// escapeArg is a lightweight string escaper for inline command building
// that doesn't depend on cmdbuilder import cycles.
func escapeArg(s string) string {
	if s == "" {
		return "''"
	}
	escaped := strings.ReplaceAll(s, "'", "''")
	return fmt.Sprintf("'%s'", escaped)
}
