//go:build windows

package types

// DeploymentType represents the deployment type
type DeploymentType string

const (
	DeployInstall   DeploymentType = "Install"
	DeployUninstall DeploymentType = "Uninstall"
	DeployRepair    DeploymentType = "Repair"
)

// DeployMode represents the interaction mode
type DeployMode string

const (
	DeployModeAuto           DeployMode = "Auto"
	DeployModeInteractive    DeployMode = "Interactive"
	DeployModeNonInteractive DeployMode = "NonInteractive"
	DeployModeSilent         DeployMode = "Silent"
)

// DialogStyle represents the visual style of dialogs
type DialogStyle string

const (
	DialogStyleFluent  DialogStyle = "Fluent"
	DialogStyleClassic DialogStyle = "Classic"
)

// DialogPosition represents dialog screen position
type DialogPosition string

const (
	DialogPositionDefault     DialogPosition = "Default"
	DialogPositionTopLeft     DialogPosition = "TopLeft"
	DialogPositionTop         DialogPosition = "Top"
	DialogPositionTopRight    DialogPosition = "TopRight"
	DialogPositionTopCenter   DialogPosition = "TopCenter"
	DialogPositionCenter      DialogPosition = "Center"
	DialogPositionBottomLeft  DialogPosition = "BottomLeft"
	DialogPositionBottom      DialogPosition = "Bottom"
	DialogPositionBottomRight DialogPosition = "BottomRight"
)

// DialogSystemIcon represents dialog icons
type DialogSystemIcon string

const (
	IconApplication DialogSystemIcon = "Application"
	IconAsterisk    DialogSystemIcon = "Asterisk"
	IconError       DialogSystemIcon = "Error"
	IconExclamation DialogSystemIcon = "Exclamation"
	IconHand        DialogSystemIcon = "Hand"
	IconInformation DialogSystemIcon = "Information"
	IconQuestion    DialogSystemIcon = "Question"
	IconShield      DialogSystemIcon = "Shield"
	IconWarning     DialogSystemIcon = "Warning"
	IconWinLogo     DialogSystemIcon = "WinLogo"
)

// DialogBoxButtons represents standard Windows dialog buttons
type DialogBoxButtons string

const (
	ButtonsOk                DialogBoxButtons = "Ok"
	ButtonsOkCancel          DialogBoxButtons = "OkCancel"
	ButtonsAbortRetryIgnore  DialogBoxButtons = "AbortRetryIgnore"
	ButtonsYesNoCancel       DialogBoxButtons = "YesNoCancel"
	ButtonsYesNo             DialogBoxButtons = "YesNo"
	ButtonsRetryCancel       DialogBoxButtons = "RetryCancel"
	ButtonsCancelTryContinue DialogBoxButtons = "CancelTryContinue"
)

// RegistryValueKind represents Windows registry value types
type RegistryValueKind string

const (
	RegString       RegistryValueKind = "String"
	RegExpandString RegistryValueKind = "ExpandString"
	RegBinary       RegistryValueKind = "Binary"
	RegDWord        RegistryValueKind = "DWord"
	RegMultiString  RegistryValueKind = "MultiString"
	RegQWord        RegistryValueKind = "QWord"
)

// ProcessWindowStyle represents process window styles
type ProcessWindowStyle string

const (
	WindowNormal    ProcessWindowStyle = "Normal"
	WindowHidden    ProcessWindowStyle = "Hidden"
	WindowMaximized ProcessWindowStyle = "Maximized"
	WindowMinimized ProcessWindowStyle = "Minimized"
)

// EnvironmentVariableTarget represents the scope of an environment variable
type EnvironmentVariableTarget string

const (
	EnvTargetProcess EnvironmentVariableTarget = "Process"
	EnvTargetUser    EnvironmentVariableTarget = "User"
	EnvTargetMachine EnvironmentVariableTarget = "Machine"
)

// ServiceStartMode represents service startup modes
type ServiceStartMode string

const (
	ServiceAutomatic             ServiceStartMode = "Automatic"
	ServiceManual                ServiceStartMode = "Manual"
	ServiceDisabled              ServiceStartMode = "Disabled"
	ServiceAutomaticDelayedStart ServiceStartMode = "AutomaticDelayedStart"
)

// MsiAction represents MSI actions
type MsiAction string

const (
	MsiInstall     MsiAction = "Install"
	MsiUninstall   MsiAction = "Uninstall"
	MsiPatch       MsiAction = "Patch"
	MsiRepair      MsiAction = "Repair"
	MsiActiveSetup MsiAction = "ActiveSetup"
)

// NameMatch represents the match type for application search
type NameMatch string

const (
	MatchContains NameMatch = "Contains"
	MatchExact    NameMatch = "Exact"
	MatchWildcard NameMatch = "Wildcard"
	MatchRegex    NameMatch = "Regex"
)

// ApplicationType filters by installer type
type ApplicationType string

const (
	AppTypeAll ApplicationType = "All"
	AppTypeMSI ApplicationType = "MSI"
	AppTypeEXE ApplicationType = "EXE"
)

// LogSeverity represents log severity
type LogSeverity int

const (
	LogInfo    LogSeverity = 1
	LogWarning LogSeverity = 2
	LogError   LogSeverity = 3
)

// BalloonTipIcon represents balloon/toast icons
type BalloonTipIcon string

const (
	BalloonNone    BalloonTipIcon = "None"
	BalloonInfo    BalloonTipIcon = "Info"
	BalloonWarning BalloonTipIcon = "Warning"
	BalloonError   BalloonTipIcon = "Error"
)

// MessageAlignment for dialogs
type MessageAlignment string

const (
	AlignLeft   MessageAlignment = "Left"
	AlignCenter MessageAlignment = "Center"
	AlignRight  MessageAlignment = "Right"
)

// DialogMessageAlignment alias for PSADT parameter compatibility
type DialogMessageAlignment = MessageAlignment
