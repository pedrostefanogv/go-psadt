//go:build windows

package types

// NewShortcutOptions options for New-ADTShortcut.
type NewShortcutOptions struct {
	Path             string `ps:"Path"`
	TargetPath       string `ps:"TargetPath"`
	Arguments        string `ps:"Arguments"`
	IconLocation     string `ps:"IconLocation"`
	IconIndex        int    `ps:"IconIndex"`
	Description      string `ps:"Description"`
	WorkingDirectory string `ps:"WorkingDirectory"`
	WindowStyle      int    `ps:"WindowStyle"`
	RunAsAdmin       bool   `ps:"RunAsAdmin,switch"`
	Hotkey           string `ps:"Hotkey"`
}

// SetShortcutOptions options for Set-ADTShortcut.
type SetShortcutOptions struct {
	Path             string `ps:"Path"`
	TargetPath       string `ps:"TargetPath"`
	Arguments        string `ps:"Arguments"`
	IconLocation     string `ps:"IconLocation"`
	IconIndex        int    `ps:"IconIndex"`
	Description      string `ps:"Description"`
	WorkingDirectory string `ps:"WorkingDirectory"`
	WindowStyle      int    `ps:"WindowStyle"`
	RunAsAdmin       bool   `ps:"RunAsAdmin,switch"`
	Hotkey           string `ps:"Hotkey"`
}

// ShortcutInfo information about a shortcut.
type ShortcutInfo struct {
	TargetPath       string `json:"TargetPath"`
	Arguments        string `json:"Arguments"`
	Description      string `json:"Description"`
	WorkingDirectory string `json:"WorkingDirectory"`
	WindowStyle      int    `json:"WindowStyle"`
	Hotkey           string `json:"Hotkey"`
	IconLocation     string `json:"IconLocation"`
	RunAsAdmin       bool   `json:"RunAsAdmin"`
}
