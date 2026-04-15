//go:build windows

package psadt

import (
	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// ShowInstallationWelcome displays the installation welcome dialog.
func (s *Session) ShowInstallationWelcome(opts types.WelcomeOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Show-ADTInstallationWelcome", opts)
	return s.executeVoid(ctx, cmd)
}

// ShowInstallationPrompt displays an installation prompt dialog and returns the user's response.
func (s *Session) ShowInstallationPrompt(opts types.PromptOptions) (*types.PromptResult, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Show-ADTInstallationPrompt", opts)
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
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Show-ADTInstallationProgress", opts)
	return s.executeVoid(ctx, cmd)
}

// CloseInstallationProgress closes the installation progress dialog.
func (s *Session) CloseInstallationProgress() error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	return s.executeVoid(ctx, "Close-ADTInstallationProgress")
}

// ShowInstallationRestartPrompt displays a restart prompt dialog.
func (s *Session) ShowInstallationRestartPrompt(opts types.RestartPromptOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Show-ADTInstallationRestartPrompt", opts)
	return s.executeVoid(ctx, cmd)
}

// ShowDialogBox displays a standard Windows dialog box and returns the result.
func (s *Session) ShowDialogBox(opts types.DialogBoxOptions) (types.DialogBoxResult, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
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

// ShowBalloonTip displays a balloon tip / toast notification.
func (s *Session) ShowBalloonTip(opts types.BalloonTipOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Show-ADTBalloonTip", opts)
	return s.executeVoid(ctx, cmd)
}

// ShowHelpConsole displays the PSADT help console.
func (s *Session) ShowHelpConsole() error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	return s.executeVoid(ctx, "Show-ADTHelpConsole")
}
