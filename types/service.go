//go:build windows

package types

// ServiceOptions options for Start/Stop-ADTServiceAndDependencies.
type ServiceOptions struct {
	Name                  string `ps:"Name"`
	SkipServiceExistsTest bool   `ps:"SkipServiceExistsTest,switch"`
	SkipDependentServices bool   `ps:"SkipDependentServices,switch"`
	PendingStatusWait     int    `ps:"PendingStatusWait"`
	PassThru              bool   `ps:"PassThru,switch"`
}
