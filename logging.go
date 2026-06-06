//go:build windows

package psadt

import (
	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/types"
)

// WriteLogEntry writes an entry to the PSADT log file.
func (s *Session) WriteLogEntry(opts types.LogEntryOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Write-ADTLogEntry", opts)
	return s.executeVoid(ctx, cmd)
}

// WriteLogEntryInfo writes an informational log entry.
func (s *Session) WriteLogEntryInfo(message, source string) error {
	return s.WriteLogEntry(types.LogEntryOptions{
		Message:  message,
		Source:   source,
		Severity: types.LogInfo,
	})
}

// WriteLogEntryWarning writes a warning log entry.
func (s *Session) WriteLogEntryWarning(message, source string) error {
	return s.WriteLogEntry(types.LogEntryOptions{
		Message:  message,
		Source:   source,
		Severity: types.LogWarning,
	})
}

// WriteLogEntryError writes an error log entry.
func (s *Session) WriteLogEntryError(message, source string) error {
	return s.WriteLogEntry(types.LogEntryOptions{
		Message:  message,
		Source:   source,
		Severity: types.LogError,
	})
}
