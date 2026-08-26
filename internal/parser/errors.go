//go:build windows

package parser

import (
	"fmt"
	"strings"
)

// PSADTError represents an error returned by a PSADT PowerShell command.
type PSADTError struct {
	Message    string
	Type       string
	StackTrace string
	ExitCode   int
}

// Error implements the error interface.
func (e *PSADTError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("PSADT error [%s]: %s", e.Type, e.Message)
	}
	return fmt.Sprintf("PSADT error: %s", e.Message)
}

// NewPSADTError creates a PSADTError from an ErrorDetail.
func NewPSADTError(detail *ErrorDetail) *PSADTError {
	if detail == nil {
		return &PSADTError{
			Message: "unknown PowerShell error",
		}
	}

	return &PSADTError{
		Message:    detail.Message,
		Type:       detail.Type,
		StackTrace: detail.StackTrace,
	}
}

// NewPSADTErrorWithCode creates a PSADTError with an exit code.
func NewPSADTErrorWithCode(message string, exitCode int) *PSADTError {
	return &PSADTError{
		Message:  message,
		ExitCode: exitCode,
	}
}

// IsPSADTError checks if an error is a PSADTError.
func IsPSADTError(err error) (*PSADTError, bool) {
	if err == nil {
		return nil, false
	}
	if psErr, ok := err.(*PSADTError); ok {
		return psErr, true
	}
	return nil, false
}

// IsExitCode checks if the error has a specific exit code.
func IsExitCode(err error, code int) bool {
	psErr, ok := IsPSADTError(err)
	if !ok {
		return false
	}
	return psErr.ExitCode == code
}

// IsRebootRequired checks if the error indicates a reboot is required.
func IsRebootRequired(err error) bool {
	return IsExitCode(err, 3010) || IsExitCode(err, 1641)
}

// IsUserCancelled checks if the error indicates the user cancelled.
func IsUserCancelled(err error) bool {
	return IsExitCode(err, 1602)
}

// IsAccessDenied checks if the error indicates access was denied.
func IsAccessDenied(err error) bool {
	if IsExitCode(err, 5) {
		return true
	}
	psErr, ok := IsPSADTError(err)
	if !ok {
		return false
	}
	return psErr.Type == "System.UnauthorizedAccessException"
}

// IsTimeout checks if the error indicates a timeout occurred.
func IsTimeout(err error) bool {
	if psErr, ok := IsPSADTError(err); ok {
		return psErr.Type == "System.TimeoutException"
	}
	// Also check context deadline exceeded
	if err != nil {
		msg := err.Error()
		return containsAny(msg, "timeout", "deadline exceeded", "context deadline exceeded")
	}
	return false
}

// IsFileNotFound checks if the error indicates a file was not found.
func IsFileNotFound(err error) bool {
	if psErr, ok := IsPSADTError(err); ok {
		return psErr.Type == "System.IO.FileNotFoundException" ||
			psErr.Type == "System.Management.Automation.ItemNotFoundException"
	}
	if err != nil {
		msg := err.Error()
		return containsAny(msg, "file not found", "cannot find path", "does not exist")
	}
	return false
}

// IsNetworkError checks if the error indicates a network-related failure.
func IsNetworkError(err error) bool {
	psErr, ok := IsPSADTError(err)
	if !ok {
		return false
	}
	return psErr.Type == "System.Net.WebException" ||
		psErr.Type == "System.Net.HttpRequestException" ||
		psErr.Type == "System.Net.Sockets.SocketException"
}

// containsAny checks if s contains any of the given substrings (case-insensitive).
func containsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
