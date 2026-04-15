//go:build windows

// Example: Typical MSI installation using go-psadt.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pedrostefanogv/go-psadt"
	"github.com/pedrostefanogv/go-psadt/types"
)

func main() {
	// Create a PSADT client
	client, err := psadt.NewClient(
		psadt.WithTimeout(10 * time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Open a deployment session
	session, err := client.OpenSession(types.SessionConfig{
		DeploymentType: types.DeployInstall,
		DeployMode:     types.DeployModeInteractive,
		AppVendor:      "Contoso",
		AppName:        "Widget Pro",
		AppVersion:     "2.0.0",
	})
	if err != nil {
		log.Fatalf("Failed to open session: %v", err)
	}
	defer session.Close(0)

	// Show welcome prompt — close running apps
	if err := session.ShowInstallationWelcome(types.WelcomeOptions{
		CloseProcesses:          []types.ProcessDefinition{{Name: "widget"}, {Name: "widgethelper"}},
		CloseProcessesCountdown: 300,
		CheckDiskSpace:          true,
	}); err != nil {
		log.Fatalf("Welcome failed: %v", err)
	}

	// Show progress
	if err := session.ShowInstallationProgress(types.ProgressOptions{
		StatusMessage: "Installing Widget Pro 2.0...",
	}); err != nil {
		log.Printf("Progress display failed: %v", err)
	}

	// Run MSI installer
	result, err := session.StartMsiProcess(types.MsiProcessOptions{
		Action:   types.MsiInstall,
		FilePath: "WidgetPro.msi",
		PassThru: true,
	})
	if err != nil {
		log.Fatalf("MSI installation failed: %v", err)
	}
	fmt.Printf("MSI exit code: %d\n", result.ExitCode)

	// Set a registry key for configuration
	if err := session.SetRegistryKey(types.SetRegistryKeyOptions{
		Key:   `HKLM\SOFTWARE\Contoso\WidgetPro`,
		Name:  "Version",
		Value: "2.0.0",
		Type:  types.RegString,
	}); err != nil {
		log.Printf("Registry write failed: %v", err)
	}

	// Close progress
	if err := session.CloseInstallationProgress(); err != nil {
		log.Printf("Close progress failed: %v", err)
	}

	fmt.Println("Installation completed successfully!")
}
