//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// ShowInstallationWelcome displays the installation welcome dialog.
func (s *Session) ShowInstallationWelcome(opts types.WelcomeOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	normalizedOpts, err := normalizeWelcomeOptions(opts)
	if err != nil {
		return err
	}
	cmd := cmdbuilder.Build("Show-ADTInstallationWelcome", normalizedOpts)
	return s.executeVoid(ctx, cmd)
}

// ShowInstallationPrompt displays an installation prompt dialog and returns the user's response.
func (s *Session) ShowInstallationPrompt(opts types.PromptOptions) (*types.PromptResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	normalizedOpts, err := normalizePromptOptions(opts)
	if err != nil {
		return nil, err
	}
	cmd := cmdbuilder.Build("Show-ADTInstallationPrompt", normalizedOpts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result types.PromptResult
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ShowInstallationProgress displays or updates the installation progress dialog.
func (s *Session) ShowInstallationProgress(opts types.ProgressOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	normalizedOpts, err := normalizeProgressOptions(opts)
	if err != nil {
		return err
	}
	cmd := cmdbuilder.Build("Show-ADTInstallationProgress", normalizedOpts)
	return s.executeVoid(ctx, cmd)
}

// CloseInstallationProgress closes the installation progress dialog.
func (s *Session) CloseInstallationProgress() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Close-ADTInstallationProgress")
}

// ShowInstallationRestartPrompt displays a restart prompt dialog.
func (s *Session) ShowInstallationRestartPrompt(opts types.RestartPromptOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Show-ADTInstallationRestartPrompt", opts)
	return s.executeVoid(ctx, cmd)
}

// ShowDialogBox displays a standard Windows dialog box and returns the result.
func (s *Session) ShowDialogBox(opts types.DialogBoxOptions) (types.DialogBoxResult, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	normalizedIcon, err := normalizeDialogBoxIcon(opts.Icon)
	if err != nil {
		return "", err
	}
	opts.Icon = normalizedIcon
	cmd := cmdbuilder.Build("Show-ADTDialogBox", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	str, err := parser.ParseString(data)
	if err != nil {
		return "", err
	}
	return types.DialogBoxResult(str), nil
}

func normalizeWelcomeOptions(opts types.WelcomeOptions) (types.WelcomeOptions, error) {
	if opts.Silent && opts.AllowDefer {
		return opts, fmt.Errorf("welcome options Silent and AllowDefer cannot be used together")
	}
	if opts.Silent && opts.AllowDeferCloseProcesses {
		return opts, fmt.Errorf("welcome options Silent and AllowDeferCloseProcesses cannot be used together")
	}
	return opts, nil
}

func normalizePromptOptions(opts types.PromptOptions) (types.PromptOptions, error) {
	if opts.TopMost && opts.NotTopMost {
		return opts, fmt.Errorf("prompt options TopMost and NotTopMost cannot both be true")
	}
	if opts.ExitOnTimeout && opts.NoExitOnTimeout {
		return opts, fmt.Errorf("prompt options ExitOnTimeout and NoExitOnTimeout cannot both be true")
	}
	if opts.RequestInput && len(opts.ListItems) > 0 {
		return opts, fmt.Errorf("prompt options RequestInput and ListItems cannot be used together")
	}
	if opts.SecureInput && !opts.RequestInput {
		return opts, fmt.Errorf("prompt option SecureInput requires RequestInput")
	}
	if len(opts.ListItems) > 0 && opts.DefaultIndex >= len(opts.ListItems) {
		return opts, fmt.Errorf("prompt option DefaultIndex=%d is out of range for %d list items", opts.DefaultIndex, len(opts.ListItems))
	}
	return opts, nil
}

func normalizeDialogBoxIcon(icon types.DialogSystemIcon) (types.DialogSystemIcon, error) {
	switch icon {
	case "", types.DialogSystemIcon("None"), types.DialogSystemIcon("Stop"), types.IconQuestion, types.IconExclamation, types.IconInformation:
		return icon, nil
	case types.IconError, types.IconHand:
		return types.DialogSystemIcon("Stop"), nil
	case types.IconWarning:
		return types.IconExclamation, nil
	case types.IconAsterisk:
		return types.IconInformation, nil
	default:
		return "", fmt.Errorf("unsupported dialog box icon %q; use None, Stop, Question, Exclamation, or Information", icon)
	}
}

func normalizeProgressOptions(opts types.ProgressOptions) (types.ProgressOptions, error) {
	if opts.TopMost && opts.NotTopMost {
		return opts, fmt.Errorf("progress options TopMost and NotTopMost cannot both be true")
	}
	if opts.InPlace {
		return opts, fmt.Errorf("progress option InPlace is not supported by Show-ADTInstallationProgress in current PSADT")
	}
	return opts, nil
}

// ShowBalloonTip displays a balloon tip / toast notification.
func (s *Session) ShowBalloonTip(opts types.BalloonTipOptions) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := cmdbuilder.Build("Show-ADTBalloonTip", opts)
	return s.executeVoid(ctx, cmd)
}

// ShowHelpConsole displays the PSADT help console.
func (s *Session) ShowHelpConsole() error {
	ctx, cancel := s.getContext()
	defer cancel()
	return s.executeVoid(ctx, "Show-ADTHelpConsole")
}
