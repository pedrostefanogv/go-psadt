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

	// Verify registry with typed access
	installedVer, err := session.GetRegistryKeyString(`HKLM\SOFTWARE\Contoso\WidgetPro`, "Version")
	if err != nil {
		log.Printf("Registry verify failed: %v", err)
	} else {
		fmt.Printf("Verified installed version: %s\n", installedVer)
	}

	// Generate log file name for this deployment
	logName, err := session.NewLogFileName(types.LogFileNameOptions{
		AppName:    "WidgetPro",
		AppVersion: "2.0.0",
		UseDate:    true,
	})
	if err != nil {
		log.Printf("Log name generation failed: %v", err)
	} else {
		fmt.Printf("Deployment log: %s\n", logName)
	}

	// Structured logging
	if err := session.WriteLogEntryInfo("Widget Pro 2.0.0 installed successfully", "Installer"); err != nil {
		log.Printf("Log write failed: %v", err)
	}

	// Update environment provider after installation
	if err := session.UpdateEnvironmentPsProvider(); err != nil {
		log.Printf("Environment update failed: %v", err)
	}

	// Close progress
	if err := session.CloseInstallationProgress(); err != nil {
		log.Printf("Close progress failed: %v", err)
	}

	fmt.Println("Installation completed successfully!")
}
