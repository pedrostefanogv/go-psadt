# Example: install

[English](#english) | [Português](#português)

---

## English

### What This Example Does

Demonstrates a complete **MSI installation workflow** using `go-psadt`. It covers the full lifecycle of a typical enterprise software deployment:

1. Start a PSADT client and open a deployment session
2. Display a welcome dialog that prompts the user to close conflicting processes
3. Show a progress bar while the installer runs
4. Execute the MSI installer with `Start-ADTMsiProcess`
5. Write installation metadata to the Windows Registry
6. Close the progress bar and finish the session

### File

```
examples/install/main.go
```

### Key Concepts Demonstrated

| Concept | Code element | Description |
|---|---|---|
| Client creation | `psadt.NewClient(psadt.WithTimeout(...))` | Starts `powershell.exe`, imports PSAppDeployToolkit |
| Session configuration | `types.SessionConfig{DeploymentType, AppVendor, AppName, ...}` | Sets the app identity used by PSADT throughout the session |
| Welcome dialog | `session.ShowInstallationWelcome(types.WelcomeOptions{...})` | Prompts users to close `widget` and `widgethelper` processes; enforces a 5-minute countdown |
| Disk space check | `CheckDiskSpace: true` | PSADT automatically verifies sufficient disk space before proceeding |
| Progress bar | `session.ShowInstallationProgress(types.ProgressOptions{...})` | Displays a non-blocking progress window |
| MSI execution | `session.StartMsiProcess(types.MsiProcessOptions{Action: MsiInstall, ...})` | Runs `msiexec.exe` through PSADT with logging and error handling |
| Exit code capture | `result.ExitCode` | MSI exit code returned as a typed `int` |
| Registry write | `session.SetRegistryKey(types.SetRegistryKeyOptions{Type: RegString, ...})` | Creates `HKLM\SOFTWARE\Contoso\WidgetPro\Version` |
| Defer cleanup | `defer session.Close(0)` / `defer client.Close()` | Ensures session and process are always cleaned up |

### Code Walkthrough

```go
// 1. Create a client with a 10-minute timeout per command
client, err := psadt.NewClient(
    psadt.WithTimeout(10 * time.Minute),
)

// 2. Open a session — sets app identity for all PSADT calls
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployInstall,
    DeployMode:     types.DeployModeInteractive,
    AppVendor:      "Contoso",
    AppName:        "Widget Pro",
    AppVersion:     "2.0.0",
})

// 3. Welcome dialog — close widget.exe and widgethelper.exe
session.ShowInstallationWelcome(types.WelcomeOptions{
    CloseProcesses:          []types.ProcessDefinition{
        {Name: "widget"},
        {Name: "widgethelper"},
    },
    CloseProcessesCountdown: 300,  // 5-minute forced-close countdown
    CheckDiskSpace:          true,
})

// 4. Progress bar
session.ShowInstallationProgress(types.ProgressOptions{
    StatusMessage: "Installing Widget Pro 2.0...",
})

// 5. Run the MSI — exit code is returned as a Go int
result, err := session.StartMsiProcess(types.MsiProcessOptions{
    Action:   types.MsiInstall,
    FilePath: "WidgetPro.msi",
    PassThru: true,
})
fmt.Printf("MSI exit code: %d\n", result.ExitCode)

// 6. Write registry key
session.SetRegistryKey(types.SetRegistryKeyOptions{
    Key:   `HKLM\SOFTWARE\Contoso\WidgetPro`,
    Name:  "Version",
    Value: "2.0.0",
    Type:  types.RegString,
})

// 7. Close progress bar
session.CloseInstallationProgress()
```

### PSADT Functions Used

| Go method | PSADT cmdlet |
|---|---|
| `client.OpenSession` | `Open-ADTSession` |
| `session.ShowInstallationWelcome` | `Show-ADTInstallationWelcome` |
| `session.ShowInstallationProgress` | `Show-ADTInstallationProgress` |
| `session.StartMsiProcess` | `Start-ADTMsiProcess` |
| `session.SetRegistryKey` | `Set-ADTRegistryKey` |
| `session.CloseInstallationProgress` | `Close-ADTInstallationProgress` |
| `session.Close` | `Close-ADTSession` |

### Running

```powershell
# Must run as Administrator
cd examples\install
go run main.go
```

> **Note:** Adapt `FilePath: "WidgetPro.msi"` to point to a real MSI file before running.

---

## Português

### O que Este Exemplo Faz

Demonstra um **fluxo completo de instalação MSI** usando `go-psadt`. Cobre o ciclo de vida completo de um deployment corporativo típico:

1. Inicia um cliente PSADT e abre uma sessão de deployment
2. Exibe um diálogo de boas-vindas que solicita ao usuário fechar processos conflitantes
3. Exibe uma barra de progresso enquanto o instalador é executado
4. Executa o instalador MSI com `Start-ADTMsiProcess`
5. Grava metadados de instalação no Registro do Windows
6. Fecha a barra de progresso e encerra a sessão

### Arquivo

```
examples/install/main.go
```

### Conceitos-Chave Demonstrados

| Conceito | Elemento de código | Descrição |
|---|---|---|
| Criação do client | `psadt.NewClient(psadt.WithTimeout(...))` | Inicia `powershell.exe`, importa PSAppDeployToolkit |
| Configuração de sessão | `types.SessionConfig{DeploymentType, AppVendor, AppName, ...}` | Define a identidade do app usada pelo PSADT durante toda a sessão |
| Diálogo de boas-vindas | `session.ShowInstallationWelcome(types.WelcomeOptions{...})` | Solicita fechar os processos `widget` e `widgethelper`; impõe uma contagem regressiva de 5 minutos |
| Verificação de espaço | `CheckDiskSpace: true` | O PSADT verifica automaticamente espaço em disco suficiente |
| Barra de progresso | `session.ShowInstallationProgress(types.ProgressOptions{...})` | Exibe uma janela de progresso não bloqueante |
| Execução MSI | `session.StartMsiProcess(types.MsiProcessOptions{Action: MsiInstall, ...})` | Executa `msiexec.exe` via PSADT com logging e tratamento de erros |
| Captura de exit code | `result.ExitCode` | Código de saída do MSI retornado como `int` tipado |
| Escrita no registro | `session.SetRegistryKey(types.SetRegistryKeyOptions{Type: RegString, ...})` | Cria `HKLM\SOFTWARE\Contoso\WidgetPro\Version` |
| Limpeza com defer | `defer session.Close(0)` / `defer client.Close()` | Garante que sessão e processo sejam sempre encerrados |

### Walkthrough do Código

```go
// 1. Cria client com timeout de 10 minutos por comando
client, err := psadt.NewClient(
    psadt.WithTimeout(10 * time.Minute),
)

// 2. Abre sessão — define identidade do app para todas as chamadas PSADT
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployInstall,
    DeployMode:     types.DeployModeInteractive,
    AppVendor:      "Contoso",
    AppName:        "Widget Pro",
    AppVersion:     "2.0.0",
})

// 3. Diálogo de boas-vindas — fecha widget.exe e widgethelper.exe
session.ShowInstallationWelcome(types.WelcomeOptions{
    CloseProcesses: []types.ProcessDefinition{
        {Name: "widget"},
        {Name: "widgethelper"},
    },
    CloseProcessesCountdown: 300,  // contagem regressiva de 5 minutos
    CheckDiskSpace:          true,
})

// 4. Barra de progresso
session.ShowInstallationProgress(types.ProgressOptions{
    StatusMessage: "Instalando Widget Pro 2.0...",
})

// 5. Executa o MSI — exit code retornado como int Go
result, err := session.StartMsiProcess(types.MsiProcessOptions{
    Action:   types.MsiInstall,
    FilePath: "WidgetPro.msi",
    PassThru: true,
})
fmt.Printf("Exit code MSI: %d\n", result.ExitCode)

// 6. Grava chave no registro
session.SetRegistryKey(types.SetRegistryKeyOptions{
    Key:   `HKLM\SOFTWARE\Contoso\WidgetPro`,
    Name:  "Version",
    Value: "2.0.0",
    Type:  types.RegString,
})

// 7. Fecha barra de progresso
session.CloseInstallationProgress()
```

### Funções PSADT Utilizadas

| Método Go | Cmdlet PSADT |
|---|---|
| `client.OpenSession` | `Open-ADTSession` |
| `session.ShowInstallationWelcome` | `Show-ADTInstallationWelcome` |
| `session.ShowInstallationProgress` | `Show-ADTInstallationProgress` |
| `session.StartMsiProcess` | `Start-ADTMsiProcess` |
| `session.SetRegistryKey` | `Set-ADTRegistryKey` |
| `session.CloseInstallationProgress` | `Close-ADTInstallationProgress` |
| `session.Close` | `Close-ADTSession` |

### Executando

```powershell
# Deve executar como Administrador
cd examples\install
go run main.go
```

> **Nota:** Adapte `FilePath: "WidgetPro.msi"` para apontar para um arquivo MSI real antes de executar.
