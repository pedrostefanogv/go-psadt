//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/types"
)

// UpdateNotification describes a high-level "app will be updated" user
// notification — the typical RMM scenario:
//
//	"O Firefox será atualizado em 30 segundos; salve seu trabalho."
//
// It maps onto Show-ADTInstallationWelcome with a close-processes countdown,
// optional save prompt, deferral and process blocking. Zero-value fields get
// sensible defaults set by Session.NotifyUpdate.
type UpdateNotification struct {
	// AppName is shown in the dialog title/subtitle when Title/Subtitle are
	// empty (e.g. "Firefox"). Required.
	AppName string

	// Processes to close for the deployment to proceed. Required in practice;
	// NotifyUpdate falls back to a process named after AppName.
	Processes []types.ProcessDefinition

	// CountdownSeconds is the visible countdown before the processes are
	// closed automatically. Default: 30 ("...será atualizado em 30 segundos").
	CountdownSeconds int

	// PromptToSave asks applications to save the user's work before closing.
	// Default: true.
	PromptToSave *bool

	// AllowDefer lets the user postpone the update.
	AllowDefer bool

	// DeferTimes limits how many times the user may defer. Used with AllowDefer.
	DeferTimes int

	// DeferDays is an alternative to DeferTimes (in days from first prompt).
	DeferDays float64

	// DeferDeadline is an absolute ISO-8601 deadline for deferral.
	DeferDeadline string

	// BlockExecution prevents the user from launching the processes while
	// the deployment runs. Default: false.
	BlockExecution bool

	// Title and Subtitle override the strings from strings.psd1.
	Title    string
	Subtitle string
}

// NotifyUpdate shows the countdown/defer welcome dialog for an application
// update and returns when the user accepts, defers, or the countdown
// expires. After it returns without error, the deployment may proceed.
//
// This is the convenience wrapper for the "Firefox vai ser atualizado em 30
// segundos, salve seu trabalho" scenario:
//
//	err := session.NotifyUpdate(psadt.UpdateNotification{
//	    AppName:   "Firefox",
//	    CountdownSeconds: 30,
//	    AllowDefer:       true,
//	    DeferTimes:       3,
//	})
func (s *Session) NotifyUpdate(n UpdateNotification) error {
	if n.AppName == "" {
		return fmt.Errorf("NotifyUpdate: AppName is required")
	}

	promptToSave := true
	if n.PromptToSave != nil {
		promptToSave = *n.PromptToSave
	}

	countdown := n.CountdownSeconds
	if countdown <= 0 {
		countdown = 30
	}

	processes := n.Processes
	if len(processes) == 0 {
		processes = []types.ProcessDefinition{{Name: n.AppName, Description: n.AppName}}
	}

	title := n.Title
	if title == "" {
		title = n.AppName
	}
	subtitle := n.Subtitle
	if subtitle == "" {
		subtitle = "Atualização de Software"
	}

	opts := types.WelcomeOptions{
		Title:                   title,
		Subtitle:                subtitle,
		CloseProcesses:          processes,
		CloseProcessesCountdown: countdown,
		PromptToSave:            promptToSave,
		BlockExecution:          n.BlockExecution,
		AllowDefer:              n.AllowDefer,
		DeferTimes:              n.DeferTimes,
		DeferDays:               n.DeferDays,
		DeferDeadline:           n.DeferDeadline,
	}
	return s.ShowInstallationWelcome(opts)
}

// NotifyProgress shows (or updates) the branded installation progress dialog
// with a status message for the given application.
func (s *Session) NotifyProgress(appName, statusMessage string) error {
	if statusMessage == "" {
		statusMessage = fmt.Sprintf("Installing %s...", appName)
	}
	return s.ShowInstallationProgress(types.ProgressOptions{
		StatusMessage: statusMessage,
	})
}

// NotifyInfo shows a toast/balloon notification with informational icon —
// a non-blocking way to tell the user something happened.
func (s *Session) NotifyInfo(title, message string) error {
	return s.ShowBalloonTip(types.BalloonTipOptions{
		BalloonTipTitle: title,
		BalloonTipText:  message,
		BalloonTipIcon:  types.BalloonInfo,
	})
}
