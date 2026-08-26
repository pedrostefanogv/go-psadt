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
	AppName         string `json:"AppName"`
	AppVendor       string `json:"AppVendor"`
	AppVersion      string `json:"AppVersion"`
	DeploymentType  string `json:"DeploymentType"`
	DeployMode      string `json:"DeployMode"`
	ScriptDirectory string `json:"ScriptDirectory"`
	LogPath         string `json:"LogPath"`
	LogName         string `json:"LogName"`
	CurrentDate     string `json:"CurrentDate"`
	CurrentDateTime string `json:"CurrentDateTime"`
	CurrentTime     string `json:"CurrentTime"`
	InstallPhase    string `json:"InstallPhase"`
	UseDefaultMsi   bool   `json:"UseDefaultMsi"`

	DeployAppScriptFriendlyName string `json:"DeployAppScriptFriendlyName"`
	DeployAppScriptParameters   string `json:"DeployAppScriptParameters"`
	DeployAppScriptVersion      string `json:"DeployAppScriptVersion"`
}

// SessionConfigBuilder provides a fluent API for building SessionConfig.
// Use it to construct deployment configurations with method chaining.
//
// Example:
//
//	cfg := types.NewSessionConfig().
//	    App("Contoso", "Widget Pro", "2.0.0").
//	    Install().
//	    Interactive()
type SessionConfigBuilder struct {
	cfg SessionConfig
}

// NewSessionConfig creates a new SessionConfigBuilder with sensible defaults.
func NewSessionConfig() *SessionConfigBuilder {
	return &SessionConfigBuilder{
		cfg: SessionConfig{
			DeploymentType: DeployInstall,
			DeployMode:     DeployModeAuto,
		},
	}
}

// App sets the application identity fields.
func (b *SessionConfigBuilder) App(vendor, name, version string) *SessionConfigBuilder {
	b.cfg.AppVendor = vendor
	b.cfg.AppName = name
	b.cfg.AppVersion = version
	return b
}

// AppArch sets the application architecture.
func (b *SessionConfigBuilder) AppArch(arch string) *SessionConfigBuilder {
	b.cfg.AppArch = arch
	return b
}

// AppLang sets the application language.
func (b *SessionConfigBuilder) AppLang(lang string) *SessionConfigBuilder {
	b.cfg.AppLang = lang
	return b
}

// AppRevision sets the application revision.
func (b *SessionConfigBuilder) AppRevision(revision string) *SessionConfigBuilder {
	b.cfg.AppRevision = revision
	return b
}

// Install sets the deployment type to Install.
func (b *SessionConfigBuilder) Install() *SessionConfigBuilder {
	b.cfg.DeploymentType = DeployInstall
	return b
}

// Uninstall sets the deployment type to Uninstall.
func (b *SessionConfigBuilder) Uninstall() *SessionConfigBuilder {
	b.cfg.DeploymentType = DeployUninstall
	return b
}

// Repair sets the deployment type to Repair.
func (b *SessionConfigBuilder) Repair() *SessionConfigBuilder {
	b.cfg.DeploymentType = DeployRepair
	return b
}

// Interactive sets deploy mode to Interactive.
func (b *SessionConfigBuilder) Interactive() *SessionConfigBuilder {
	b.cfg.DeployMode = DeployModeInteractive
	return b
}

// Silent sets deploy mode to Silent.
func (b *SessionConfigBuilder) Silent() *SessionConfigBuilder {
	b.cfg.DeployMode = DeployModeSilent
	return b
}

// NonInteractive sets deploy mode to NonInteractive.
func (b *SessionConfigBuilder) NonInteractive() *SessionConfigBuilder {
	b.cfg.DeployMode = DeployModeNonInteractive
	return b
}

// Auto sets deploy mode to Auto.
func (b *SessionConfigBuilder) Auto() *SessionConfigBuilder {
	b.cfg.DeployMode = DeployModeAuto
	return b
}

// RequireAdmin marks that the deployment requires administrator privileges.
func (b *SessionConfigBuilder) RequireAdmin() *SessionConfigBuilder {
	b.cfg.RequireAdmin = true
	return b
}

// TerminalServer enables Terminal Server install mode.
func (b *SessionConfigBuilder) TerminalServer() *SessionConfigBuilder {
	b.cfg.TerminalServerMode = true
	return b
}

// DisableLogging disables file logging.
func (b *SessionConfigBuilder) DisableLogging() *SessionConfigBuilder {
	b.cfg.DisableLogging = true
	return b
}

// SuppressRebootPassThru suppresses reboot exit codes.
func (b *SessionConfigBuilder) SuppressRebootPassThru() *SessionConfigBuilder {
	b.cfg.SuppressRebootPassThru = true
	return b
}

// CloseProcesses specifies processes to close during deployment.
func (b *SessionConfigBuilder) CloseProcesses(processes ...ProcessDefinition) *SessionConfigBuilder {
	b.cfg.AppProcessesToClose = processes
	return b
}

// SuccessExitCodes sets the expected success exit codes.
func (b *SessionConfigBuilder) SuccessExitCodes(codes ...int) *SessionConfigBuilder {
	b.cfg.AppSuccessExitCodes = codes
	return b
}

// RebootExitCodes sets the exit codes indicating reboot is required.
func (b *SessionConfigBuilder) RebootExitCodes(codes ...int) *SessionConfigBuilder {
	b.cfg.AppRebootExitCodes = codes
	return b
}

// InstallName sets the install name.
func (b *SessionConfigBuilder) InstallName(name string) *SessionConfigBuilder {
	b.cfg.InstallName = name
	return b
}

// InstallTitle sets the install title.
func (b *SessionConfigBuilder) InstallTitle(title string) *SessionConfigBuilder {
	b.cfg.InstallTitle = title
	return b
}

// LogName sets the log file name.
func (b *SessionConfigBuilder) LogName(name string) *SessionConfigBuilder {
	b.cfg.LogName = name
	return b
}

// Dirs sets the script, files, and support directories.
func (b *SessionConfigBuilder) Dirs(scriptDir, filesDir, supportDir string) *SessionConfigBuilder {
	b.cfg.ScriptDirectory = scriptDir
	b.cfg.DirFiles = filesDir
	b.cfg.DirSupportFiles = supportDir
	return b
}

// DefaultMsi sets the default MSI file and optional transforms/patches.
func (b *SessionConfigBuilder) DefaultMsi(msiFile string, mstFile ...string) *SessionConfigBuilder {
	b.cfg.DefaultMsiFile = msiFile
	if len(mstFile) > 0 && mstFile[0] != "" {
		b.cfg.DefaultMstFile = mstFile[0]
	}
	return b
}

// ScriptInfo sets the script metadata.
func (b *SessionConfigBuilder) ScriptInfo(version, date, author string) *SessionConfigBuilder {
	b.cfg.AppScriptVersion = version
	b.cfg.AppScriptDate = date
	b.cfg.AppScriptAuthor = author
	return b
}

// Build returns the constructed SessionConfig.
func (b *SessionConfigBuilder) Build() SessionConfig {
	return b.cfg
}
