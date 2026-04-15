<div align="center">

# go-psadt

**Go wrapper library for [PSAppDeployToolkit](https://psappdeploytoolkit.com/) v4.1.x**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Windows-0078D6?style=flat-square&logo=windows)](https://microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![PSADT](https://img.shields.io/badge/PSADT-%3E%3D4.1.0-purple?style=flat-square)](https://psappdeploytoolkit.com/)

Orchestrate Windows software deployments, display UI dialogs, manage registry/services/filesystem, and invoke installers — all through an idiomatic, type-safe Go API.

[English](#english) | [Português](#português)

</div>

---

> [!NOTE]
> **AI-Assisted Development**: This project was developed with the assistance of AI tools (GitHub Copilot / Claude). All generated code was reviewed, tested, and validated to ensure correctness and quality. The architecture, design decisions, and implementation plan were human-directed.
>
> **Desenvolvimento Assistido por IA**: Este projeto foi desenvolvido com auxílio de ferramentas de IA (GitHub Copilot / Claude). Todo o código gerado foi revisado, testado e validado para garantir correção e qualidade. A arquitetura, decisões de design e plano de implementação foram dirigidos por humanos.

---

# English

## Overview

**go-psadt** is a Go library that wraps the [PSAppDeployToolkit (PSADT)](https://psappdeploytoolkit.com/) v4.1.x — a PowerShell framework with **135+ exported functions** for Windows software deployment automation. This library exposes **~105 public methods** as strongly-typed Go functions, enabling Go applications to leverage PSADT's full power without writing PowerShell directly.

### Why go-psadt?

- **Enterprise deployment automation** from Go applications
- **Type-safe API** — catch errors at compile time, not runtime
- **No PowerShell knowledge required** — the library handles all command construction and response parsing
- **Persistent process** — a single PowerShell instance serves all calls, avoiding per-command startup overhead
- **Full PSADT coverage** — UI dialogs, MSI/EXE/MSP installers, registry, filesystem, services, and more

## Features

| Feature | Description |
|---|---|
| **~105 wrapped functions** | Strongly-typed Go methods covering all major PSADT capabilities |
| **Persistent PowerShell process** | Single `powershell.exe` instance with stdin/stdout JSON protocol |
| **Type-safe options** | All parameters are Go structs with `ps:` struct tags — no raw strings |
| **Session management** | Full Open/Close ADT session lifecycle with configuration |
| **Dual PowerShell support** | Windows PowerShell 5.1 and PowerShell 7+ |
| **Structured error handling** | PSADT errors are parsed into Go error types with stack traces |
| **Context support** | All operations support `context.Context` for timeouts and cancellation |
| **Windows-only build** | Uses `//go:build windows` — cleanly excluded on other platforms |

## Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| **Go** | ≥ 1.21 | Module-based project |
| **Windows** | 10/11 or Server 2016+ | Library is Windows-only |
| **PowerShell** | ≥ 5.1 | Windows PowerShell (built-in) or PowerShell 7+ |
| **.NET Framework** | ≥ 4.7.2 | Required by PSADT C# assemblies |
| **PSAppDeployToolkit** | ≥ 4.1.0 | The underlying PowerShell module |

```powershell
# Install PSADT module
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
    // Create a PSADT client (starts PowerShell, imports module)
    client, err := psadt.NewClient(
        psadt.WithTimeout(10 * time.Minute),
    )
    if err != nil {
        log.Fatal(err)
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
        log.Fatal(err)
    }
    defer session.Close(0)

    // Show welcome dialog — prompt to close running apps
    session.ShowInstallationWelcome(types.WelcomeOptions{
        CloseProcesses:          []types.ProcessDefinition{{Name: "widget"}},
        CloseProcessesCountdown: 300,
        CheckDiskSpace:          true,
    })

    // Run MSI installer
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

The library maintains a **persistent PowerShell process** that communicates via stdin/stdout using a JSON-based protocol with delimited markers. This avoids the overhead of starting a new process for each command.

```
┌──────────────────────────────────────────────────────────────────┐
│                        Go Application                            │
│                                                                  │
│   client, _ := psadt.NewClient(opts...)                          │
│   session, _ := client.OpenSession(cfg)                          │
│   session.ShowInstallationWelcome(opts)                          │
│   session.StartMsiProcess(opts)                                  │
│   session.Close(0)                                               │
└───────────────────────┬──────────────────────────────────────────┘
                        │
                        │  Typed Go API (compile-time safety)
                        │
┌───────────────────────▼──────────────────────────────────────────┐
│                       go-psadt Library                            │
│                                                                  │
│  ┌──────────────┐   ┌──────────────────┐   ┌──────────────────┐ │
│  │  cmdbuilder   │   │     runner        │   │     parser       │ │
│  │              │   │                  │   │                  │ │
│  │  Go struct   │   │  Persistent PS   │   │  JSON response   │ │
│  │  + ps: tags  │   │  process with    │   │  → Go struct     │ │
│  │  → PS command│   │  stdin/stdout    │   │  + error types   │ │
│  │  string      │   │  pipes + mutex   │   │                  │ │
│  └──────┬───────┘   └────────┬─────────┘   └────────┬─────────┘ │
│         │                    │                       │           │
└─────────┼────────────────────┼───────────────────────┼───────────┘
          │                    │                       │
          │   ┌────────────────▼──────────────────┐    │
          └──►│       powershell.exe               │◄──┘
              │       (persistent process)         │
              │                                    │
              │  ┌────────────────────────────┐    │
              │  │  Import-Module PSADT       │    │
              │  │  Open-ADTSession ...       │    │
              │  │                            │    │
              │  │  <<<PSADT_BEGIN>>>          │    │
              │  │  { JSON response }         │    │
              │  │  <<<PSADT_END>>>            │    │
              │  └────────────────────────────┘    │
              └────────────────────────────────────┘
```

### Communication Protocol

Each command is wrapped in a PowerShell `try/catch` block with JSON serialization:

1. **Go → PowerShell**: `cmdbuilder` converts typed Go structs into PowerShell command strings using reflection and `ps:` struct tags
2. **Execution**: The command runs inside the persistent PowerShell process with `<<<PSADT_BEGIN>>>` / `<<<PSADT_END>>>` delimiters
3. **PowerShell → Go**: `parser` extracts the JSON between markers and deserializes into Go types
4. **Concurrency**: A mutex ensures commands are serialized — one command at a time per client

### Internal Packages

| Package | Responsibility |
|---|---|
| `internal/cmdbuilder` | Converts Go option structs into PowerShell command strings via reflection on `ps:` tags. Handles string escaping, arrays, hashtables, switch parameters, and script blocks. |
| `internal/runner` | Manages the persistent `powershell.exe` process lifecycle: start, stop, heartbeat, module import, version validation. Provides `Execute()` and `ExecuteVoid()` for command dispatch. |
| `internal/parser` | Parses JSON responses from PowerShell into Go types. Handles success/error discrimination, typed error extraction (`PSADTError`), and convenience parsers for bool, string, uint64. |

## Package Structure

```
go-psadt/
│
├── psadt.go              # Client struct, NewClient(), options pattern
├── session.go            # Session lifecycle (Open/Close/GetProperties)
├── environment.go        # Client.GetEnvironment() — ~90 PSADT variables
│
├── ui.go                 # 8 methods: Welcome, Prompt, Progress, Dialog, Balloon, etc.
├── process.go            # 9 methods: Start EXE/MSI/MSP (+ AsUser), Block/Unblock
├── application.go        # 2 methods: GetApplication, UninstallApplication
├── registry.go           # 5 methods: Get/Set/Remove keys, Test, AllUsers action
├── filesystem.go         # 8 methods: Copy/Remove files/folders, UserProfiles, Cache
├── ini.go                # 6 methods: Get/Set/Remove values and sections
├── envvar.go             # 3 methods: Get/Set/Remove environment variables
├── shortcut.go           # 3 methods: New/Set/Get shortcuts
├── service.go            # 5 methods: Start/Stop, Get/Set start mode, TestExists
├── wim.go                # 3 methods: Mount/Dismount WIM, NewZipFile
├── sysinfo.go            # 11 methods: Users, disk, reboot, OS, profiles, versions
├── checks.go             # 10 methods: Battery, admin, network, mutex, busy state
├── dll.go                # 3 methods: Register/Unregister DLL, RegSvr32
├── msi.go                # 4 methods: Exit codes, table properties, transforms
├── activesetup.go        # 1 method: SetActiveSetup
├── edge.go               # 2 methods: Add/Remove Edge extensions
├── system.go             # 7 methods: Desktop, GPO, Updates, SCCM, Terminal Server
├── logging.go            # 1 method: WriteLogEntry
├── config.go             # 6 methods: Config, StringTable, DeferHistory, Culture
├── util.go               # 8 methods: SendKeys, permissions, retry, encoding
│
├── types/                # All type definitions (22 files)
│   ├── enums.go          #   DeploymentType, DeployMode, DialogStyle, ...
│   ├── exitcodes.go      #   Standard PSADT exit code constants
│   ├── session.go        #   SessionConfig, SessionProperties
│   ├── process.go        #   ProcessResult, StartProcessOptions, MsiProcessOptions, ...
│   ├── application.go    #   InstalledApplication, GetApplicationOptions, ...
│   ├── ui.go             #   WelcomeOptions, PromptOptions, DialogBoxOptions, ...
│   ├── environment.go    #   EnvironmentInfo (hierarchical: OS, PowerShell, Users, ...)
│   ├── registry.go       #   Get/Set/Remove RegistryKeyOptions, ...
│   ├── filesystem.go     #   CopyFileOptions, RemoveFolderOptions, ...
│   └── ...               #   service, sysinfo, shortcut, wim, dll, msi, config, ...
│
├── internal/
│   ├── cmdbuilder/       # Go struct → PowerShell command string (reflection + ps: tags)
│   ├── parser/           # JSON protocol → Go structs (success/error handling)
│   └── runner/           # PowerShell process management (lifecycle + protocol)
│
├── examples/
│   ├── install/          # Complete MSI installation example
│   ├── uninstall/        # Application uninstallation example
│   ├── dialog/           # UI dialogs and prompts demo
│   └── query/            # System information queries
│
├── go.mod
├── LICENSE               # MIT
└── README.md
```

## Function Categories

| Category | Methods | PSADT Functions |
|---|---|---|
| **UI** | 8 | `Show-ADTInstallationWelcome`, `Show-ADTInstallationPrompt`, `Show-ADTInstallationProgress`, `Close-ADTInstallationProgress`, `Show-ADTInstallationRestartPrompt`, `Show-ADTDialogBox`, `Show-ADTBalloonTip`, `Show-ADTHelpConsole` |
| **Process** | 9 | `Start-ADTProcess`, `Start-ADTProcessAsUser`, `Start-ADTMsiProcess`, `Start-ADTMsiProcessAsUser`, `Start-ADTMspProcess`, `Start-ADTMspProcessAsUser`, `Block-ADTAppExecution`, `Unblock-ADTAppExecution`, `Get-ADTRunningProcesses` |
| **Application** | 2 | `Get-ADTApplication`, `Uninstall-ADTApplication` |
| **Registry** | 5 | `Get-ADTRegistryKey`, `Set-ADTRegistryKey`, `Remove-ADTRegistryKey`, `Test-ADTRegistryValue`, `Invoke-ADTAllUsersRegistryAction` |
| **Filesystem** | 8 | `Copy-ADTFile`, `Copy-ADTFileToUserProfiles`, `Remove-ADTFile`, `Remove-ADTFileFromUserProfiles`, `New-ADTFolder`, `Remove-ADTFolder`, `Copy-ADTContentToCache`, `Remove-ADTContentFromCache` |
| **INI** | 6 | `Get-ADTIniValue`, `Set-ADTIniValue`, `Remove-ADTIniValue`, `Get-ADTIniSection`, `Set-ADTIniSection`, `Remove-ADTIniSection` |
| **Environment** | 3 | `Get-ADTEnvironmentVariable`, `Set-ADTEnvironmentVariable`, `Remove-ADTEnvironmentVariable` |
| **Shortcut** | 3 | `New-ADTShortcut`, `Set-ADTShortcut`, `Get-ADTShortcut` |
| **Service** | 5 | `Start-ADTServiceAndDependencies`, `Stop-ADTServiceAndDependencies`, `Get-ADTServiceStartMode`, `Set-ADTServiceStartMode`, `Test-ADTServiceExists` |
| **WIM/ZIP** | 3 | `Mount-ADTWimFile`, `Dismount-ADTWimFile`, `New-ADTZipFile` |
| **System Info** | 11 | `Get-ADTLoggedOnUser`, `Get-ADTFreeDiskSpace`, `Get-ADTPendingReboot`, `Get-ADTOperatingSystemInfo`, `Get-ADTUserProfiles`, `Get-ADTFileVersion`, `Get-ADTExecutableInfo`, `Get-ADTPEFileArchitecture`, `Get-ADTWindowTitle`, `Get-ADTPresentationSettingsEnabledUsers`, `Get-ADTUserNotificationState` |
| **Checks** | 10 | `Test-ADTBattery`, `Test-ADTCallerIsAdmin`, `Test-ADTNetworkConnection`, `Test-ADTMutexAvailability`, `Test-ADTPowerPoint`, `Test-ADTMicrophoneInUse`, `Test-ADTUserIsBusy`, `Test-ADTEspActive`, `Test-ADTOobeCompleted`, `Test-ADTMSUpdates` |
| **DLL** | 3 | `Register-ADTDll`, `Unregister-ADTDll`, `Invoke-ADTRegSvr32` |
| **MSI** | 4 | `Get-ADTMsiExitCodeMessage`, `Get-ADTMsiTableProperty`, `Set-ADTMsiProperty`, `New-ADTMsiTransform` |
| **Active Setup** | 1 | `Set-ADTActiveSetup` |
| **Edge** | 2 | `Add-ADTEdgeExtension`, `Remove-ADTEdgeExtension` |
| **System** | 7 | `Update-ADTDesktop`, `Update-ADTGroupPolicy`, `Install-ADTMSUpdates`, `Install-ADTSCCMSoftwareUpdates`, `Invoke-ADTSCCMTask`, `Enable-ADTTerminalServerInstallMode`, `Disable-ADTTerminalServerInstallMode` |
| **Logging** | 1 | `Write-ADTLogEntry` |
| **Config** | 6 | `Get-ADTConfig`, `Get-ADTStringTable`, `Get-ADTDeferHistory`, `Set-ADTDeferHistory`, `Reset-ADTDeferHistory`, `Set-ADTPowerShellCulture` |
| **Utilities** | 8 | `Send-ADTKeys`, `Convert-ADTToNTAccountOrSID`, `Set-ADTItemPermission`, `Invoke-ADTCommandWithRetries`, `Get-ADTUniversalDate`, `Remove-ADTInvalidFileNameChars`, `Out-ADTPowerShellEncodedCommand`, `New-ADTTemplate` |

## Client Options

```go
psadt.WithTimeout(10 * time.Minute)     // Command execution timeout (default: 5 min)
psadt.WithPSPath("pwsh.exe")            // Custom PowerShell executable path
psadt.WithPowerShell7()                 // Use PowerShell 7+ (pwsh.exe)
psadt.WithMinModuleVersion("4.1.0")     // Minimum PSADT module version
psadt.WithLogger(myLogger)              // Custom *slog.Logger for diagnostics
```

## Examples

The [`examples/`](examples/) directory contains complete, runnable programs:

| Example | Description |
|---|---|
| [`examples/install/`](examples/install/) | Full MSI installation with welcome dialog, progress bar, and registry configuration |
| [`examples/uninstall/`](examples/uninstall/) | Application search and uninstallation with cleanup |
| [`examples/dialog/`](examples/dialog/) | UI interactions: dialog boxes, balloon tips, installation prompts |
| [`examples/query/`](examples/query/) | System information queries: environment, admin status, disk space, users, services |

## Error Handling

PSADT errors are automatically parsed into structured Go errors:

```go
result, err := session.StartMsiProcess(opts)
if err != nil {
    // Check for specific PSADT error types
    if parser.IsRebootRequired(err) {
        log.Println("Reboot required after installation")
    } else if parser.IsUserCancelled(err) {
        log.Println("User cancelled the installation")
    } else {
        log.Fatalf("Installation failed: %v", err)
    }
}
```

## Contributing

Contributions are welcome. Please open an issue first to discuss what you would like to change.

---

# Português

## Visão Geral

**go-psadt** é uma biblioteca Go que encapsula o [PSAppDeployToolkit (PSADT)](https://psappdeploytoolkit.com/) v4.1.x — um framework PowerShell com **135+ funções exportadas** para automação de deployment de software Windows. Esta biblioteca expõe **~105 métodos públicos** como funções Go fortemente tipadas, permitindo que aplicações Go utilizem todo o poder do PSADT sem escrever PowerShell diretamente.

### Por que go-psadt?

- **Automação de deployment corporativo** a partir de aplicações Go
- **API type-safe** — erros são capturados em tempo de compilação, não em runtime
- **Sem necessidade de conhecimento PowerShell** — a biblioteca constrói todos os comandos e faz o parsing das respostas
- **Processo persistente** — uma única instância PowerShell atende todas as chamadas, evitando overhead de inicialização por comando
- **Cobertura completa do PSADT** — diálogos UI, instaladores MSI/EXE/MSP, registro, filesystem, serviços e mais

## Funcionalidades

| Funcionalidade | Descrição |
|---|---|
| **~105 funções encapsuladas** | Métodos Go fortemente tipados cobrindo todas as capacidades principais do PSADT |
| **Processo PowerShell persistente** | Instância única de `powershell.exe` com protocolo JSON via stdin/stdout |
| **Opções type-safe** | Todos os parâmetros são structs Go com tags `ps:` — sem strings cruas |
| **Gerenciamento de sessão** | Ciclo de vida completo de sessão ADT (Open/Close) com configuração |
| **Suporte dual PowerShell** | Windows PowerShell 5.1 e PowerShell 7+ |
| **Tratamento estruturado de erros** | Erros PSADT são parseados em tipos Go com stack traces |
| **Suporte a Context** | Todas as operações suportam `context.Context` para timeouts e cancelamento |
| **Build Windows-only** | Usa `//go:build windows` — excluído automaticamente em outras plataformas |

## Pré-requisitos

| Requisito | Versão | Notas |
|---|---|---|
| **Go** | ≥ 1.21 | Projeto baseado em módulos |
| **Windows** | 10/11 ou Server 2016+ | Biblioteca é Windows-only |
| **PowerShell** | ≥ 5.1 | Windows PowerShell (nativo) ou PowerShell 7+ |
| **.NET Framework** | ≥ 4.7.2 | Necessário para assemblies C# do PSADT |
| **PSAppDeployToolkit** | ≥ 4.1.0 | O módulo PowerShell subjacente |

```powershell
# Instalar módulo PSADT
Install-Module -Name PSAppDeployToolkit -Scope AllUsers
```

## Instalação

```bash
go get github.com/peterondra/go-psadt
```

## Início Rápido

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
    // Cria um cliente PSADT (inicia PowerShell, importa módulo)
    client, err := psadt.NewClient(
        psadt.WithTimeout(10 * time.Minute),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Abre uma sessão de deployment
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

    // Exibe diálogo de boas-vindas — solicita fechar apps em execução
    session.ShowInstallationWelcome(types.WelcomeOptions{
        CloseProcesses:          []types.ProcessDefinition{{Name: "widget"}},
        CloseProcessesCountdown: 300,
        CheckDiskSpace:          true,
    })

    // Executa instalador MSI
    result, err := session.StartMsiProcess(types.MsiProcessOptions{
        Action:   types.MsiInstall,
        FilePath: "WidgetPro.msi",
        PassThru: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Código de saída: %d\n", result.ExitCode)
}
```

## Arquitetura

A biblioteca mantém um **processo PowerShell persistente** que se comunica via stdin/stdout usando um protocolo baseado em JSON com marcadores delimitadores. Isso evita o overhead de iniciar um novo processo para cada comando.

```
┌──────────────────────────────────────────────────────────────────┐
│                       Aplicação Go                               │
│                                                                  │
│   client, _ := psadt.NewClient(opts...)                          │
│   session, _ := client.OpenSession(cfg)                          │
│   session.ShowInstallationWelcome(opts)                          │
│   session.StartMsiProcess(opts)                                  │
│   session.Close(0)                                               │
└───────────────────────┬──────────────────────────────────────────┘
                        │
                        │  API Go tipada (segurança em tempo de compilação)
                        │
┌───────────────────────▼──────────────────────────────────────────┐
│                      Biblioteca go-psadt                          │
│                                                                  │
│  ┌──────────────┐   ┌──────────────────┐   ┌──────────────────┐ │
│  │  cmdbuilder   │   │     runner        │   │     parser       │ │
│  │              │   │                  │   │                  │ │
│  │  Struct Go   │   │  Processo PS     │   │  Resposta JSON   │ │
│  │  + tags ps:  │   │  persistente com │   │  → Struct Go     │ │
│  │  → comando   │   │  pipes stdin/out │   │  + tipos de erro │ │
│  │  PowerShell  │   │  + mutex         │   │                  │ │
│  └──────┬───────┘   └────────┬─────────┘   └────────┬─────────┘ │
│         │                    │                       │           │
└─────────┼────────────────────┼───────────────────────┼───────────┘
          │                    │                       │
          │   ┌────────────────▼──────────────────┐    │
          └──►│       powershell.exe               │◄──┘
              │       (processo persistente)        │
              │                                    │
              │  ┌────────────────────────────┐    │
              │  │  Import-Module PSADT       │    │
              │  │  Open-ADTSession ...       │    │
              │  │                            │    │
              │  │  <<<PSADT_BEGIN>>>          │    │
              │  │  { resposta JSON }         │    │
              │  │  <<<PSADT_END>>>            │    │
              │  └────────────────────────────┘    │
              └────────────────────────────────────┘
```

### Protocolo de Comunicação

Cada comando é envolvido em um bloco `try/catch` do PowerShell com serialização JSON:

1. **Go → PowerShell**: O `cmdbuilder` converte structs Go tipadas em strings de comando PowerShell usando reflection e tags `ps:`
2. **Execução**: O comando é executado dentro do processo PowerShell persistente com delimitadores `<<<PSADT_BEGIN>>>` / `<<<PSADT_END>>>`
3. **PowerShell → Go**: O `parser` extrai o JSON entre os marcadores e desserializa em tipos Go
4. **Concorrência**: Um mutex garante que os comandos são serializados — um comando por vez por cliente

### Pacotes Internos

| Pacote | Responsabilidade |
|---|---|
| `internal/cmdbuilder` | Converte structs de opções Go em strings de comando PowerShell via reflection nas tags `ps:`. Trata escaping de strings, arrays, hashtables, parâmetros switch e script blocks. |
| `internal/runner` | Gerencia o ciclo de vida do processo `powershell.exe` persistente: start, stop, heartbeat, importação de módulo, validação de versão. Fornece `Execute()` e `ExecuteVoid()` para despacho de comandos. |
| `internal/parser` | Faz o parsing de respostas JSON do PowerShell em tipos Go. Trata discriminação sucesso/erro, extração de erros tipados (`PSADTError`) e parsers de conveniência para bool, string, uint64. |

## Estrutura de Pacotes

```
go-psadt/
│
├── psadt.go              # Struct Client, NewClient(), padrão options
├── session.go            # Ciclo de vida da sessão (Open/Close/GetProperties)
├── environment.go        # Client.GetEnvironment() — ~90 variáveis PSADT
│
├── ui.go                 # 8 métodos: Welcome, Prompt, Progress, Dialog, Balloon, etc.
├── process.go            # 9 métodos: Iniciar EXE/MSI/MSP (+ AsUser), Block/Unblock
├── application.go        # 2 métodos: GetApplication, UninstallApplication
├── registry.go           # 5 métodos: Get/Set/Remove chaves, Test, ação AllUsers
├── filesystem.go         # 8 métodos: Copy/Remove arquivos/pastas, UserProfiles, Cache
├── ini.go                # 6 métodos: Get/Set/Remove valores e seções
├── envvar.go             # 3 métodos: Get/Set/Remove variáveis de ambiente
├── shortcut.go           # 3 métodos: New/Set/Get atalhos
├── service.go            # 5 métodos: Start/Stop, Get/Set modo de início, TestExists
├── wim.go                # 3 métodos: Mount/Dismount WIM, NewZipFile
├── sysinfo.go            # 11 métodos: Usuários, disco, reboot, OS, perfis, versões
├── checks.go             # 10 métodos: Bateria, admin, rede, mutex, estado ocupado
├── dll.go                # 3 métodos: Register/Unregister DLL, RegSvr32
├── msi.go                # 4 métodos: Códigos de saída, propriedades, transforms
├── activesetup.go        # 1 método: SetActiveSetup
├── edge.go               # 2 métodos: Add/Remove extensões Edge
├── system.go             # 7 métodos: Desktop, GPO, Updates, SCCM, Terminal Server
├── logging.go            # 1 método: WriteLogEntry
├── config.go             # 6 métodos: Config, StringTable, DeferHistory, Culture
├── util.go               # 8 métodos: SendKeys, permissões, retry, encoding
│
├── types/                # Todas as definições de tipos (22 arquivos)
├── internal/             # Pacotes internos (cmdbuilder, parser, runner)
├── examples/             # Programas de exemplo executáveis
├── go.mod
├── LICENSE               # MIT
└── README.md
```

## Categorias de Funções

| Categoria | Métodos | Descrição |
|---|---|---|
| **UI** | 8 | Boas-vindas, prompts, progresso, diálogos, notificações |
| **Processo** | 9 | Iniciar EXE/MSI/MSP, bloquear/desbloquear apps |
| **Aplicação** | 2 | Buscar e desinstalar aplicações |
| **Registro** | 5 | Get/Set/Remove chaves, testar valores, ação multi-usuário |
| **Filesystem** | 8 | Copiar/remover arquivos/pastas, perfis de usuário, cache |
| **INI** | 6 | Get/Set/Remove valores e seções de arquivos INI |
| **Ambiente** | 3 | Get/Set/Remove variáveis de ambiente |
| **Atalho** | 3 | Criar/modificar/consultar atalhos |
| **Serviço** | 5 | Iniciar/parar serviços, modo de início, verificar existência |
| **WIM/ZIP** | 3 | Montar/desmontar WIM, criar ZIP |
| **Info Sistema** | 11 | Usuários, disco, reboot, OS, perfis, versões, janelas |
| **Verificações** | 10 | Bateria, admin, rede, mutex, PowerPoint, estado ocupado |
| **DLL** | 3 | Registrar/desregistrar DLL, RegSvr32 |
| **MSI** | 4 | Códigos de saída, propriedades de tabela, transforms |
| **Active Setup** | 1 | Configurar entradas de Active Setup |
| **Edge** | 2 | Adicionar/remover políticas de extensão Edge |
| **Sistema** | 7 | Desktop, GPO, MS Updates, SCCM, Terminal Server |
| **Logging** | 1 | Escrever entradas de log |
| **Config** | 6 | Configuração, tabela de strings, histórico de adiamento |
| **Utilitários** | 8 | SendKeys, permissões, retry, encoding, templates |

## Opções do Cliente

```go
psadt.WithTimeout(10 * time.Minute)     // Timeout de execução de comando (padrão: 5 min)
psadt.WithPSPath("pwsh.exe")            // Caminho customizado do executável PowerShell
psadt.WithPowerShell7()                 // Usar PowerShell 7+ (pwsh.exe)
psadt.WithMinModuleVersion("4.1.0")     // Versão mínima do módulo PSADT
psadt.WithLogger(myLogger)              // *slog.Logger customizado para diagnósticos
```

## Exemplos

O diretório [`examples/`](examples/) contém programas completos e executáveis:

| Exemplo | Descrição |
|---|---|
| [`examples/install/`](examples/install/) | Instalação MSI completa com diálogo de boas-vindas, barra de progresso e configuração de registro |
| [`examples/uninstall/`](examples/uninstall/) | Busca e desinstalação de aplicação com limpeza |
| [`examples/dialog/`](examples/dialog/) | Interações UI: caixas de diálogo, balloon tips, prompts de instalação |
| [`examples/query/`](examples/query/) | Consultas de informação do sistema: ambiente, status admin, espaço em disco, usuários, serviços |

## Tratamento de Erros

Erros PSADT são automaticamente parseados em erros Go estruturados:

```go
result, err := session.StartMsiProcess(opts)
if err != nil {
    // Verificar tipos específicos de erro PSADT
    if parser.IsRebootRequired(err) {
        log.Println("Reboot necessário após instalação")
    } else if parser.IsUserCancelled(err) {
        log.Println("Usuário cancelou a instalação")
    } else {
        log.Fatalf("Instalação falhou: %v", err)
    }
}
```

## Contribuindo

Contribuições são bem-vindas. Por favor, abra uma issue primeiro para discutir o que você gostaria de alterar.

---

## License / Licença

MIT — see / veja [LICENSE](LICENSE).
