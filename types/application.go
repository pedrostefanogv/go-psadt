//go:build windows

package types

// InstalledApplication represents an installed application on the system.
type InstalledApplication struct {
	PSPath               string `json:"PSPath"`
	PSParentPath         string `json:"PSParentPath"`
	PSChildName          string `json:"PSChildName"`
	ProductCode          string `json:"ProductCode"`
	DisplayName          string `json:"DisplayName"`
	DisplayVersion       string `json:"DisplayVersion"`
	UninstallString      string `json:"UninstallString"`
	QuietUninstallString string `json:"QuietUninstallString"`
	InstallSource        string `json:"InstallSource"`
	InstallLocation      string `json:"InstallLocation"`
	InstallDate          string `json:"InstallDate"`
	Publisher            string `json:"Publisher"`
	HelpLink             string `json:"HelpLink"`
	EstimatedSize        int64  `json:"EstimatedSize"`
	SystemComponent      int    `json:"SystemComponent"`
	WindowsInstaller     int    `json:"WindowsInstaller"`
	Is64BitApplication   bool   `json:"Is64BitApplication"`
}

// GetApplicationOptions options for Get-ADTApplication.
type GetApplicationOptions struct {
	Name                      []string        `ps:"Name"`
	NameMatch                 NameMatch       `ps:"NameMatch"`
	ProductCode               string          `ps:"ProductCode"`
	ApplicationType           ApplicationType `ps:"ApplicationType"`
	IncludeUpdatesAndHotfixes bool            `ps:"IncludeUpdatesAndHotfixes,switch"`
	FilterScript              string          `ps:"FilterScript"`
}

// UninstallApplicationOptions options for Uninstall-ADTApplication.
type UninstallApplicationOptions struct {
	Name                      []string        `ps:"Name"`
	NameMatch                 NameMatch       `ps:"NameMatch"`
	ProductCode               string          `ps:"ProductCode"`
	ApplicationType           ApplicationType `ps:"ApplicationType"`
	IncludeUpdatesAndHotfixes bool            `ps:"IncludeUpdatesAndHotfixes,switch"`
	FilterScript              string          `ps:"FilterScript"`
	AdditionalArgumentList    []string        `ps:"AdditionalArgumentList"`
	SecureArgumentList        bool            `ps:"SecureArgumentList,switch"`
	LoggingOptions            string          `ps:"LoggingOptions"`
	LogFileName               string          `ps:"LogFileName"`
	PassThru                  bool            `ps:"PassThru,switch"`
}

// RunningProcess represents a running process.
type RunningProcess struct {
	ProcessName     string `json:"ProcessName"`
	ProcessId       int    `json:"ProcessId"`
	MainWindowTitle string `json:"MainWindowTitle"`
}
