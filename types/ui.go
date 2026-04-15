//go:build windows

package types

// WelcomeOptions options for Show-ADTInstallationWelcome.
type WelcomeOptions struct {
	CloseProcesses               []ProcessDefinition `ps:"CloseProcesses"`
	DeferCloseProcesses          []ProcessDefinition `ps:"DeferCloseProcesses"`
	CloseProcessesCountdown      int                 `ps:"CloseProcessesCountdown"`
	ForceCloseProcessesCountdown int                 `ps:"ForceCloseProcessesCountdown"`
	PromptToSave                 bool                `ps:"PromptToSave,switch"`
	PersistPrompt                bool                `ps:"PersistPrompt,switch"`
	BlockExecution               bool                `ps:"BlockExecution,switch"`
	AllowDefer                   bool                `ps:"AllowDefer,switch"`
	AllowDeferCloseProcesses     bool                `ps:"AllowDeferCloseProcesses,switch"`
	DeferTimes                   int                 `ps:"DeferTimes"`
	DeferDays                    int                 `ps:"DeferDays"`
	DeferDeadline                string              `ps:"DeferDeadline"`
	CheckDiskSpace               bool                `ps:"CheckDiskSpace,switch"`
	RequiredDiskSpace            int                 `ps:"RequiredDiskSpace"`
	MinimizeWindows              bool                `ps:"MinimizeWindows,switch"`
	TopMost                      bool                `ps:"TopMost,switch"`
	ForceCountdown               int                 `ps:"ForceCountdown"`
	CustomText                   bool                `ps:"CustomText,switch"`
	NotTopMost                   bool                `ps:"NotTopMost,switch"`
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
	NoWait           bool             `ps:"NoWait,switch"`
	PersistPrompt    bool             `ps:"PersistPrompt,switch"`
	MinimizeWindows  bool             `ps:"MinimizeWindows,switch"`
	Timeout          int              `ps:"Timeout"`
	ExitOnTimeout    bool             `ps:"ExitOnTimeout,switch"`
	TopMost          bool             `ps:"TopMost,switch"`
	NotTopMost       bool             `ps:"NotTopMost,switch"`
}

// PromptResult result of ShowInstallationPrompt.
type PromptResult struct {
	ButtonClicked string `json:"ButtonClicked"`
	InputText     string `json:"InputText,omitempty"`
}

// ProgressOptions options for Show-ADTInstallationProgress.
type ProgressOptions struct {
	StatusMessage       string `ps:"StatusMessage"`
	StatusMessageDetail string `ps:"StatusMessageDetail"`
	TopMost             bool   `ps:"TopMost,switch"`
	NotTopMost          bool   `ps:"NotTopMost,switch"`
	InPlace             bool   `ps:"InPlace,switch"`
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
	Title         string           `ps:"Title"`
	Text          string           `ps:"Text"`
	Buttons       DialogBoxButtons `ps:"Buttons"`
	DefaultButton string           `ps:"DefaultButton"`
	Icon          DialogSystemIcon `ps:"Icon"`
	Timeout       int              `ps:"Timeout"`
	TopMost       bool             `ps:"TopMost,switch"`
	NotTopMost    bool             `ps:"NotTopMost,switch"`
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
