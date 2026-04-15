# Example: dialog

[English](#english) | [Português](#português)

---

## English

### What This Example Does

Demonstrates the **UI dialog capabilities** of `go-psadt`. It shows three different types of user interaction surfaces provided by PSAppDeployToolkit:

1. **Dialog Box** — a standard Windows message box with configurable buttons and icon
2. **Balloon Tip** — a Windows system tray notification
3. **Installation Prompt** — a PSADT-branded modal prompt with custom message and alignment

### File

```
examples/dialog/main.go
```

### Key Concepts Demonstrated

| Concept | Code element | Description |
|---|---|---|
| Dialog box | `session.ShowDialogBox(types.DialogBoxOptions{...})` | Displays a Win32 `MessageBox`-style dialog; returns the button clicked as a string (e.g. `"Yes"`, `"No"`) |
| Button set | `types.ButtonsYesNo` | Enum controlling which buttons appear (`YesNo`, `OkCancel`, `AbortRetryIgnore`, etc.) |
| Icon type | `types.IconQuestion` | Icon shown in the dialog (`IconQuestion`, `IconWarning`, `IconError`, `IconInformation`) |
| Balloon tip | `session.ShowBalloonTip(types.BalloonTipOptions{...})` | Posts a notification to the Windows system tray |
| Balloon icon | `types.BalloonInfo` | Icon for the balloon (`BalloonInfo`, `BalloonWarning`, `BalloonError`, `BalloonNone`) |
| Install prompt | `session.ShowInstallationPrompt(types.PromptOptions{...})` | PSADT-branded modal dialog; returns a `*types.PromptResult` |
| Text alignment | `types.AlignCenter` | Controls message alignment (`AlignLeft`, `AlignCenter`, `AlignRight`) |

### Code Walkthrough

```go
// 1. No explicit timeout — uses the default (5 minutes)
client, err := psadt.NewClient()

// 2. Interactive session for UI dialogs
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployInstall,
    DeployMode:     types.DeployModeInteractive,
    AppVendor:      "Demo",
    AppName:        "Dialog Demo",
    AppVersion:     "1.0",
})

// 3. Show a Yes/No dialog box — result is the button label string
result, err := session.ShowDialogBox(types.DialogBoxOptions{
    Title:   "Confirmation",
    Text:    "Do you want to proceed with the installation?",
    Buttons: types.ButtonsYesNo,
    Icon:    types.IconQuestion,
})
fmt.Printf("User chose: %s\n", result)  // "Yes" or "No"

// 4. Balloon tip — non-blocking system tray notification
session.ShowBalloonTip(types.BalloonTipOptions{
    BalloonTipText:  "Installation is starting...",
    BalloonTipTitle: "Widget Pro Setup",
    BalloonTipIcon:  types.BalloonInfo,
})

// 5. Installation prompt — PSADT-branded modal dialog
promptResult, err := session.ShowInstallationPrompt(types.PromptOptions{
    Title:            "Ready to Install",
    Message:          "Click OK to begin or Cancel to abort.",
    MessageAlignment: types.AlignCenter,
})
fmt.Printf("Prompt result: %+v\n", promptResult)
```

### Dialog Types Comparison

| Type | Blocking? | Return value | Best use |
|---|---|---|---|
| `ShowDialogBox` | Yes | `string` (button label) | Simple confirmations, warnings |
| `ShowBalloonTip` | No (fire-and-forget) | none | Status notifications |
| `ShowInstallationPrompt` | Yes | `*types.PromptResult` | PSADT-branded decisions with rich formatting |
| `ShowInstallationWelcome` | Yes | none (blocks until user confirms) | Pre-installation process closure |
| `ShowInstallationProgress` | No (updates a running window) | none | Long-running operations |

### PSADT Functions Used

| Go method | PSADT cmdlet |
|---|---|
| `client.OpenSession` | `Open-ADTSession` |
| `session.ShowDialogBox` | `Show-ADTDialogBox` |
| `session.ShowBalloonTip` | `Show-ADTBalloonTip` |
| `session.ShowInstallationPrompt` | `Show-ADTInstallationPrompt` |
| `session.Close` | `Close-ADTSession` |

### Running

```powershell
# Must run as Administrator in an interactive desktop session
# (UI dialogs will not appear in headless/service contexts)
cd examples\dialog
go run main.go
```

> **Note:** UI dialogs require an interactive Windows desktop session. Running as a Windows Service or in a non-interactive context will cause the dialogs to fail or be suppressed.

---

## Português

### O que Este Exemplo Faz

Demonstra as **capacidades de diálogos de UI** do `go-psadt`. Mostra três tipos diferentes de superfícies de interação com o usuário fornecidas pelo PSAppDeployToolkit:

1. **Dialog Box** — uma caixa de mensagem Windows padrão com botões e ícone configuráveis
2. **Balloon Tip** — uma notificação na bandeja do sistema Windows
3. **Installation Prompt** — um prompt modal com a marca PSADT e mensagem e alinhamento personalizados

### Arquivo

```
examples/dialog/main.go
```

### Conceitos-Chave Demonstrados

| Conceito | Elemento de código | Descrição |
|---|---|---|
| Caixa de diálogo | `session.ShowDialogBox(types.DialogBoxOptions{...})` | Exibe um diálogo estilo `MessageBox` Win32; retorna o botão clicado como string (ex.: `"Yes"`, `"No"`) |
| Conjunto de botões | `types.ButtonsYesNo` | Enum que controla quais botões aparecem (`YesNo`, `OkCancel`, `AbortRetryIgnore`, etc.) |
| Tipo de ícone | `types.IconQuestion` | Ícone exibido no diálogo (`IconQuestion`, `IconWarning`, `IconError`, `IconInformation`) |
| Balloon tip | `session.ShowBalloonTip(types.BalloonTipOptions{...})` | Posta uma notificação na bandeja do sistema Windows |
| Ícone balloon | `types.BalloonInfo` | Ícone do balloon (`BalloonInfo`, `BalloonWarning`, `BalloonError`, `BalloonNone`) |
| Prompt de instalação | `session.ShowInstallationPrompt(types.PromptOptions{...})` | Diálogo modal com marca PSADT; retorna `*types.PromptResult` |
| Alinhamento de texto | `types.AlignCenter` | Controla alinhamento da mensagem (`AlignLeft`, `AlignCenter`, `AlignRight`) |

### Walkthrough do Código

```go
// 1. Sem timeout explícito — usa o padrão (5 minutos)
client, err := psadt.NewClient()

// 2. Sessão interativa para diálogos de UI
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployInstall,
    DeployMode:     types.DeployModeInteractive,
    AppVendor:      "Demo",
    AppName:        "Dialog Demo",
    AppVersion:     "1.0",
})

// 3. Exibe diálogo Sim/Não — resultado é a string do botão
result, err := session.ShowDialogBox(types.DialogBoxOptions{
    Title:   "Confirmação",
    Text:    "Deseja prosseguir com a instalação?",
    Buttons: types.ButtonsYesNo,
    Icon:    types.IconQuestion,
})
fmt.Printf("Usuário escolheu: %s\n", result)  // "Yes" ou "No"

// 4. Balloon tip — notificação na bandeja, não bloqueante
session.ShowBalloonTip(types.BalloonTipOptions{
    BalloonTipText:  "A instalação está começando...",
    BalloonTipTitle: "Widget Pro Setup",
    BalloonTipIcon:  types.BalloonInfo,
})

// 5. Prompt de instalação — diálogo modal com marca PSADT
promptResult, err := session.ShowInstallationPrompt(types.PromptOptions{
    Title:            "Pronto para Instalar",
    Message:          "Clique em OK para iniciar ou Cancelar para abortar.",
    MessageAlignment: types.AlignCenter,
})
fmt.Printf("Resultado do prompt: %+v\n", promptResult)
```

### Comparação dos Tipos de Diálogo

| Tipo | Bloqueante? | Retorno | Melhor uso |
|---|---|---|---|
| `ShowDialogBox` | Sim | `string` (label do botão) | Confirmações simples, avisos |
| `ShowBalloonTip` | Não (fire-and-forget) | nenhum | Notificações de status |
| `ShowInstallationPrompt` | Sim | `*types.PromptResult` | Decisões com marca PSADT e formatação rica |
| `ShowInstallationWelcome` | Sim | nenhum (bloqueia até confirmação) | Fechamento de processos pré-instalação |
| `ShowInstallationProgress` | Não (atualiza janela em andamento) | nenhum | Operações de longa duração |

### Funções PSADT Utilizadas

| Método Go | Cmdlet PSADT |
|---|---|
| `client.OpenSession` | `Open-ADTSession` |
| `session.ShowDialogBox` | `Show-ADTDialogBox` |
| `session.ShowBalloonTip` | `Show-ADTBalloonTip` |
| `session.ShowInstallationPrompt` | `Show-ADTInstallationPrompt` |
| `session.Close` | `Close-ADTSession` |

### Executando

```powershell
# Deve executar como Administrador em uma sessão de desktop interativa
# (diálogos de UI não aparecem em contextos headless/serviço)
cd examples\dialog
go run main.go
```

> **Nota:** Diálogos de UI requerem uma sessão de desktop Windows interativa. Executar como Serviço Windows ou em contexto não-interativo fará os diálogos falharem ou serem suprimidos.
