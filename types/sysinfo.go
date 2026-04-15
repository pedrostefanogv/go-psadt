//go:build windows

package types

// LoggedOnUser represents a logged-on user.
type LoggedOnUser struct {
	NTAccount string `json:"NTAccount"`
	SID       string `json:"SID"`
	IsAdmin   bool   `json:"IsAdmin"`
	SessionID int    `json:"SessionId"`
}

// UserProfile represents a user profile.
type UserProfile struct {
	NTAccount   string `json:"NTAccount"`
	SID         string `json:"SID"`
	ProfilePath string `json:"ProfilePath"`
}

// PendingRebootInfo result of Get-ADTPendingReboot.
type PendingRebootInfo struct {
	ComputerName          string `json:"ComputerName"`
	LastBootUpTime        string `json:"LastBootUpTime"`
	IsSystemRebootPending bool   `json:"IsSystemRebootPending"`
	IsCBServicing         bool   `json:"IsCBServicing"`
	IsWindowsUpdate       bool   `json:"IsWindowsUpdate"`
	IsSCCMClientReboot    bool   `json:"IsSCCMClientReboot"`
	IsFileRenameOps       bool   `json:"IsFileRenameOps"`
}

// OSInfo represents operating system information.
type OSInfo struct {
	Name         string `json:"Name"`
	Version      string `json:"Version"`
	Architecture string `json:"Architecture"`
	ServicePack  string `json:"ServicePack"`
	BuildNumber  string `json:"BuildNumber"`
}

// BatteryInfo represents battery information.
type BatteryInfo struct {
	IsLaptop           bool `json:"IsLaptop"`
	IsUsingACPower     bool `json:"IsUsingACPower"`
	BatteryChargeLevel int  `json:"BatteryChargeLevel,omitempty"`
}

// ExecutableInfo represents information about an executable.
type ExecutableInfo struct {
	FileName         string `json:"FileName"`
	FileVersion      string `json:"FileVersion"`
	ProductVersion   string `json:"ProductVersion"`
	ProductName      string `json:"ProductName"`
	CompanyName      string `json:"CompanyName"`
	FileDescription  string `json:"FileDescription"`
	InternalName     string `json:"InternalName"`
	OriginalFilename string `json:"OriginalFilename"`
	LegalCopyright   string `json:"LegalCopyright"`
}

// WindowTitle represents a window title result.
type WindowTitle struct {
	WindowTitle   string `json:"WindowTitle"`
	ParentProcess string `json:"ParentProcess"`
	ProcessId     int    `json:"ProcessId"`
}

// GetWindowTitleOptions options for Get-ADTWindowTitle.
type GetWindowTitleOptions struct {
	WindowTitle             string `ps:"WindowTitle"`
	GetDialogBoxTitle       bool   `ps:"GetDialogBoxTitle,switch"`
	DisableWildcardMatching bool   `ps:"DisableWildcardMatching,switch"`
}

// UserProfileOptions options for Get-ADTUserProfiles.
type UserProfileOptions struct {
	ExcludeNTAccount       []string `ps:"ExcludeNTAccount"`
	IncludeNTAccount       []string `ps:"IncludeNTAccount"`
	ExcludeSystemProfiles  bool     `ps:"ExcludeSystemProfiles,switch"`
	ExcludeServiceProfiles bool     `ps:"ExcludeServiceProfiles,switch"`
	ExcludeDefaultUser     bool     `ps:"ExcludeDefaultUser,switch"`
}
