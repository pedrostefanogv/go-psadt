//go:build windows

package types

// WelcomeOptions options for Show-ADTInstallationWelcome.
type WelcomeOptions struct {
	Title                        string              `ps:"Title"`
	Subtitle                     string              `ps:"Subtitle"`
	CloseProcesses               []ProcessDefinition `ps:"CloseProcesses"`
	Silent                       bool                `ps:"Silent,switch"`
	HideCloseButton              bool                `ps:"HideCloseButton,switch"`
	AllowDefer                   bool                `ps:"AllowDefer,switch"`
	AllowDeferCloseProcesses     bool                `ps:"AllowDeferCloseProcesses,switch"`
	DeferTimes                   int                 `ps:"DeferTimes"`
	DeferDays                    float64             `ps:"DeferDays"`
	DeferDeadline                string              `ps:"DeferDeadline"`
	DeferRunInterval             string              `ps:"DeferRunInterval"`
	CloseProcessesCountdown      int                 `ps:"CloseProcessesCountdown"`
	ForceCloseProcessesCountdown int                 `ps:"ForceCloseProcessesCountdown"`
	ForceCountdown               int                 `ps:"ForceCountdown"`
	WindowLocation               DialogPosition      `ps:"WindowLocation"`
	BlockExecution               bool                `ps:"BlockExecution,switch"`
	PromptToSave                 bool                `ps:"PromptToSave,switch"`
	PersistPrompt                bool                `ps:"PersistPrompt,switch"`
	ContinueOnProcessClosure     bool                `ps:"ContinueOnProcessClosure,switch"`
	MinimizeWindows              bool                `ps:"MinimizeWindows,switch"`
	NotTopMost                   bool                `ps:"NotTopMost,switch"`
	AllowMove                    bool                `ps:"AllowMove,switch"`
	AllowMinimize                bool                `ps:"AllowMinimize,switch"`
	CustomText                   bool                `ps:"CustomText,switch"`
	CheckDiskSpace               bool                `ps:"CheckDiskSpace,switch"`
	RequiredDiskSpace            int                 `ps:"RequiredDiskSpace"`
	PassThru                     bool                `ps:"PassThru,switch"`
	TopMost                      bool                `ps:"-"` // Deprecated: welcome is topmost by default unless NotTopMost is set.
	DeferCloseProcesses          []ProcessDefinition `ps:"-"` // Deprecated: use AllowDeferCloseProcesses switch instead.
}

// PromptOptions options for Show-ADTInstallationPrompt.
type PromptOptions struct {
	Title            string           `ps:"Title"`
	Message          string           `ps:"Message"`
	MessageAlignment MessageAlignment `ps:"MessageAlignment"`
	ButtonLeftText   string           `ps:"ButtonLeftText"`
	ButtonRightText  string           `ps:"ButtonRightText"`
	ButtonMiddleText string           `ps:"ButtonMiddleText"`
	Icon             DialogSystemIcon `ps:"Icon"`
	RequestInput     bool             `ps:"RequestInput,switch"`
	DefaultValue     string           `ps:"DefaultValue"`
	SecureInput      bool             `ps:"SecureInput,switch"`
	ListItems        []string         `ps:"ListItems"`
	DefaultIndex     int              `ps:"DefaultIndex"`
	WindowLocation   DialogPosition   `ps:"WindowLocation"`
	NoWait           bool             `ps:"NoWait,switch"`
	PersistPrompt    bool             `ps:"PersistPrompt,switch"`
	MinimizeWindows  bool             `ps:"MinimizeWindows,switch"`
	Timeout          int              `ps:"Timeout"`
	NoExitOnTimeout  bool             `ps:"NoExitOnTimeout,switch"`
	AllowMove        bool             `ps:"AllowMove,switch"`
	Force            bool             `ps:"Force,switch"`
	NotTopMost       bool             `ps:"NotTopMost,switch"`
	ExitOnTimeout    bool             `ps:"-"` // Deprecated compatibility field; current PSADT uses NoExitOnTimeout.
	TopMost          bool             `ps:"-"` // Deprecated compatibility field; prompt is topmost by default unless NotTopMost is set.
}

// PromptResult result of ShowInstallationPrompt.
type PromptResult struct {
	ButtonClicked string `json:"ButtonClicked"`
	InputText     string `json:"InputText,omitempty"`
}

// ProgressOptions options for Show-ADTInstallationProgress.
type ProgressOptions struct {
	StatusMessage       string           `ps:"StatusMessage"`
	StatusMessageDetail string           `ps:"StatusMessageDetail"`
	StatusBarPercentage float64          `ps:"StatusBarPercentage"`
	MessageAlignment    MessageAlignment `ps:"MessageAlignment"`
	WindowLocation      DialogPosition   `ps:"WindowLocation"`
	AllowMove           bool             `ps:"AllowMove,switch"`
	NotTopMost          bool             `ps:"NotTopMost,switch"`
	TopMost             bool             `ps:"-"` // Deprecated compatibility field; progress is topmost by default unless NotTopMost is set.
	InPlace             bool             `ps:"-"` // Deprecated compatibility field; unsupported by current PSADT public cmdlet.
}

// RestartPromptOptions options for Show-ADTInstallationRestartPrompt.
type RestartPromptOptions struct {
	CountdownSeconds       int  `ps:"CountdownSeconds"`
	CountdownNoHideSeconds int  `ps:"CountdownNoHideSeconds"`
	SilentCountdownSeconds int  `ps:"SilentCountdownSeconds"`
	SilentRestart          bool `ps:"SilentRestart,switch"`
	NoCountdown            bool `ps:"NoCountdown,switch"`
	NotTopMost             bool `ps:"NotTopMost,switch"`
	TopMost                bool `ps:"TopMost,switch"`
}

// DialogBoxOptions options for Show-ADTDialogBox.
type DialogBoxOptions struct {
	Title         string                 `ps:"Title"`
	Text          string                 `ps:"Text"`
	Buttons       DialogBoxButtons       `ps:"Buttons"`
	DefaultButton DialogBoxDefaultButton `ps:"DefaultButton"`
	Icon          DialogSystemIcon       `ps:"Icon"` // Supported values: None, Stop, Question, Exclamation, Information.
	NoWait        bool                   `ps:"NoWait,switch"`
	ExitOnTimeout bool                   `ps:"ExitOnTimeout,switch"`
	NotTopMost    bool                   `ps:"NotTopMost,switch"`
	Force         bool                   `ps:"Force,switch"`
	Timeout       int                    `ps:"Timeout"`
	TopMost       bool                   `ps:"-"` // Deprecated: dialog is topmost by default unless NotTopMost is set.
}

// DialogBoxResult result of ShowDialogBox.
type DialogBoxResult string

// BalloonTipOptions options for Show-ADTBalloonTip.
type BalloonTipOptions struct {
	BalloonTipTitle string         `ps:"BalloonTipTitle"`
	BalloonTipText  string         `ps:"BalloonTipText"`
	BalloonTipIcon  BalloonTipIcon `ps:"BalloonTipIcon"`
	BalloonTipTime  int            `ps:"BalloonTipTime"`
	NoWait          bool           `ps:"NoWait,switch"`
}
