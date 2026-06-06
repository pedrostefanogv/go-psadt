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

	// Get PowerShell process path
	psInfo, err := session.GetPowerShellProcessPath()
	if err != nil {
		log.Printf("PowerShell path query failed: %v", err)
	} else {
		fmt.Printf("PowerShell path: %s (%s)\n", psInfo.Path, psInfo.Version)
	}

	// Test module initialization
	modInitialized, err := session.TestModuleInitialized()
	if err != nil {
		log.Printf("Module check failed: %v", err)
	} else {
		fmt.Printf("PSADT module initialized: %v\n", modInitialized)
	}

	// Get session properties
	sessProps, err := session.GetSession()
	if err != nil {
		log.Printf("Session query failed: %v", err)
	} else {
		fmt.Printf("Session: %s/%s v%s (log: %s)\n", sessProps.AppName, sessProps.DeploymentType, sessProps.AppVersion, sessProps.LogName)
	}

	// Get environment table (flat key-value map)
	envTable, err := session.GetEnvironmentTable()
	if err != nil {
		log.Printf("Environment table failed: %v", err)
	} else {
		fmt.Printf("Environment variables: %d\n", len(envTable))
		if name, ok := envTable["envComputerName"]; ok {
			fmt.Printf("  ComputerName: %s\n", name)
		}
	}

	// Generate a log file name
	logFileName, err := session.NewLogFileName(types.LogFileNameOptions{
		AppName:    "SystemCheck",
		AppVersion: "1.0",
		UseDate:    true,
	})
	if err != nil {
		log.Printf("Log file name generation failed: %v", err)
	} else {
		fmt.Printf("Log file name: %s\n", logFileName)
	}

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

	// Typed registry query examples
	regVer, err := session.GetRegistryKeyString(`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "ProductName")
	if err != nil {
		log.Printf("Registry string read failed: %v", err)
	} else {
		fmt.Printf("Registry (String): ProductName = %s\n", regVer)
	}

	// Convert registry path (long to short form)
	regPath, err := session.ConvertRegistryPath(`HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft`, true)
	if err != nil {
		log.Printf("Registry path conversion failed: %v", err)
	} else {
		fmt.Printf("Registry path: %s -> %s\n", regPath.RegistryPath, regPath.RegistryKeyPath)
	}

	// Structured logging example
	if err := session.WriteLogEntryInfo("System query completed successfully", "QueryExample"); err != nil {
		log.Printf("Log write failed: %v", err)
	}

	fmt.Println("\nSystem query completed!")
}
