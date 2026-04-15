//go:build windows

package parser

import (
	"fmt"
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
