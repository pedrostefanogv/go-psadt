//go:build windows

package psadt

import (
	"github.com/pedrostefanogv/go-psadt/internal/parser"
)

// PSADTError is the typed error returned by PSADT commands through the
// wrapper. It is re-exported here so consumers of this library (e.g. RMM
// agents) can classify errors without importing internal packages.
type PSADTError = parser.PSADTError

// IsPSADTError checks if err is a PSADTError and returns it.
func IsPSADTError(err error) (*PSADTError, bool) {
	return parser.IsPSADTError(err)
}

// IsExitCode reports whether err carries the given process exit code.
func IsExitCode(err error, code int) bool {
	return parser.IsExitCode(err, code)
}

// IsRebootRequired reports whether the error indicates a reboot is required
// (exit codes 3010 and 1641).
func IsRebootRequired(err error) bool {
	return parser.IsRebootRequired(err)
}

// IsUserCancelled reports whether the error indicates the user cancelled the
// operation (exit code 1602).
func IsUserCancelled(err error) bool {
	return parser.IsUserCancelled(err)
}

// IsAccessDenied reports whether the error indicates insufficient privileges
// (exit code 5 or UnauthorizedAccessException).
func IsAccessDenied(err error) bool {
	return parser.IsAccessDenied(err)
}

// IsTimeout reports whether the error indicates a timeout — either a PSADT
// TimeoutException or a Go context deadline exceeded.
func IsTimeout(err error) bool {
	return parser.IsTimeout(err)
}

// IsFileNotFound reports whether the error indicates a missing file or path.
func IsFileNotFound(err error) bool {
	return parser.IsFileNotFound(err)
}

// IsNetworkError reports whether the error indicates a network-related
// failure (WebException, HttpRequestException or SocketException).
func IsNetworkError(err error) bool {
	return parser.IsNetworkError(err)
}
