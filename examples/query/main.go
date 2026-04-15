//go:build windows

// Example: System queries and checks using go-psadt.
package main

import (
	"fmt"
	"log"

	"github.com/pedrostefanogv/go-psadt"
	"github.com/pedrostefanogv/go-psadt/types"
)

func main() {
	client, err := psadt.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Get environment info (no session needed)
	env, err := client.GetEnvironment()
	if err != nil {
		log.Fatalf("Failed to get environment: %v", err)
	}
	fmt.Printf("OS: %s %s\n", env.OS.Name, env.OS.Version)
	fmt.Printf("Architecture: %s\n", env.OS.Architecture)
	fmt.Printf("PowerShell: %s\n", env.PowerShell.PSVersion)

	// Open a session for additional queries
	session, err := client.OpenSession(types.SessionConfig{
		DeploymentType: types.DeployInstall,
		DeployMode:     types.DeployModeSilent,
		AppVendor:      "Query",
		AppName:        "System Check",
		AppVersion:     "1.0",
	})
	if err != nil {
		log.Fatalf("Failed to open session: %v", err)
	}
	defer session.Close(0)

	// Check admin privileges
	isAdmin, err := session.TestCallerIsAdmin()
	if err != nil {
		log.Printf("Admin check failed: %v", err)
	} else {
		fmt.Printf("Running as admin: %v\n", isAdmin)
	}

	// Check network
	hasNetwork, err := session.TestNetworkConnection()
	if err != nil {
		log.Printf("Network check failed: %v", err)
	} else {
		fmt.Printf("Network connected: %v\n", hasNetwork)
	}

	// Get free disk space
	freeSpace, err := session.GetFreeDiskSpace()
	if err != nil {
		log.Printf("Disk space check failed: %v", err)
	} else {
		fmt.Printf("Free disk space: %d MB\n", freeSpace)
	}

	// Get logged-on users
	users, err := session.GetLoggedOnUser()
	if err != nil {
		log.Printf("User query failed: %v", err)
	} else {
		fmt.Printf("Logged-on users: %d\n", len(users))
		for _, u := range users {
			fmt.Printf("  - %s (Session: %d)\n", u.NTAccount, u.SessionID)
		}
	}

	// Check pending reboot
	reboot, err := session.GetPendingReboot()
	if err != nil {
		log.Printf("Reboot check failed: %v", err)
	} else {
		fmt.Printf("Pending reboot: %v\n", reboot.IsSystemRebootPending)
	}

	// Check if service exists
	svcExists, err := session.TestServiceExists("Spooler")
	if err != nil {
		log.Printf("Service check failed: %v", err)
	} else {
		fmt.Printf("Spooler service exists: %v\n", svcExists)
	}

	fmt.Println("\nSystem query completed!")
}
