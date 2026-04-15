//go:build windows

// Example: Application uninstallation using go-psadt.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/peterondra/go-psadt"
	"github.com/peterondra/go-psadt/types"
)

func main() {
	client, err := psadt.NewClient(
		psadt.WithTimeout(5 * time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	session, err := client.OpenSession(types.SessionConfig{
		DeploymentType: types.DeployUninstall,
		DeployMode:     types.DeployModeInteractive,
		AppVendor:      "Contoso",
		AppName:        "Widget Pro",
		AppVersion:     "2.0.0",
	})
	if err != nil {
		log.Fatalf("Failed to open session: %v", err)
	}
	defer session.Close(0)

	// Show welcome
	if err := session.ShowInstallationWelcome(types.WelcomeOptions{
		CloseProcesses: []types.ProcessDefinition{{Name: "widget"}, {Name: "widgethelper"}},
	}); err != nil {
		log.Fatalf("Welcome failed: %v", err)
	}

	// Search for installed application
	apps, err := session.GetApplication(types.GetApplicationOptions{
		Name: []string{"Widget Pro"},
	})
	if err != nil {
		log.Fatalf("Failed to search applications: %v", err)
	}
	fmt.Printf("Found %d matching application(s)\n", len(apps))

	// Uninstall
	if err := session.UninstallApplication(types.UninstallApplicationOptions{
		Name: []string{"Widget Pro"},
	}); err != nil {
		log.Fatalf("Uninstallation failed: %v", err)
	}

	// Clean up registry
	if err := session.RemoveRegistryKey(types.RemoveRegistryKeyOptions{
		Key: `HKLM\SOFTWARE\Contoso\WidgetPro`,
	}); err != nil {
		log.Printf("Registry cleanup failed: %v", err)
	}

	fmt.Println("Uninstallation completed successfully!")
}
