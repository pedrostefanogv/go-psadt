//go:build windows

// Example: UI dialogs and prompts using go-psadt.
package main

import (
	"fmt"
	"log"

	"github.com/peterondra/go-psadt"
	"github.com/peterondra/go-psadt/types"
)

func main() {
	client, err := psadt.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	session, err := client.OpenSession(types.SessionConfig{
		DeploymentType: types.DeployInstall,
		DeployMode:     types.DeployModeInteractive,
		AppVendor:      "Demo",
		AppName:        "Dialog Demo",
		AppVersion:     "1.0",
	})
	if err != nil {
		log.Fatalf("Failed to open session: %v", err)
	}
	defer session.Close(0)

	// Show a dialog box
	result, err := session.ShowDialogBox(types.DialogBoxOptions{
		Title:   "Confirmation",
		Text:    "Do you want to proceed with the installation?",
		Buttons: types.ButtonsYesNo,
		Icon:    types.IconQuestion,
	})
	if err != nil {
		log.Fatalf("Dialog failed: %v", err)
	}
	fmt.Printf("User chose: %s\n", result)

	// Show a balloon tip notification
	if err := session.ShowBalloonTip(types.BalloonTipOptions{
		BalloonTipText:  "Installation is starting...",
		BalloonTipTitle: "Widget Pro Setup",
		BalloonTipIcon:  types.BalloonInfo,
	}); err != nil {
		log.Printf("Balloon tip failed: %v", err)
	}

	// Show installation prompt
	promptResult, err := session.ShowInstallationPrompt(types.PromptOptions{
		Title:            "Ready to Install",
		Message:          "Click OK to begin or Cancel to abort.",
		MessageAlignment: types.AlignCenter,
	})
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}
	fmt.Printf("Prompt result: %+v\n", promptResult)

	fmt.Println("Dialog demo completed!")
}
