//go:build windows

package types

// ActiveSetupOptions options for Set-ADTActiveSetup.
type ActiveSetupOptions struct {
	StubExePath           string `ps:"StubExePath"`
	Arguments             string `ps:"Arguments"`
	Description           string `ps:"Description"`
	Key                   string `ps:"Key"`
	Wow6432Node           bool   `ps:"Wow6432Node,switch"`
	Version               string `ps:"Version"`
	Locale                string `ps:"Locale"`
	DisableActiveSetup    bool   `ps:"DisableActiveSetup,switch"`
	PurgeActiveSetupKey   bool   `ps:"PurgeActiveSetupKey,switch"`
	ExecuteForCurrentUser bool   `ps:"ExecuteForCurrentUser,switch"`
}
