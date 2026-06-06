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

// ConvertRegistryPathOptions options for Convert-ADTRegistryPath.
type ConvertRegistryPathOptions struct {
	RegistryPath string `ps:"RegistryPath"`
	ToShortPath  bool   `ps:"ToShortPath,switch"`
}

// RegistryPathInfo result of Convert-ADTRegistryPath.
type RegistryPathInfo struct {
	RegistryPath    string `json:"RegistryPath"`
	RegistryHive    string `json:"RegistryHive"`
	RegistryKey     string `json:"RegistryKey"`
	RegistryKeyPath string `json:"RegistryKeyPath"`
}

// AllUsersRegistryOptions options for Invoke-ADTAllUsersRegistryAction.
type AllUsersRegistryOptions struct {
	UserProfiles []UserProfile `ps:"UserProfiles"`
}
