# go-psadt

Go wrapper library for [PSAppDeployToolkit](https://psappdeploytoolkit.com/) v4.1.x.

Allows Go applications to orchestrate Windows software deployments, display UI dialogs, manage registry/services/filesystem, and invoke installers — all through an idiomatic, type-safe Go API.

## Features

- **~105 wrapped PSADT functions** as strongly-typed Go methods
- **Persistent PowerShell process** with stdin/stdout JSON protocol
- **Type-safe options** — all parameters are Go structs with `ps:` tags
- **Session management** — Open/Close ADT sessions with full configuration
- **Windows-only** — uses `//go:build windows` build constraint
- **PowerShell 5.1 and 7+ support**

## Prerequisites

| Requirement | Version |
|---|---|
| Go | ≥ 1.21 |
| Windows | 10/11 or Server 2016+ |
| PowerShell | ≥ 5.1 |
| PSAppDeployToolkit | ≥ 4.1.0 |

```powershell
Install-Module -Name PSAppDeployToolkit -Scope AllUsers
```

## Installation

```bash
go get github.com/peterondra/go-psadt
```

## Quick Start

```go
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
        psadt.WithTimeout(10 * time.Minute),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    session, err := client.OpenSession(types.SessionConfig{
        DeploymentType: types.DeployInstall,
        DeployMode:     types.DeployModeInteractive,
        AppVendor:      "Contoso",
        AppName:        "Widget Pro",
        AppVersion:     "2.0.0",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close(0)

    // Show welcome dialog
    session.ShowInstallationWelcome(types.WelcomeOptions{
        CloseProcesses:          []types.ProcessDefinition{{Name: "widget"}},
        CloseProcessesCountdown: 300,
        CheckDiskSpace:          true,
    })

    // Install MSI
    result, err := session.StartMsiProcess(types.MsiProcessOptions{
        Action:   types.MsiInstall,
        FilePath: "WidgetPro.msi",
        PassThru: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Exit code: %d\n", result.ExitCode)
}
```

## Architecture

```
┌─────────────────────────┐
│     Go Application      │
│                         │
│  client.OpenSession()   │
│  session.Method(opts)   │
│  session.Close(0)       │
└──────────┬──────────────┘
           │ Go API
┌──────────▼──────────────┐
│      go-psadt           │
│                         │
│  cmdbuilder → PS string │
│  runner     → stdin/out │
│  parser     → Go struct │
└──────────┬──────────────┘
           │ JSON protocol
┌──────────▼──────────────┐
│  powershell.exe         │
│  (persistent process)   │
│  Import-Module PSADT    │
└─────────────────────────┘
```

## Package Structure

```
go-psadt/
├── psadt.go            # Client, options, NewClient
├── session.go          # Session lifecycle
├── environment.go      # Client.GetEnvironment()
├── ui.go               # Dialogs, prompts, progress
├── process.go          # Start processes (EXE, MSI, MSP)
├── application.go      # Search/uninstall applications
├── registry.go         # Registry operations
├── filesystem.go       # File/folder operations
├── ini.go              # INI file operations
├── envvar.go           # Environment variables
├── shortcut.go         # Shortcut management
├── service.go          # Windows services
├── wim.go              # WIM/ZIP operations
├── sysinfo.go          # System information queries
├── checks.go           # System state checks
├── dll.go              # DLL registration
├── msi.go              # MSI database operations
├── activesetup.go      # Active Setup
├── edge.go             # Edge extension policies
├── system.go           # Desktop/GPO/SCCM/TS
├── logging.go          # Log entries
├── config.go           # Configuration/defer history
├── util.go             # Utilities (SendKeys, permissions, etc.)
├── types/              # All type definitions
├── internal/
│   ├── cmdbuilder/     # Go struct → PS command string
│   ├── parser/         # JSON response → Go struct
│   └── runner/         # PowerShell process management
└── examples/
    ├── install/        # MSI installation example
    ├── uninstall/      # Uninstallation example
    ├── dialog/         # UI dialog examples
    └── query/          # System query examples
```

## Function Categories

| Category | Methods | Description |
|---|---|---|
| **UI** | 8 | Welcome, prompts, progress, dialogs, balloon tips |
| **Process** | 9 | Start EXE/MSI/MSP, block/unblock apps |
| **Application** | 2 | Search and uninstall applications |
| **Registry** | 5 | Get/Set/Remove keys, test values, all-users action |
| **Filesystem** | 8 | Copy/remove files/folders, user profiles, cache |
| **INI** | 6 | Get/Set/Remove values and sections |
| **Environment** | 3 | Get/Set/Remove environment variables |
| **Shortcut** | 3 | New/Set/Get shortcuts |
| **Service** | 5 | Start/Stop services, start mode, existence check |
| **WIM/ZIP** | 3 | Mount/dismount WIM, create ZIP |
| **System Info** | 11 | Users, disk, reboot, OS, profiles, versions, windows |
| **Checks** | 10 | Battery, admin, network, mutex, PowerPoint, etc. |
| **DLL** | 3 | Register/unregister DLL, RegSvr32 |
| **MSI** | 4 | Exit codes, table properties, transforms |
| **Active Setup** | 1 | Set Active Setup entries |
| **Edge** | 2 | Add/remove Edge extension policies |
| **System** | 7 | Desktop, GPO, MS Updates, SCCM, Terminal Server |
| **Logging** | 1 | Write log entries |
| **Config** | 6 | Config, string table, defer history, culture |
| **Utilities** | 8 | SendKeys, permissions, retry, encoding, templates |

## Client Options

```go
psadt.WithTimeout(10 * time.Minute)     // Command timeout
psadt.WithPSPath("pwsh.exe")            // Custom PowerShell path
psadt.WithPowerShell7()                 // Use PowerShell 7
psadt.WithMinModuleVersion("4.1.0")     // Minimum PSADT version
psadt.WithLogger(myLogger)              // Custom logger interface
```

## License

MIT — see [LICENSE](LICENSE).
