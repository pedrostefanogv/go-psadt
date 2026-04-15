//go:build windows

package types

// SessionConfig is the configuration for opening an ADT session.
type SessionConfig struct {
	AppVendor   string `ps:"AppVendor"`
	AppName     string `ps:"AppName"`
	AppVersion  string `ps:"AppVersion"`
	AppArch     string `ps:"AppArch"`
	AppLang     string `ps:"AppLang"`
	AppRevision string `ps:"AppRevision"`

	AppScriptVersion string `ps:"AppScriptVersion"`
	AppScriptDate    string `ps:"AppScriptDate"`
	AppScriptAuthor  string `ps:"AppScriptAuthor"`

	DeploymentType         DeploymentType `ps:"DeploymentType"`
	DeployMode             DeployMode     `ps:"DeployMode"`
	RequireAdmin           bool           `ps:"RequireAdmin,switch"`
	TerminalServerMode     bool           `ps:"TerminalServerMode,switch"`
	DisableLogging         bool           `ps:"DisableLogging,switch"`
	SuppressRebootPassThru bool           `ps:"SuppressRebootPassThru,switch"`

	AppProcessesToClose []ProcessDefinition `ps:"AppProcessesToClose"`
	AppSuccessExitCodes []int               `ps:"AppSuccessExitCodes"`
	AppRebootExitCodes  []int               `ps:"AppRebootExitCodes"`

	InstallName  string `ps:"InstallName"`
	InstallTitle string `ps:"InstallTitle"`
	LogName      string `ps:"LogName"`

	ScriptDirectory string `ps:"ScriptDirectory"`
	DirFiles        string `ps:"DirFiles"`
	DirSupportFiles string `ps:"DirSupportFiles"`

	DefaultMsiFile               string   `ps:"DefaultMsiFile"`
	DefaultMstFile               string   `ps:"DefaultMstFile"`
	DefaultMspFiles              []string `ps:"DefaultMspFiles"`
	DisableDefaultMsiProcessList bool     `ps:"DisableDefaultMsiProcessList,switch"`
	ForceMsiDetection            bool     `ps:"ForceMsiDetection,switch"`

	ForceWimDetection  bool `ps:"ForceWimDetection,switch"`
	NoSessionDetection bool `ps:"NoSessionDetection,switch"`
	NoOobeDetection    bool `ps:"NoOobeDetection,switch"`
	NoProcessDetection bool `ps:"NoProcessDetection,switch"`
}

// SessionProperties contains read-only properties of an open ADT session.
type SessionProperties struct {
	CurrentDate     string `json:"CurrentDate"`
	CurrentDateTime string `json:"CurrentDateTime"`
	CurrentTime     string `json:"CurrentTime"`
	InstallPhase    string `json:"InstallPhase"`
	LogPath         string `json:"LogPath"`
	UseDefaultMsi   bool   `json:"UseDefaultMsi"`

	DeployAppScriptFriendlyName string `json:"DeployAppScriptFriendlyName"`
	DeployAppScriptParameters   string `json:"DeployAppScriptParameters"`
	DeployAppScriptVersion      string `json:"DeployAppScriptVersion"`
}
