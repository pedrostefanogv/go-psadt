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

> [!WARNING]
> **⚠️ AI-Generated Project — Use at Your Own Risk**
>
> This project was **entirely generated and assisted by AI tools** (GitHub Copilot / Claude). It is currently in an **experimental/review phase and is NOT recommended for production use**. The code has not been fully audited, may contain bugs, and could behave unexpectedly in real-world environments.
>
> By using this project, you accept full responsibility for any consequences. If you are not comfortable using AI-generated code, **do not use this project**.
>
> ---
>
> **⚠️ Projeto Gerado por IA — Use por Sua Própria Conta e Risco**
>
> Este projeto foi **inteiramente gerado e assistido por ferramentas de IA** (GitHub Copilot / Claude). Atualmente está em **fase experimental/revisão e NÃO é recomendado para uso em produção**. O código não passou por auditoria completa, pode conter bugs e pode se comportar de forma inesperada em ambientes reais.
>
> Ao utilizar este projeto, você assume total responsabilidade por quaisquer consequências. Se não se sentir confortável com o uso de código gerado por IA, **não utilize este projeto**.

---

## Documentation Map / Mapa de Documentacao

Core docs:

- [README.md](README.md) - project overview and quick start
- [ARCHITECTURE.md](ARCHITECTURE.md) - full architecture, protocol, and package design
- [PLAN.md](PLAN.md) - implementation plan and design rationale

Examples hub:

- [examples/README.md](examples/README.md) - examples index and execution guidance
- [examples/build-and-run.ps1](examples/build-and-run.ps1) - build all examples on Windows

Example-specific docs and entry points:

- Install: [examples/install/README.md](examples/install/README.md) | [examples/install/main.go](examples/install/main.go)
- Uninstall: [examples/uninstall/README.md](examples/uninstall/README.md) | [examples/uninstall/main.go](examples/uninstall/main.go)
- Dialog: [examples/dialog/README.md](examples/dialog/README.md) | [examples/dialog/main.go](examples/dialog/main.go)
- Query: [examples/query/README.md](examples/query/README.md) | [examples/query/main.go](examples/query/main.go)
- GUI Lab: [examples/gui/README.md](examples/gui/README.md) | [examples/gui/main.go](examples/gui/main.go)

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
go get github.com/pedrostefanogv/go-psadt
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/pedrostefanogv/go-psadt"
    "github.com/pedrostefanogv/go-psadt/types"
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

    // Close-ADTSession may terminate the underlying PowerShell runner.
    // Session.Close still treats that expected shutdown as success.

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

The detailed architecture is documented in [ARCHITECTURE.md](ARCHITECTURE.md).

Quick summary:

- The library keeps a **persistent PowerShell process** per client.
- Commands are serialized through a mutex and exchanged via stdin/stdout.
- Responses use a JSON envelope with `<<<PSADT_BEGIN>>>` / `<<<PSADT_END>>>` markers.
- `Session.Close()` treats runner termination during `Close-ADTSession` as a successful session shutdown; callers that reuse clients should check `client.IsAlive()` and recreate the runner when needed.
- Internal design is split into three core packages: `internal/cmdbuilder`, `internal/runner`, `internal/parser`.

For complete package mapping, protocol details, and lifecycle diagrams, see [ARCHITECTURE.md](ARCHITECTURE.md).

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

| Example | Guide | Entry Point | Description |
|---|---|---|---|
| `install` | [examples/install/README.md](examples/install/README.md) | [examples/install/main.go](examples/install/main.go) | Full MSI installation with welcome dialog, progress bar, and registry configuration |
| `uninstall` | [examples/uninstall/README.md](examples/uninstall/README.md) | [examples/uninstall/main.go](examples/uninstall/main.go) | Application search and uninstallation with cleanup |
| `dialog` | [examples/dialog/README.md](examples/dialog/README.md) | [examples/dialog/main.go](examples/dialog/main.go) | UI interactions: dialog boxes, balloon tips, installation prompts |
| `query` | [examples/query/README.md](examples/query/README.md) | [examples/query/main.go](examples/query/main.go) | System information queries: environment, admin status, disk space, users, services |
| `gui` | [examples/gui/README.md](examples/gui/README.md) | [examples/gui/main.go](examples/gui/main.go) | Graphical UI lab (local web app) to test modals, prompt styles, alerts, theme preview, and toolkit install checks |

Examples index:

- [examples/README.md](examples/README.md)

Build all example executables on Windows:

```powershell
cd examples
.\build-and-run.ps1
```

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
go get github.com/pedrostefanogv/go-psadt
```

## Início Rápido

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/pedrostefanogv/go-psadt"
    "github.com/pedrostefanogv/go-psadt/types"
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

A arquitetura detalhada esta em [ARCHITECTURE.md](ARCHITECTURE.md).

Resumo rapido:

- A biblioteca mantém um **processo PowerShell persistente** por client.
- Os comandos são serializados por mutex e trafegam via stdin/stdout.
- As respostas usam envelope JSON com marcadores `<<<PSADT_BEGIN>>>` / `<<<PSADT_END>>>`.
- O design interno é dividido em três pacotes centrais: `internal/cmdbuilder`, `internal/runner`, `internal/parser`.

Para mapeamento completo de pacotes, detalhes do protocolo e diagramas de ciclo de vida, consulte [ARCHITECTURE.md](ARCHITECTURE.md).

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

| Exemplo | Guia | Arquivo de Entrada | Descrição |
|---|---|---|---|
| `install` | [examples/install/README.md](examples/install/README.md) | [examples/install/main.go](examples/install/main.go) | Instalação MSI completa com diálogo de boas-vindas, barra de progresso e configuração de registro |
| `uninstall` | [examples/uninstall/README.md](examples/uninstall/README.md) | [examples/uninstall/main.go](examples/uninstall/main.go) | Busca e desinstalação de aplicação com limpeza |
| `dialog` | [examples/dialog/README.md](examples/dialog/README.md) | [examples/dialog/main.go](examples/dialog/main.go) | Interações UI: caixas de diálogo, balloon tips, prompts de instalação |
| `query` | [examples/query/README.md](examples/query/README.md) | [examples/query/main.go](examples/query/main.go) | Consultas de informação do sistema: ambiente, status admin, espaço em disco, usuários, serviços |
| `gui` | [examples/gui/README.md](examples/gui/README.md) | [examples/gui/main.go](examples/gui/main.go) | Laboratório gráfico (web local) para testar modais, estilos de prompt, alertas, prévia de tema e checagem de instalação do toolkit |

Indice de exemplos:

- [examples/README.md](examples/README.md)

Compilar todos os executáveis de exemplo no Windows:

```powershell
cd examples
.\build-and-run.ps1
```

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
