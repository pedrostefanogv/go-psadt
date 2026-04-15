//go:build windows

package types

// ItemPermissionOptions options for Set-ADTItemPermission.
type ItemPermissionOptions struct {
	Path           string `ps:"Path"`
	User           string `ps:"User"`
	Permission     string `ps:"Permission"`
	PermissionType string `ps:"PermissionType"`
	Inheritance    string `ps:"Inheritance"`
	Propagation    string `ps:"Propagation"`
}
