//go:build windows

package types

// GetRegistryKeyOptions options for Get-ADTRegistryKey.
type GetRegistryKeyOptions struct {
	Key         string `ps:"Key"`
	Name        string `ps:"Name"`
	LiteralPath string `ps:"LiteralPath"`
	Wow6432Node bool   `ps:"Wow6432Node,switch"`
}

// SetRegistryKeyOptions options for Set-ADTRegistryKey.
type SetRegistryKeyOptions struct {
	Key         string            `ps:"Key"`
	Name        string            `ps:"Name"`
	Value       interface{}       `ps:"Value"`
	Type        RegistryValueKind `ps:"Type"`
	LiteralPath string            `ps:"LiteralPath"`
	Wow6432Node bool              `ps:"Wow6432Node,switch"`
}

// RemoveRegistryKeyOptions options for Remove-ADTRegistryKey.
type RemoveRegistryKeyOptions struct {
	Key         string `ps:"Key"`
	Name        string `ps:"Name"`
	LiteralPath string `ps:"LiteralPath"`
	Recurse     bool   `ps:"Recurse,switch"`
	Wow6432Node bool   `ps:"Wow6432Node,switch"`
}

// TestRegistryValueOptions options for Test-ADTRegistryValue.
type TestRegistryValueOptions struct {
	Key         string `ps:"Key"`
	Name        string `ps:"Name"`
	LiteralPath string `ps:"LiteralPath"`
	Wow6432Node bool   `ps:"Wow6432Node,switch"`
}

// AllUsersRegistryOptions options for Invoke-ADTAllUsersRegistryAction.
type AllUsersRegistryOptions struct {
	UserProfiles []UserProfile `ps:"UserProfiles"`
}
