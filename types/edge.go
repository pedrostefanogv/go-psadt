//go:build windows

package types

// EdgeExtensionOptions options for Add-ADTEdgeExtension.
type EdgeExtensionOptions struct {
	ExtensionID            string `ps:"ExtensionID"`
	UpdateURL              string `ps:"UpdateUrl"`
	InstallationMode       string `ps:"InstallationMode"`
	MinimumVersionRequired string `ps:"MinimumVersionRequired"`
}
