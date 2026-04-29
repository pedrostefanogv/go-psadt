//go:build windows

// Example: RMM Agent integration using go-psadt.
// Demonstrates how an RMM agent would use the library to:
// - Check system state before deployment
// - Respect user activity (Focus Assist, presentations, OOBE)
// - Show notifications and interactive prompts
// - Install/uninstall software with progress tracking
// - Handle runner reconnection
// - Report results back to the RMM platform
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pedrostefanogv/go-psadt"
	"github.com/pedrostefanogv/go-psadt/types"
)

// RMMTask represents a deployment task received from the RMM platform.
type RMMTask struct {
	TaskID         string   `json:"task_id"`
	Action         string   `json:"action"` // install, uninstall, notify, query
	AppName        string   `json:"app_name"`
	AppVendor      string   `json:"app_vendor"`
	AppVersion     string   `json:"app_version"`
	InstallerPath  string   `json:"installer_path"`
	InstallerType  string   `json:"installer_type"` // msi, exe
	Silent         bool     `json:"silent"`
	Deferrable     bool     `json:"deferrable"`
	CloseProcesses []string `json:"close_processes"`
	TimeoutMinutes int      `json:"timeout_minutes"`
	Message        string   `json:"message"` // for notifications
}

// RMMResult represents the result reported back to the RMM platform.
type RMMResult struct {
	TaskID   string   `json:"task_id"`
	Success  bool     `json:"success"`
	ExitCode int      `json:"exit_code,omitempty"`
	Message  string   `json:"message"`
	Logs     []string `json:"logs,omitempty"`
}

func main() {
	fmt.Println("=== go-psadt RMM Agent Example ===")

	// Simulate receiving a task from RMM platform
	task := RMMTask{
		TaskID:         "rmm-task-12345",
		Action:         "install",
		AppName:        "Example App",
		AppVendor:      "Contoso",
		AppVersion:     "2.0.0",
		InstallerPath:  "ExampleApp.msi",
		InstallerType:  "msi",
		Silent:         false,
		Deferrable:     true,
		CloseProcesses: []string{"exampleapp", "examplehelper"},
		TimeoutMinutes: 15,
	}

	result := executeRMMTask(task)

	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("RMM Result: %s\n", output)
}

func executeRMMTask(task RMMTask) RMMResult {
	logs := []string{}
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		fmt.Println(msg)
		logs = append(logs, msg)
	}

	// Create structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 1. Create PSADT client with timeout
	timeout := time.Duration(task.TimeoutMinutes) * time.Minute
	client, err := psadt.NewClient(
		psadt.WithTimeout(timeout),
		psadt.WithLogger(logger),
	)
	if err != nil {
		logf("ERROR: Failed to create PSADT client: %v", err)
		return RMMResult{TaskID: task.TaskID, Success: false, Message: err.Error(), Logs: logs}
	}
	defer func() {
		if client.IsAlive() {
			client.Close()
		}
	}()

	// 2. Get environment info (no session needed yet)
	env, err := client.GetEnvironment()
	if err != nil {
		logf("WARNING: Could not get environment: %v", err)
	} else {
		logf("System: %s %s (%s)", env.OS.Name, env.OS.Version, env.OS.Architecture)
		if env.Domain.IsMachinePartOfDomain {
			logf("Domain: %s", env.Domain.MachineDNSDomain)
		}
	}

	// 3. Open session
	deployMode := types.DeployModeInteractive
	if task.Silent {
		deployMode = types.DeployModeSilent
	}

	session, err := client.OpenSession(types.SessionConfig{
		DeploymentType: types.DeployInstall,
		DeployMode:     deployMode,
		AppVendor:      task.AppVendor,
		AppName:        task.AppName,
		AppVersion:     task.AppVersion,
	})
	if err != nil {
		logf("ERROR: Failed to open session: %v", err)
		return RMMResult{TaskID: task.TaskID, Success: false, Message: err.Error(), Logs: logs}
	}

	// 4. Pre-installation checks
	if !runPreflightChecks(session, task, &logs) {
		session.Close(60001) // preflight failure exit code
		return RMMResult{TaskID: task.TaskID, Success: false, Message: "Preflight checks failed", Logs: logs}
	}

	// 5. Check for user interruptions (RMM-specific)
	// Use WithContext to propagate the RMM task deadline to all PSADT calls
	taskCtx, taskCancel := context.WithTimeout(context.Background(), timeout)
	defer taskCancel()

	// All subsequent session calls use this context via WithContext()
	scopedSession := session.WithContext(taskCtx)

	if err := checkUserState(scopedSession, taskCtx, &logs); err != nil {
		if task.Deferrable {
			logf("Deployment deferred: %v", err)
			session.Close(60002) // deferred
			return RMMResult{TaskID: task.TaskID, Success: false, Message: "Deferred", ExitCode: 60002, Logs: logs}
		}
	}

	// Start live log streaming goroutine — captures PSADT output in real time
	liveCh := session.LiveOutput()
	logDone := make(chan struct{})
	go func() {
		defer close(logDone)
		for line := range liveCh {
			logf("  [PSADT] %s", line)
		}
	}()

	// 6. Show installation welcome
	welcomeOpts := types.WelcomeOptions{
		CloseProcessesCountdown: 300,
		AllowDefer:              task.Deferrable,
		CheckDiskSpace:          true,
	}
	for _, p := range task.CloseProcesses {
		welcomeOpts.CloseProcesses = append(welcomeOpts.CloseProcesses, types.ProcessDefinition{Name: p})
	}
	if len(task.CloseProcesses) == 0 && !task.Silent {
		welcomeOpts.CloseProcesses = []types.ProcessDefinition{{Name: task.AppName}}
	}

	if err := scopedSession.ShowInstallationWelcome(welcomeOpts); err != nil {
		logf("WARNING: Welcome dialog failed (non-fatal): %v", err)
	}

	// 7. Show progress
	if err := scopedSession.ShowInstallationProgress(types.ProgressOptions{
		StatusMessage: fmt.Sprintf("Installing %s %s...", task.AppName, task.AppVersion),
	}); err != nil {
		logf("WARNING: Progress display failed: %v", err)
	}
	defer session.CloseInstallationProgress()

	// 8. Execute installation based on installer type
	var exitCode int
	switch task.InstallerType {
	case "msi":
		result, err := scopedSession.StartMsiProcess(types.MsiProcessOptions{
			Action:   types.MsiInstall,
			FilePath: task.InstallerPath,
			PassThru: true,
		})
		if err != nil {
			logf("ERROR: MSI installation failed: %v", err)
			session.Close(60003)
			return RMMResult{TaskID: task.TaskID, Success: false, Message: err.Error(), Logs: logs}
		}
		exitCode = result.ExitCode
		logf("MSI exit code: %d", exitCode)

	case "exe":
		result, err := scopedSession.StartProcess(types.StartProcessOptions{
			FilePath:    task.InstallerPath,
			PassThru:    true,
			WindowStyle: types.WindowNormal,
		})
		if err != nil {
			logf("ERROR: EXE installation failed: %v", err)
			session.Close(60003)
			return RMMResult{TaskID: task.TaskID, Success: false, Message: err.Error(), Logs: logs}
		}
		exitCode = result.ExitCode
		logf("EXE exit code: %d", exitCode)

	default:
		logf("ERROR: Unknown installer type: %s", task.InstallerType)
		session.Close(60004)
		return RMMResult{TaskID: task.TaskID, Success: false, Message: "Unknown installer type", Logs: logs}
	}

	// 9. Validate installation using batch for performance
	// Batch combines multiple PSADT calls in a single round-trip
	installChecks := []string{
		fmt.Sprintf("(Get-ADTApplication -Name '%s').DisplayName", task.AppName),
		fmt.Sprintf("(Get-ADTApplication -Name '%s').DisplayVersion", task.AppName),
		"Test-ADTNetworkConnection",
	}
	batchData, err := session.ExecuteBatch(taskCtx, installChecks)
	if err != nil {
		logf("WARNING: Batch validation error: %v", err)
	} else {
		logf("Batch validation result: %s", string(batchData))
	}

	// Typed registry check (instead of interface{})
	versionCheck, err := session.GetRegistryKeyString(
		fmt.Sprintf(`HKLM\SOFTWARE\%s\%s`, task.AppVendor, task.AppName),
		"Version",
	)
	if err != nil {
		logf("WARNING: Registry version check: %v", err)
	} else if versionCheck != "" {
		logf("Registry: installed version = %s", versionCheck)
	}

	// Check for pending reboot
	reboot, err := session.GetPendingReboot()
	if err == nil && reboot.IsSystemRebootPending {
		logf("INFO: System reboot pending after installation")
		exitCode = 3010
	}

	// Demonstrate raw script execution (escape hatch for RMM)
	rawScript := `Write-ADTLogEntry -Message "RMM task completed" -Source "go-psadt-rmm" -Severity 1`
	if err := session.ExecuteRawVoidScript(taskCtx, rawScript); err != nil {
		logf("WARNING: Raw script log entry failed: %v", err)
	}
	if err := session.Close(exitCode); err != nil {
		logf("WARNING: Session close issue: %v", err)
	}

	success := exitCode == 0 || exitCode == 3010
	logf("Task %s completed (exit=%d, success=%v)", task.TaskID, exitCode, success)

	return RMMResult{
		TaskID:   task.TaskID,
		Success:  success,
		ExitCode: exitCode,
		Message:  fmt.Sprintf("Installation completed with exit code %d", exitCode),
		Logs:     logs,
	}
}

// runPreflightChecks performs system readiness checks before deployment.
func runPreflightChecks(session *psadt.Session, task RMMTask, logs *[]string) bool {
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		fmt.Println(msg)
		*logs = append(*logs, msg)
	}

	// Check admin privileges
	isAdmin, err := session.TestCallerIsAdmin()
	if err != nil || !isAdmin {
		logf("ERROR: Administrative privileges required")
		return false
	}
	logf("OK: Running with admin privileges")

	// Check battery (don't install on battery for RMM deployments)
	battery, err := session.TestBattery()
	if err == nil && battery.IsLaptop && !battery.IsUsingACPower {
		logf("WARNING: System is on battery power (%d%%)", battery.BatteryChargeLevel)
		if battery.BatteryChargeLevel < 20 {
			logf("ERROR: Battery level too low for deployment")
			return false
		}
	}

	// Check network connectivity
	hasNetwork, err := session.TestNetworkConnection()
	if err != nil || !hasNetwork {
		logf("WARNING: No network connectivity detected")
	}
	if hasNetwork {
		logf("OK: Network connected")
	}

	// Check disk space
	freeSpace, err := session.GetFreeDiskSpace()
	if err == nil {
		logf("OK: Free disk space: %d MB", freeSpace)
		if freeSpace < 1024 {
			logf("WARNING: Low disk space (< 1GB)")
		}
	}

	// Check ESP/OOBE status
	isOOBE, err := session.TestOobeCompleted()
	if err == nil && !isOOBE {
		logf("WARNING: Device is still in OOBE/Autopilot ESP")
	}

	return true
}

// checkUserState verifies user state and shows appropriate notifications.
// Returns an error if deployment should be deferred.
func checkUserState(session *psadt.Session, _ context.Context, logs *[]string) error {
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		fmt.Println(msg)
		*logs = append(*logs, msg)
	}

	// Check if user is in Focus Assist mode
	inFocus, err := session.TestUserInFocusMode()
	if err == nil && inFocus {
		logf("INFO: User is in Focus Assist mode - deferring deployment")
		// Show a balloon tip to notify user
		session.ShowBalloonTip(types.BalloonTipOptions{
			BalloonTipText:  "Deployment deferred. We'll try again when you're available.",
			BalloonTipTitle: "Software Update",
			BalloonTipIcon:  types.BalloonInfo,
			BalloonTipTime:  10000,
		})
		return fmt.Errorf("user is in focus mode")
	}

	// Check if PowerPoint is in presentation mode
	pptRunning, err := session.TestPowerPoint()
	if err == nil && pptRunning {
		logf("INFO: PowerPoint presentation detected - deferring deployment")
		return fmt.Errorf("user is presenting")
	}

	// Check microphone usage (user may be in a call)
	micInUse, err := session.TestMicrophoneInUse()
	if err == nil && micInUse {
		logf("INFO: Microphone in use - user may be in a meeting")
	}

	// Check overall user busy status
	isBusy, err := session.TestUserIsBusy()
	if err == nil && isBusy {
		logf("INFO: User appears to be busy")
		// Show notification but continue
		session.ShowBalloonTip(types.BalloonTipOptions{
			BalloonTipText:  "Software update will begin shortly.",
			BalloonTipTitle: "Update Notice",
			BalloonTipIcon:  types.BalloonInfo,
			BalloonTipTime:  8000,
		})
	}

	return nil
}

// ensureClientHealth verifies the client is alive and reconnects if needed.
// This is useful for long-running RMM agents that hold a client reference.
func ensureClientHealth(client *psadt.Client, ctx context.Context) error {
	if !client.IsAlive() {
		return client.Reconnect(ctx)
	}
	return nil
}
