# Example: RMM Agent Integration

This example demonstrates how an RMM (Remote Monitoring and Management) agent can use `go-psadt` for enterprise software deployment scenarios.

## Features Demonstrated

| Feature | Description |
|---|---|
| **Pre-flight Checks** | Admin rights, battery, network, disk space, OOBE status |
| **User State Detection** | Focus Assist, PowerPoint, microphone, busy status |
| **Deferral Logic** | Respect user activity, defer when appropriate |
| **Interactive Prompt** | Show welcome dialogs, allow user deferral |
| **Progress Tracking** | Real-time installation progress display |
| **MSI & EXE Support** | Both installer types with exit code handling |
| **Batch Execution** | Multiple PSADT calls in a single round-trip |
| **Typed Registry Access** | `GetRegistryKeyString` / `GetRegistryKeyDWord` instead of `interface{}` |
| **Raw Script Escape Hatch** | `ExecuteRawScript` / `ExecuteRawVoidScript` for custom PowerShell |
| **Client Health** | Runner reconnection for long-running agents |
| **Environment Caching** | `GetEnvironment()` caches results, `InvalidateEnvCache()` to refresh |
| **Structured Logging** | `slog` logger with task-level log collection |
| **Result Reporting** | JSON output for RMM platform integration |

## Usage

```powershell
# Build and run from project root
go build -o bin/rmm-example.exe ./examples/rmm/
.\bin\rmm-example.exe
```

## Key Patterns for RMM Integration

### 1. Session Lifecycle
```go
client, _ := psadt.NewClient(psadt.WithTimeout(15 * time.Minute))
defer client.Close()

session, _ := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployInstall,
    DeployMode:     types.DeployModeInteractive,
    AppVendor:      "Contoso",
    AppName:        "Example App",
    AppVersion:     "2.0.0",
})
defer session.Close(0)
```

### 2. Respecting User State
```go
// Check Focus Assist before interrupting
inFocus, _ := session.TestUserInFocusMode()
if inFocus {
    // Defer deployment
    session.ShowBalloonTip(types.BalloonTipOptions{
        BalloonTipText:  "Deployment deferred.",
        BalloonTipTitle: "Software Update",
        BalloonTipIcon:  types.BalloonInfo,
    })
    return
}
```

### 3. Client Health Monitoring
```go
if !client.IsAlive() {
    client.Reconnect(ctx)
}
```

### 4. Batch Execution (reduces round-trips)
```go
data, err := session.ExecuteBatch(ctx, []string{
    "(Get-ADTApplication -Name 'MyApp').DisplayName",
    "(Get-ADTApplication -Name 'MyApp').DisplayVersion",
    "Test-ADTNetworkConnection",
})
```

### 5. Typed Registry Access
```go
// Instead of interface{}, use typed helpers:
version, _ := session.GetRegistryKeyString(`HKLM\SOFTWARE\Contoso`, "Version")
settings, _ := session.GetRegistryKeyDWord(`HKLM\SOFTWARE\Contoso`, "ConfigFlag")
```

### 6. Raw Script Execution (escape hatch)
```go
// Run any PowerShell in the PSADT session context
err := session.ExecuteRawVoidScript(ctx, 
    `Write-ADTLogEntry -Message "Custom logic" -Source "rmm-agent" -Severity 1`)
```

### 7. Environment Caching
```go
env, _ := client.GetEnvironment() // Fetches once
client.InvalidateEnvCache()       // Force refresh on next call
```go
result := RMMResult{
    TaskID:   task.TaskID,
    Success:  exitCode == 0,
    ExitCode: exitCode,
    Message:  "Installation completed",
    Logs:     logs,
}
jsonBytes, _ := json.MarshalIndent(result, "", "  ")
// Send to RMM platform
```
