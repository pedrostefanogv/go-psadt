# Examples — go-psadt

[English](#english) | [Português](#português)

---

## English

### Overview

This directory contains four standalone, runnable Go programs demonstrating the main use cases of the `go-psadt` library. Each example is self-contained and can be compiled and run independently on a Windows machine with PSAppDeployToolkit installed.

### Prerequisites

- Windows 10/11 or Windows Server 2016+
- Go ≥ 1.21
- PSAppDeployToolkit ≥ 4.1.0 installed (`Install-Module PSAppDeployToolkit -Scope AllUsers`)
- Administrator privileges (required by PSADT for most operations)

### Examples at a Glance

| Directory | What it demonstrates | Key PSADT functions |
|---|---|---|
| [`install/`](install/) | Full MSI installation workflow with welcome dialog, progress bar, registry write | `Open-ADTSession`, `Show-ADTInstallationWelcome`, `Show-ADTInstallationProgress`, `Start-ADTMsiProcess`, `Set-ADTRegistryKey`, `Close-ADTInstallationProgress` |
| [`uninstall/`](uninstall/) | Application search and uninstallation with registry cleanup | `Open-ADTSession`, `Show-ADTInstallationWelcome`, `Get-ADTApplication`, `Uninstall-ADTApplication`, `Remove-ADTRegistryKey` |
| [`dialog/`](dialog/) | UI dialogs: confirmation box, balloon tip notification, installation prompt | `Show-ADTDialogBox`, `Show-ADTBalloonTip`, `Show-ADTInstallationPrompt` |
| [`query/`](query/) | System information queries without touching the filesystem | `Get-ADTEnvironment`, `Test-ADTCallerIsAdmin`, `Test-ADTNetworkConnection`, `Get-ADTFreeDiskSpace`, `Get-ADTLoggedOnUser`, `Get-ADTPendingReboot`, `Test-ADTServiceExists` |

### Running an Example

```powershell
# From the repository root — must run as Administrator
cd examples\install
go run main.go
```

Or build first:

```powershell
cd examples\install
go build -o install.exe .
.\install.exe
```

> **Note:** All examples use `//go:build windows` and will not compile on Linux or macOS.

### Architecture at a Glance

All examples follow the same three-step pattern:

```
1. psadt.NewClient(opts...)     → starts powershell.exe, imports PSAppDeployToolkit
2. client.OpenSession(cfg)      → calls Open-ADTSession, sets app metadata
3. session.Method(opts)         → calls individual PSADT functions
   session.Close(exitCode)      → calls Close-ADTSession
   client.Close()               → stops the PowerShell process
```

The `defer` pattern ensures proper cleanup even when errors occur:

```go
client, _ := psadt.NewClient(...)
defer client.Close()

session, _ := client.OpenSession(...)
defer session.Close(0)
```

### Error Handling Pattern

Every example demonstrates safe error handling:

```go
result, err := session.StartMsiProcess(opts)
if err != nil {
    if parser.IsRebootRequired(err) {
        os.Exit(3010)
    }
    log.Fatalf("failed: %v", err)
}
```

---

## Português

### Visão Geral

Este diretório contém quatro programas Go independentes e executáveis que demonstram os principais casos de uso da biblioteca `go-psadt`. Cada exemplo é autocontido e pode ser compilado e executado individualmente em uma máquina Windows com o PSAppDeployToolkit instalado.

### Pré-requisitos

- Windows 10/11 ou Windows Server 2016+
- Go ≥ 1.21
- PSAppDeployToolkit ≥ 4.1.0 instalado (`Install-Module PSAppDeployToolkit -Scope AllUsers`)
- Privilégios de administrador (exigidos pelo PSADT para a maioria das operações)

### Resumo dos Exemplos

| Diretório | O que demonstra | Principais funções PSADT |
|---|---|---|
| [`install/`](install/) | Fluxo completo de instalação MSI com diálogo de boas-vindas, barra de progresso e escrita no registro | `Open-ADTSession`, `Show-ADTInstallationWelcome`, `Show-ADTInstallationProgress`, `Start-ADTMsiProcess`, `Set-ADTRegistryKey`, `Close-ADTInstallationProgress` |
| [`uninstall/`](uninstall/) | Busca e desinstalação de aplicação com limpeza do registro | `Open-ADTSession`, `Show-ADTInstallationWelcome`, `Get-ADTApplication`, `Uninstall-ADTApplication`, `Remove-ADTRegistryKey` |
| [`dialog/`](dialog/) | Diálogos de UI: caixa de confirmação, notificação balloon tip, prompt de instalação | `Show-ADTDialogBox`, `Show-ADTBalloonTip`, `Show-ADTInstallationPrompt` |
| [`query/`](query/) | Consultas de informação do sistema sem tocar no filesystem | `Get-ADTEnvironment`, `Test-ADTCallerIsAdmin`, `Test-ADTNetworkConnection`, `Get-ADTFreeDiskSpace`, `Get-ADTLoggedOnUser`, `Get-ADTPendingReboot`, `Test-ADTServiceExists` |

### Executando um Exemplo

```powershell
# Na raiz do repositório — deve executar como Administrador
cd examples\install
go run main.go
```

Ou compilar antes:

```powershell
cd examples\install
go build -o install.exe .
.\install.exe
```

> **Nota:** Todos os exemplos usam `//go:build windows` e não compilarão em Linux ou macOS.

### Arquitetura Resumida

Todos os exemplos seguem o mesmo padrão de três etapas:

```
1. psadt.NewClient(opts...)     → inicia powershell.exe, importa PSAppDeployToolkit
2. client.OpenSession(cfg)      → chama Open-ADTSession, define metadados do app
3. session.Method(opts)         → chama funções individuais do PSADT
   session.Close(exitCode)      → chama Close-ADTSession
   client.Close()               → encerra o processo PowerShell
```

O padrão `defer` garante limpeza adequada mesmo em caso de erro:

```go
client, _ := psadt.NewClient(...)
defer client.Close()

session, _ := client.OpenSession(...)
defer session.Close(0)
```

### Padrão de Tratamento de Erros

Cada exemplo demonstra tratamento seguro de erros:

```go
result, err := session.StartMsiProcess(opts)
if err != nil {
    if parser.IsRebootRequired(err) {
        os.Exit(3010)
    }
    log.Fatalf("falha: %v", err)
}
```
