# go-psadt — Plano de Implementação Completo

> **Biblioteca Go wrapper para PSAppDeployToolkit v4.1.x**
>
> Permite que aplicações Go orquestrem deployments Windows, exibam diálogos UI,
> gerenciem registro/serviços/filesystem e invoquem instaladores — tudo via uma
> API Go idiomática com type-safety.

---

## Sumário

- [1. Visão Geral](#1-visão-geral)
- [2. Pré-requisitos](#2-pré-requisitos)
- [3. Arquitetura](#3-arquitetura)
- [4. Estrutura de Diretórios](#4-estrutura-de-diretórios)
- [5. Sistema de Tipos](#5-sistema-de-tipos)
- [6. Personalização UI/UX](#6-personalização-uiux)
- [7. Fases de Implementação](#7-fases-de-implementação)
- [8. Mapeamento Completo de Funções PSADT → Go](#8-mapeamento-completo-de-funções-psadt--go)
- [9. Funções Internas Excluídas](#9-funções-internas-excluídas)
- [10. Exemplos de Uso da API](#10-exemplos-de-uso-da-api)
- [11. Estratégia de Testes](#11-estratégia-de-testes)
- [12. Decisões Arquiteturais](#12-decisões-arquiteturais)
- [13. Considerações Adicionais](#13-considerações-adicionais)
- [14. Referência PSADT](#14-referência-psadt)

---

## 1. Visão Geral

O **PSAppDeployToolkit (PSADT)** v4.1.x é um framework PowerShell com **135 funções exportadas** para automação de deployments de software Windows. Ele possui:

- UI Fluent (WPF) e Classic (WinForms) com diálogos de boas-vindas, progresso, prompts e restart
- Execução de processos (EXE, MSI, MSP) com controle fino de tokens, child processes e timeouts
- Gerenciamento de registro, filesystem, serviços, variáveis de ambiente, INI, shortcuts
- Arquitetura Client/Server que permite exibir UI na sessão do usuário mesmo quando rodando como SYSTEM (Intune/SCCM)
- Suporte a WIM, ZIP, Active Setup, Edge Extensions, Group Policy, SCCM tasks
- Sistema de logging, configuração e localização (multi-idioma)

A lib `go-psadt` encapsula **~102 funções públicas** (excluindo ~33 funções de framework interno) como métodos Go fortemente tipados, mantendo um processo PowerShell persistente com sessão ADT aberta.

---

## 2. Pré-requisitos

| Requisito | Versão | Notas |
|-----------|--------|-------|
| **Go** | ≥ 1.21 | Módulo Go com generics |
| **Windows** | 10/11 ou Server 2016+ | Lib é Windows-only |
| **PowerShell** | ≥ 5.1.14393.0 | Windows PowerShell ou PowerShell 7+ |
| **.NET Framework** | ≥ 4.7.2 | Para assemblies C# do PSADT |
| **PSAppDeployToolkit** | ≥ 4.1.0 | Via `Install-Module PSAppDeployToolkit` |

```powershell
# Instalar PSADT
Install-Module -Name PSAppDeployToolkit -Scope AllUsers
```

---

## 3. Arquitetura

### 3.1 Diagrama Geral

```
┌──────────────────────────────────────────────────────────────┐
│                      Aplicação Go                            │
│                                                              │
│  client := psadt.NewClient()                                 │
│  session := client.OpenSession(cfg)                          │
│  session.ShowInstallationWelcome(opts)                       │
│  session.StartMsiProcess(opts)                               │
│  session.Close(0)                                            │
└──────────────┬───────────────────────────────────────────────┘
               │
               │  API Go (métodos tipados)
               │
┌──────────────▼───────────────────────────────────────────────┐
│                    go-psadt Library                           │
│                                                              │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────┐ │
│  │ cmdbuilder   │  │   runner      │  │     parser          │ │
│  │              │  │              │  │                     │ │
│  │ Struct Go →  │  │ PS Process   │  │ JSON response →    │ │
│  │ PS command   │  │ persistente  │  │ Struct Go           │ │
│  │ string       │  │ stdin/stdout │  │                     │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬──────────────┘ │
│         │                 │                  │                │
└─────────┼─────────────────┼──────────────────┼────────────────┘
          │                 │                  │
          │    ┌────────────▼────────────┐     │
          └───►│   powershell.exe        │◄────┘
               │   (processo persistente)│
               │                        │
               │   Import-Module PSADT  │
               │   Open-ADTSession ...  │
               │   <comandos via stdin> │
               │   → JSON via stdout    │
               └────────────┬───────────┘
                            │
               ┌────────────▼───────────────────────────┐
               │  PSAppDeployToolkit Module (4.1.x)      │
               │                                        │
               │  ┌──────────────┐  ┌────────────────┐  │
               │  │ 135 funções  │  │ Assemblies C#  │  │
               │  │ Verb-ADTNoun │  │ PSADT.dll      │  │
               │  │              │  │ UI.dll         │  │
               │  └──────────────┘  │ ClientServer   │  │
               │                    └────────────────┘  │
               └────────────────────────────────────────┘
```

### 3.2 Camada de Transporte: PowerShell Runner

- Processo `powershell.exe` (ou `pwsh.exe`) persistente via `os/exec` com pipes `stdin`/`stdout`/`stderr`
- **Protocolo**: cada comando PS é enviado via `stdin` e retorna JSON estruturado via `stdout`
  - Delimitadores especiais para identificar início/fim de resposta (ex: `<<<PSADT_BEGIN>>>` / `<<<PSADT_END>>>`)
  - Wrapper `try/catch` em cada comando para capturar erros como JSON
- **Mutex interno** para serialização de chamadas (1 comando por vez no runner)
- **Heartbeat**: comando `$true` periódico para verificar se o processo está vivo
- **Timeout por comando** configurável
- **Detecção automática** de `powershell.exe` (5.1) vs `pwsh.exe` (7+)
- **Encoding UTF-8** com BOM para compatibilidade

### 3.3 Camada de Sessão: ADTSession Manager

- `Open-ADTSession` / `Close-ADTSession` encapsulados no ciclo de vida da `Session`
- Estado da sessão mantido no lado Go (`DeploymentType`, `DeployMode`, `AppName`, etc.)
- Propriedades do `ADTSession` acessíveis como struct Go após `OpenSession`

### 3.4 Camada de Funções

- Cada categoria PSADT = arquivo `.go` separado com métodos da `Session`
- Parâmetros como structs Go com tags para mapeamento automático
- Retornos tipados (`ProcessResult`, `InstalledApplication`, etc.)

### 3.5 Arquitetura Client/Server do PSADT (referência)

O PSADT 4.1 introduziu um modelo Client/Server que a lib Go herda automaticamente:

```
SYSTEM (Intune/SCCM)                    User Session
┌─────────────────────┐                ┌──────────────────────┐
│ go-psadt → PS Runner│                │ PSADT.Client.exe     │
│   → ServerInstance   │──AnonymousPipe──►  WPF/WinForms UI   │
│   (processo SYSTEM)  │  (encrypted)  │  Diálogos na sessão  │
│                     │                │  do usuário ativo    │
└─────────────────────┘                └──────────────────────┘
```

Quando a lib Go roda como SYSTEM e chama funções de UI (`Show-ADTInstallation*`), o PSADT automaticamente:
1. Lança `PSADT.ClientServer.Client.exe` na sessão do usuário ativo
2. Comunica via Anonymous Pipes com encryption ECDH
3. Exibe os diálogos no desktop do usuário

**A lib Go não precisa implementar nada disso** — é transparente via chamadas PowerShell ao módulo.

---

## 4. Estrutura de Diretórios

```
go-psadt/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
│
├── psadt.go                    # Package principal: NewClient(), Client, Option
├── session.go                  # Session lifecycle: OpenSession(), Close(), GetProperties()
│
│   ── Funções organizadas por categoria ──
├── ui.go                       # 7 funções: Show*Welcome/Prompt/Progress/RestartPrompt/DialogBox/BalloonTip + Close*Progress
├── process.go                  # 8 funções: Start*Process*/Block/Unblock
├── application.go              # 3 funções: Get/UninstallApplication, GetRunningProcesses
├── registry.go                 # 5 funções: Get/Set/RemoveRegistryKey, TestRegistryValue, AllUsersRegistryAction
├── filesystem.go               # 8 funções: Copy/RemoveFile*, NewFolder, RemoveFolder, ContentToCache
├── ini.go                      # 6 funções: Get/Set/RemoveIniValue, Get/Set/RemoveIniSection
├── environment.go              # 3 funções: Get/Set/RemoveEnvironmentVariable
├── shortcut.go                 # 3 funções: New/Set/GetShortcut
├── service.go                  # 5 funções: Start/StopService, Get/SetServiceStartMode, TestServiceExists
├── wim.go                      # 3 funções: Mount/DismountWimFile, NewZipFile
├── sysinfo.go                  # 8 funções: GetLoggedOnUser/FreeDiskSpace/PendingReboot/OSInfo/UserProfiles/FileVersion/ExecutableInfo/PEFileArch
├── environment.go              # 1 função: Client.GetEnvironment() → EnvironmentInfo (todas as ~90 variáveis PSADT)
├── checks.go                   # 9 funções: TestBattery/Admin/Network/PowerPoint/Microphone/UserIsBusy/Esp/Oobe/MutexAvailability
├── logging.go                  # 1 função: WriteLogEntry
├── config.go                   # 4 funções: GetConfig, SetConfig, GetStringTable, SetStringTable
├── activesetup.go              # 1 função: SetActiveSetup
├── edge.go                     # 2 funções: Add/RemoveEdgeExtension
├── dll.go                      # 3 funções: Register/UnregisterDll, InvokeRegSvr32
├── msi.go                      # 4 funções: GetMsiExitCodeMessage/MsiTableProperty, SetMsiProperty, NewMsiTransform
├── system.go                   # 5 funções: UpdateDesktop/GroupPolicy, InstallMSUpdates/SCCMSoftwareUpdates, InvokeSCCMTask
├── util.go                     # 9 funções: SendKeys, ConvertToNTAccount, DeferHistory, SetItemPermission, etc.
│
├── types/
│   ├── enums.go                # DeploymentType, DeployMode, DialogStyle, DialogPosition, Icons, Buttons, etc.
│   ├── exitcodes.go            # Constantes: ExitCodeSuccess, ExitCodeModuleError, etc. (60000-79999)
│   ├── session.go              # SessionConfig, SessionProperties
│   ├── process.go              # ProcessResult, ProcessDefinition, StartProcessOptions, etc.
│   ├── application.go          # InstalledApplication, GetApplicationOptions, RunningProcess
│   ├── ui.go                   # WelcomeOptions, PromptOptions, PromptResult, DialogBoxOptions, etc.
│   ├── uiconfig.go             # UIConfig, AssetsConfig, ToolkitConfig — personalização visual
│   ├── registry.go             # GetRegistryKeyOptions, SetRegistryKeyOptions, etc.
│   ├── filesystem.go           # CopyFileOptions, RemoveFolderOptions, etc.
│   ├── ini.go                  # (funções com params simples — sem struct extra)
│   ├── service.go              # ServiceOptions, ServiceStartMode
│   ├── sysinfo.go              # LoggedOnUser, PendingRebootInfo, OSInfo, BatteryInfo, UserProfile, etc.
│   ├── environment.go          # EnvironmentInfo, ToolkitInfo, CultureInfo, SystemPaths, DomainInfo, OSEnvironment, ProcessInfo, PSVersionInfo, PermissionsInfo, UsersInfo, OfficeInfo, MiscInfo
│   ├── shortcut.go             # NewShortcutOptions, ShortcutInfo
│   ├── wim.go                  # MountWimOptions, DismountWimOptions, NewZipOptions
│   ├── dll.go                  # RegSvr32Options
│   ├── msi.go                  # MsiTableOptions, SetMsiPropertyOptions, MsiTransformOptions
│   ├── activesetup.go          # ActiveSetupOptions
│   ├── edge.go                 # EdgeExtensionOptions
│   ├── logging.go              # LogEntryOptions, LogSeverity
│   └── permission.go           # ItemPermissionOptions
│
├── internal/
│   ├── runner/
│   │   ├── runner.go           # PowerShell process lifecycle (start/stop/restart/healthcheck)
│   │   ├── command.go          # Command execution: send PS string → receive JSON
│   │   ├── protocol.go         # Delimitadores, wrapper try/catch, JSON extraction
│   │   └── runner_test.go
│   ├── cmdbuilder/
│   │   ├── builder.go          # Struct de params Go → string de comando PowerShell
│   │   ├── escape.go           # Escape de strings, arrays, hashtables, scriptblocks
│   │   └── builder_test.go
│   └── parser/
│       ├── parser.go           # JSON → struct Go (encoding/json)
│       ├── errors.go           # PSADTError type, exit code mapping
│       └── parser_test.go
│
├── examples/
│   ├── install/main.go         # Exemplo: install completo com Welcome, Progress, MSI
│   ├── uninstall/main.go       # Exemplo: uninstall com GetApplication + Uninstall
│   ├── dialog/main.go          # Exemplo: todos os tipos de diálogos UI
│   └── query/main.go           # Exemplo: consultas (apps instalados, registry, disk space)
│
└── testdata/
    ├── responses/              # Fixtures JSON para testes unitários
    └── scripts/                # Scripts PS auxiliares para testes de integração
```

---

## 5. Sistema de Tipos

### 5.1 Enums

```go
// DeploymentType representa o tipo de deployment
type DeploymentType string
const (
    DeployInstall   DeploymentType = "Install"
    DeployUninstall DeploymentType = "Uninstall"
    DeployRepair    DeploymentType = "Repair"
)

// DeployMode representa o modo de interação
type DeployMode string
const (
    DeployModeAuto           DeployMode = "Auto"
    DeployModeInteractive    DeployMode = "Interactive"
    DeployModeNonInteractive DeployMode = "NonInteractive"
    DeployModeSilent         DeployMode = "Silent"
)

// DialogStyle representa o estilo visual dos diálogos
type DialogStyle string
const (
    DialogStyleFluent  DialogStyle = "Fluent"   // WPF moderno (padrão v4.1)
    DialogStyleClassic DialogStyle = "Classic"  // WinForms legado
)

// DialogPosition representa a posição do diálogo na tela
type DialogPosition string
const (
    DialogPositionDefault     DialogPosition = "Default"
    DialogPositionTopLeft     DialogPosition = "TopLeft"
    DialogPositionTop         DialogPosition = "Top"
    DialogPositionTopRight    DialogPosition = "TopRight"
    DialogPositionTopCenter   DialogPosition = "TopCenter"
    DialogPositionCenter      DialogPosition = "Center"
    DialogPositionBottomLeft  DialogPosition = "BottomLeft"
    DialogPositionBottom      DialogPosition = "Bottom"
    DialogPositionBottomRight DialogPosition = "BottomRight"
)

// DialogSystemIcon representa ícones de diálogo
type DialogSystemIcon string
const (
    IconApplication DialogSystemIcon = "Application"
    IconAsterisk    DialogSystemIcon = "Asterisk"
    IconError       DialogSystemIcon = "Error"
    IconExclamation DialogSystemIcon = "Exclamation"
    IconHand        DialogSystemIcon = "Hand"
    IconInformation DialogSystemIcon = "Information"
    IconQuestion    DialogSystemIcon = "Question"
    IconShield      DialogSystemIcon = "Shield"
    IconWarning     DialogSystemIcon = "Warning"
    IconWinLogo     DialogSystemIcon = "WinLogo"
)

// DialogBoxButtons representa botões de diálogo padrão Windows
type DialogBoxButtons string
const (
    ButtonsOk                 DialogBoxButtons = "Ok"
    ButtonsOkCancel           DialogBoxButtons = "OkCancel"
    ButtonsAbortRetryIgnore   DialogBoxButtons = "AbortRetryIgnore"
    ButtonsYesNoCancel        DialogBoxButtons = "YesNoCancel"
    ButtonsYesNo              DialogBoxButtons = "YesNo"
    ButtonsRetryCancel        DialogBoxButtons = "RetryCancel"
    ButtonsCancelTryContinue  DialogBoxButtons = "CancelTryContinue"
)

// RegistryValueKind representa tipos de valor do registro Windows
type RegistryValueKind string
const (
    RegString       RegistryValueKind = "String"
    RegExpandString RegistryValueKind = "ExpandString"
    RegBinary       RegistryValueKind = "Binary"
    RegDWord        RegistryValueKind = "DWord"
    RegMultiString  RegistryValueKind = "MultiString"
    RegQWord        RegistryValueKind = "QWord"
)

// ProcessWindowStyle representa estilos de janela de processo
type ProcessWindowStyle string
const (
    WindowNormal    ProcessWindowStyle = "Normal"
    WindowHidden    ProcessWindowStyle = "Hidden"
    WindowMaximized ProcessWindowStyle = "Maximized"
    WindowMinimized ProcessWindowStyle = "Minimized"
)

// EnvironmentVariableTarget representa o escopo da variável de ambiente
type EnvironmentVariableTarget string
const (
    EnvTargetProcess EnvironmentVariableTarget = "Process"
    EnvTargetUser    EnvironmentVariableTarget = "User"
    EnvTargetMachine EnvironmentVariableTarget = "Machine"
)

// ServiceStartMode representa modos de inicialização de serviço
type ServiceStartMode string
const (
    ServiceAutomatic             ServiceStartMode = "Automatic"
    ServiceManual                ServiceStartMode = "Manual"
    ServiceDisabled              ServiceStartMode = "Disabled"
    ServiceAutomaticDelayedStart ServiceStartMode = "AutomaticDelayedStart"
)

// MsiAction representa ações MSI
type MsiAction string
const (
    MsiInstall     MsiAction = "Install"
    MsiUninstall   MsiAction = "Uninstall"
    MsiPatch       MsiAction = "Patch"
    MsiRepair      MsiAction = "Repair"
    MsiActiveSetup MsiAction = "ActiveSetup"
)

// NameMatch representa o tipo de match para busca de aplicações
type NameMatch string
const (
    MatchContains NameMatch = "Contains"
    MatchExact    NameMatch = "Exact"
    MatchWildcard NameMatch = "Wildcard"
    MatchRegex    NameMatch = "Regex"
)

// ApplicationType filtra por tipo de instalador
type ApplicationType string
const (
    AppTypeAll ApplicationType = "All"
    AppTypeMSI ApplicationType = "MSI"
    AppTypeEXE ApplicationType = "EXE"
)

// LogSeverity representa severidade de log
type LogSeverity int
const (
    LogInfo    LogSeverity = 1
    LogWarning LogSeverity = 2
    LogError   LogSeverity = 3
)

// BalloonTipIcon representa ícones de balloon/toast
type BalloonTipIcon string
const (
    BalloonNone    BalloonTipIcon = "None"
    BalloonInfo    BalloonTipIcon = "Info"
    BalloonWarning BalloonTipIcon = "Warning"
    BalloonError   BalloonTipIcon = "Error"
)

// MessageAlignment para diálogos
type MessageAlignment string
const (
    AlignLeft   MessageAlignment = "Left"
    AlignCenter MessageAlignment = "Center"
    AlignRight  MessageAlignment = "Right"
)

// DialogMessageAlignment (alias para uso nos parâmetros PSADT)
type DialogMessageAlignment = MessageAlignment
```

### 5.2 Exit Codes

```go
const (
    ExitCodeSuccess                  = 0
    ExitCodeReboot3010               = 3010
    ExitCodeReboot1641               = 1641
    ExitCodeUserCancelled            = 1602
    ExitCodeScriptError              = 60001
    ExitCodeModuleImportError        = 60008
    ExitCodeExePreLaunchError        = 60010
    ExitCodeExeLaunchError           = 60011
    // Faixas reservadas
    ExitCodeBuiltInRangeStart        = 60000
    ExitCodeBuiltInRangeEnd          = 68999
    ExitCodeUserCustomRangeStart     = 69000
    ExitCodeUserCustomRangeEnd       = 69999
    ExitCodeExtensionCustomRangeStart = 70000
    ExitCodeExtensionCustomRangeEnd  = 79999
)
```

### 5.3 Structs Principais

```go
// ProcessResult é o retorno de StartProcess quando PassThru=true
type ProcessResult struct {
    ExitCode    int    `json:"ExitCode"`
    StdOut      string `json:"StdOut"`
    StdErr      string `json:"StdErr"`
    Interleaved string `json:"Interleaved"`
}

// ProcessDefinition define um processo para CloseProcesses/BlockExecution
type ProcessDefinition struct {
    Name        string `json:"Name"`
    Description string `json:"Description,omitempty"`
}

// InstalledApplication representa uma aplicação instalada no sistema
type InstalledApplication struct {
    PSPath             string `json:"PSPath"`
    PSParentPath       string `json:"PSParentPath"`
    PSChildName        string `json:"PSChildName"`
    ProductCode        string `json:"ProductCode"`
    DisplayName        string `json:"DisplayName"`
    DisplayVersion     string `json:"DisplayVersion"`
    UninstallString    string `json:"UninstallString"`
    QuietUninstallString string `json:"QuietUninstallString"`
    InstallSource      string `json:"InstallSource"`
    InstallLocation    string `json:"InstallLocation"`
    InstallDate        string `json:"InstallDate"`
    Publisher          string `json:"Publisher"`
    HelpLink           string `json:"HelpLink"`
    EstimatedSize      int64  `json:"EstimatedSize"`
    SystemComponent    int    `json:"SystemComponent"`
    WindowsInstaller   int    `json:"WindowsInstaller"`
    Is64BitApplication bool   `json:"Is64BitApplication"`
}

// LoggedOnUser representa um usuário logado no sistema
type LoggedOnUser struct {
    NTAccount string `json:"NTAccount"`
    SID       string `json:"SID"`
    IsAdmin   bool   `json:"IsAdmin"`
    SessionID int    `json:"SessionId"`
}

// UserProfile representa um perfil de usuário
type UserProfile struct {
    NTAccount   string `json:"NTAccount"`
    SID         string `json:"SID"`
    ProfilePath string `json:"ProfilePath"`
}

// PendingRebootInfo resultado de Get-ADTPendingReboot
type PendingRebootInfo struct {
    ComputerName       string `json:"ComputerName"`
    LastBootUpTime     string `json:"LastBootUpTime"`
    IsSystemRebootPending bool `json:"IsSystemRebootPending"`
    IsCBServicing      bool   `json:"IsCBServicing"`
    IsWindowsUpdate    bool   `json:"IsWindowsUpdate"`
    IsSCCMClientReboot bool   `json:"IsSCCMClientReboot"`
    IsFileRenameOps    bool   `json:"IsFileRenameOps"`
}

// SessionConfig configuração para abrir uma sessão ADT
type SessionConfig struct {
    // Deployment Properties
    AppVendor   string `ps:"AppVendor"`
    AppName     string `ps:"AppName"`
    AppVersion  string `ps:"AppVersion"`
    AppArch     string `ps:"AppArch"`
    AppLang     string `ps:"AppLang"`
    AppRevision string `ps:"AppRevision"`

    // Script Properties
    AppScriptVersion string `ps:"AppScriptVersion"`
    AppScriptDate    string `ps:"AppScriptDate"`
    AppScriptAuthor  string `ps:"AppScriptAuthor"`

    // Deployment Settings
    DeploymentType            DeploymentType    `ps:"DeploymentType"`
    DeployMode                DeployMode        `ps:"DeployMode"`
    RequireAdmin              bool              `ps:"RequireAdmin,switch"`
    TerminalServerMode        bool              `ps:"TerminalServerMode,switch"`
    DisableLogging            bool              `ps:"DisableLogging,switch"`
    SuppressRebootPassThru    bool              `ps:"SuppressRebootPassThru,switch"`

    // Application Settings
    AppProcessesToClose       []ProcessDefinition `ps:"AppProcessesToClose"`
    AppSuccessExitCodes       []int               `ps:"AppSuccessExitCodes"`
    AppRebootExitCodes        []int               `ps:"AppRebootExitCodes"`

    // Display Settings
    InstallName  string `ps:"InstallName"`
    InstallTitle string `ps:"InstallTitle"`
    LogName      string `ps:"LogName"`

    // Path Settings
    ScriptDirectory string `ps:"ScriptDirectory"`
    DirFiles        string `ps:"DirFiles"`
    DirSupportFiles string `ps:"DirSupportFiles"`

    // MSI Settings
    DefaultMsiFile               string   `ps:"DefaultMsiFile"`
    DefaultMstFile               string   `ps:"DefaultMstFile"`
    DefaultMspFiles              []string `ps:"DefaultMspFiles"`
    DisableDefaultMsiProcessList bool     `ps:"DisableDefaultMsiProcessList,switch"`
    ForceMsiDetection            bool     `ps:"ForceMsiDetection,switch"`

    // Detection Bypass
    ForceWimDetection  bool `ps:"ForceWimDetection,switch"`
    NoSessionDetection bool `ps:"NoSessionDetection,switch"`
    NoOobeDetection    bool `ps:"NoOobeDetection,switch"`
    NoProcessDetection bool `ps:"NoProcessDetection,switch"`
}

// SessionProperties propriedades read-only da sessão (ADTSession object)
type SessionProperties struct {
    // Additional Properties (read-only, geradas ao abrir sessão)
    CurrentDate     string `json:"CurrentDate"`
    CurrentDateTime string `json:"CurrentDateTime"`
    CurrentTime     string `json:"CurrentTime"`
    InstallPhase    string `json:"InstallPhase"`
    LogPath         string `json:"LogPath"`
    UseDefaultMsi   bool   `json:"UseDefaultMsi"`

    // Script Properties (preenchidas durante OpenSession)
    DeployAppScriptFriendlyName string `json:"DeployAppScriptFriendlyName"`
    DeployAppScriptParameters   string `json:"DeployAppScriptParameters"`
    DeployAppScriptVersion      string `json:"DeployAppScriptVersion"`
}

// PromptResult resultado de ShowInstallationPrompt
type PromptResult struct {
    ButtonClicked string `json:"ButtonClicked"`
    InputText     string `json:"InputText,omitempty"`
}

// DialogBoxResult resultado de ShowDialogBox
type DialogBoxResult string

// ShortcutInfo informações de um atalho
type ShortcutInfo struct {
    TargetPath       string `json:"TargetPath"`
    Arguments        string `json:"Arguments"`
    Description      string `json:"Description"`
    WorkingDirectory string `json:"WorkingDirectory"`
    WindowStyle      int    `json:"WindowStyle"`
    Hotkey           string `json:"Hotkey"`
    IconLocation     string `json:"IconLocation"`
    RunAsAdmin       bool   `json:"RunAsAdmin"`
}

// OSInfo informações do sistema operacional
type OSInfo struct {
    Name         string `json:"Name"`
    Version      string `json:"Version"`
    Architecture string `json:"Architecture"`
    ServicePack  string `json:"ServicePack"`
    BuildNumber  string `json:"BuildNumber"`
}

// BatteryInfo informações de bateria
type BatteryInfo struct {
    IsLaptop          bool `json:"IsLaptop"`
    IsUsingACPower    bool `json:"IsUsingACPower"`
    BatteryChargeLevel int  `json:"BatteryChargeLevel,omitempty"`
}
```

### 5.4 Environment Info (Variáveis PSADT)

O PSADT expõe ~90+ variáveis ao abrir sessão (`Export-ADTEnvironmentTableToSessionState`).
A lib Go encapsula todas em um struct hierárquico acessível via `Client.GetEnvironment()`.

```go
// EnvironmentInfo contém todas as variáveis de ambiente expostas pelo PSADT
// Ref: https://psappdeploytoolkit.com/docs/4.1.x/reference/variables
type EnvironmentInfo struct {
    Toolkit     ToolkitInfo     `json:"Toolkit"`
    Culture     CultureInfo     `json:"Culture"`
    Paths       SystemPaths     `json:"Paths"`
    Domain      DomainInfo      `json:"Domain"`
    OS          OSEnvironment   `json:"OS"`
    Process     ProcessInfo     `json:"Process"`
    PowerShell  PSVersionInfo   `json:"PowerShell"`
    Permissions PermissionsInfo `json:"Permissions"`
    Users       UsersInfo       `json:"Users"`
    Office      OfficeInfo      `json:"Office"`
    Misc        MiscInfo        `json:"Misc"`
}

// ToolkitInfo — metadados do toolkit
type ToolkitInfo struct {
    FriendlyName string `json:"FriendlyName"` // $appDeployMainScriptFriendlyName
    ShortName    string `json:"ShortName"`    // $appDeployToolkitName
    Version      string `json:"Version"`      // $appDeployMainScriptVersion
}

// CultureInfo — idioma e cultura do sistema
type CultureInfo struct {
    Language   string `json:"Language"`   // $currentLanguage (e.g. "EN")
    UILanguage string `json:"UILanguage"` // $currentUILanguage
}

// SystemPaths — caminhos do sistema e do usuário (~40 variáveis agrupadas)
type SystemPaths struct {
    // System
    ProgramFiles       string `json:"ProgramFiles"`       // $envProgramFiles
    ProgramFilesX86    string `json:"ProgramFilesX86"`    // $envProgramFilesX86
    ProgramData        string `json:"ProgramData"`        // $envProgramData
    SystemRoot         string `json:"SystemRoot"`         // $envSystemRoot
    SystemDrive        string `json:"SystemDrive"`        // $envSystemDrive
    System32Directory  string `json:"System32Directory"`  // $envSystem32Directory
    WinDir             string `json:"WinDir"`             // $envWinDir
    Temp               string `json:"Temp"`               // $envTemp
    CommonProgramFiles string `json:"CommonProgramFiles"` // $envCommonProgramFiles
    CommonProgramFilesX86 string `json:"CommonProgramFilesX86"` // $envCommonProgramFilesX86
    Public             string `json:"Public"`             // $envPublic

    // User-specific
    UserProfile        string `json:"UserProfile"`        // $envUserProfile
    AppData            string `json:"AppData"`            // $envAppData
    LocalAppData       string `json:"LocalAppData"`       // $envLocalAppData
    UserDesktop        string `json:"UserDesktop"`        // $envUserDesktop
    UserDocuments      string `json:"UserDocuments"`      // $envUserMyDocuments
    UserStartMenu      string `json:"UserStartMenu"`      // $envUserStartMenu
    UserStartMenuPrograms string `json:"UserStartMenuPrograms"` // $envUserStartMenuPrograms
    UserStartUp        string `json:"UserStartUp"`        // $envUserStartUp

    // Common (All Users)
    AllUsersProfile    string `json:"AllUsersProfile"`    // $envAllUsersProfile
    CommonDesktop      string `json:"CommonDesktop"`      // $envCommonDesktop
    CommonDocuments    string `json:"CommonDocuments"`    // $envCommonDocuments
    CommonStartMenu    string `json:"CommonStartMenu"`    // $envCommonStartMenu
    CommonStartMenuPrograms string `json:"CommonStartMenuPrograms"` // $envCommonStartMenuPrograms
    CommonStartUp      string `json:"CommonStartUp"`      // $envCommonStartUp
    CommonTemplates    string `json:"CommonTemplates"`    // $envCommonTemplates

    // Identity
    HomeDrive          string   `json:"HomeDrive"`        // $envHomeDrive
    HomePath           string   `json:"HomePath"`         // $envHomePath
    HomeShare          string   `json:"HomeShare"`        // $envHomeShare
    ComputerName       string   `json:"ComputerName"`     // $envComputerName
    ComputerNameFQDN   string   `json:"ComputerNameFQDN"` // $envComputerNameFQDN
    UserName           string   `json:"UserName"`         // $envUserName
    LogicalDrives      []string `json:"LogicalDrives"`    // $envLogicalDrives
    SystemRAM          int      `json:"SystemRAM"`        // $envSystemRAM (MB)
}

// DomainInfo — informações de domínio AD
type DomainInfo struct {
    IsMachinePartOfDomain  bool   `json:"IsMachinePartOfDomain"`  // $IsMachinePartOfDomain
    MachineADDomain        string `json:"MachineADDomain"`        // $envMachineADDomain
    MachineDNSDomain       string `json:"MachineDNSDomain"`       // $envMachineDNSDomain
    MachineWorkgroup       string `json:"MachineWorkgroup"`       // $envMachineWorkgroup
    MachineDomainController string `json:"MachineDomainController"` // $MachineDomainController
    UserDNSDomain          string `json:"UserDNSDomain"`          // $envUserDNSDomain
    UserDomain             string `json:"UserDomain"`             // $envUserDomain
    LogonServer            string `json:"LogonServer"`            // $envLogonServer
}

// OSEnvironment — informações do sistema operacional (expandido)
type OSEnvironment struct {
    Name            string `json:"Name"`            // $envOSName
    Version         string `json:"Version"`         // $envOSVersion
    VersionMajor    int    `json:"VersionMajor"`    // $envOSVersionMajor
    VersionMinor    int    `json:"VersionMinor"`    // $envOSVersionMinor
    VersionBuild    int    `json:"VersionBuild"`    // $envOSVersionBuild
    VersionRevision int    `json:"VersionRevision"` // $envOSVersionRevision
    Architecture    string `json:"Architecture"`    // $envOSArchitecture (32-Bit/64-Bit)
    ServicePack     string `json:"ServicePack"`     // $envOSServicePack
    ProductType     int    `json:"ProductType"`     // $envOSProductType (1/2/3)
    ProductTypeName string `json:"ProductTypeName"` // $envOSProductTypeName (Server/Workstation/DC)
    Is64Bit         bool   `json:"Is64Bit"`         // $Is64Bit
    IsServerOS      bool   `json:"IsServerOS"`      // $IsServerOS
    IsWorkStationOS bool   `json:"IsWorkStationOS"` // $IsWorkStationOS
    IsDomainControllerOS bool `json:"IsDomainControllerOS"` // $IsDomainControllerOS
}

// ProcessInfo — arquitetura do processo atual
type ProcessInfo struct {
    Is64BitProcess bool   `json:"Is64BitProcess"` // $Is64BitProcess
    Architecture   string `json:"Architecture"`   // $psArchitecture (x86/x64)
}

// PSVersionInfo — versões PowerShell e CLR/.NET
type PSVersionInfo struct {
    PSVersion          string `json:"PSVersion"`          // $envPSVersion
    PSVersionMajor     int    `json:"PSVersionMajor"`     // $envPSVersionMajor
    PSVersionMinor     int    `json:"PSVersionMinor"`     // $envPSVersionMinor
    PSVersionBuild     int    `json:"PSVersionBuild"`     // $envPSVersionBuild
    PSVersionRevision  int    `json:"PSVersionRevision"`  // $envPSVersionRevision
    CLRVersion         string `json:"CLRVersion"`         // $envCLRVersion
    CLRVersionMajor    int    `json:"CLRVersionMajor"`    // $envCLRVersionMajor
    CLRVersionMinor    int    `json:"CLRVersionMinor"`    // $envCLRVersionMinor
}

// PermissionsInfo — permissões e contas do processo atual
type PermissionsInfo struct {
    IsAdmin                   bool   `json:"IsAdmin"`                   // $IsAdmin
    IsLocalSystemAccount      bool   `json:"IsLocalSystemAccount"`      // $IsLocalSystemAccount
    IsLocalServiceAccount     bool   `json:"IsLocalServiceAccount"`     // $IsLocalServiceAccount
    IsNetworkServiceAccount   bool   `json:"IsNetworkServiceAccount"`   // $IsNetworkServiceAccount
    IsServiceAccount          bool   `json:"IsServiceAccount"`          // $IsServiceAccount
    IsProcessUserInteractive  bool   `json:"IsProcessUserInteractive"`  // $IsProcessUserInteractive
    SessionZero               bool   `json:"SessionZero"`               // $SessionZero
    ProcessNTAccount          string `json:"ProcessNTAccount"`          // $ProcessNTAccount
    ProcessNTAccountSID       string `json:"ProcessNTAccountSID"`       // $ProcessNTAccountSID
    CurrentProcessSID         string `json:"CurrentProcessSID"`         // $CurrentProcessSID
    LocalSystemNTAccount      string `json:"LocalSystemNTAccount"`      // $LocalSystemNTAccount
    LocalAdministratorsGroup  string `json:"LocalAdministratorsGroup"`  // $LocalAdministratorsGroup
    LocalUsersGroup           string `json:"LocalUsersGroup"`           // $LocalUsersGroup
}

// UsersInfo — informações de usuários logados
type UsersInfo struct {
    LoggedOnUserSessions      []LoggedOnUserSession `json:"LoggedOnUserSessions"`      // $LoggedOnUserSessions
    CurrentConsoleUserSession *LoggedOnUserSession  `json:"CurrentConsoleUserSession"` // $CurrentConsoleUserSession
    CurrentLoggedOnUserSession *LoggedOnUserSession `json:"CurrentLoggedOnUserSession"` // $CurrentLoggedOnUserSession
    RunAsActiveUser           *LoggedOnUserSession  `json:"RunAsActiveUser"`           // $RunAsActiveUser
    UsersLoggedOn             []string              `json:"UsersLoggedOn"`             // $UsersLoggedOn (NTAccount names)
}

// LoggedOnUserSession — detalhes de uma sessão de usuário
type LoggedOnUserSession struct {
    NTAccount        string `json:"NTAccount"`
    SID              string `json:"SID"`
    SessionID        int    `json:"SessionId"`
    IsConsoleSession bool   `json:"IsConsoleSession"`
    IsCurrentSession bool   `json:"IsCurrentSession"`
    IsAdmin          bool   `json:"IsAdmin"`
}

// OfficeInfo — informações do Microsoft Office instalado
type OfficeInfo struct {
    Bitness string `json:"Bitness"` // $envOfficeBitness (x86/x64)
    Channel string `json:"Channel"` // $envOfficeChannel (e.g. "Monthly Enterprise")
    Version string `json:"Version"` // $envOfficeVersion (e.g. "16.0.x.x")
}

// MiscInfo — variáveis diversas
type MiscInfo struct {
    RunningTaskSequence bool `json:"RunningTaskSequence"` // $RunningTaskSequence (SCCM task sequence?)
}
```

---

## 6. Fases de Implementação

### Fase 1: Core / Fundação

| # | Tarefa | Arquivo(s) | Descrição |
|---|--------|-----------|-----------|
| 1.1 | Criar módulo Go | `go.mod` | `module github.com/<org>/go-psadt` |
| 1.2 | PowerShell Runner | `internal/runner/*.go` | Processo PS persistente: start/stop/restart, pipes stdin/stdout, mutex, heartbeat, timeout, detecção `powershell.exe`/`pwsh.exe`, encoding UTF-8 |
| 1.3 | Protocolo de comunicação | `internal/runner/protocol.go` | Delimitadores `<<<PSADT_BEGIN>>>`/`<<<PSADT_END>>>`, wrapper `try/catch` por comando, extração de JSON da resposta |
| 1.4 | Gerador de comandos | `internal/cmdbuilder/*.go` | Struct Go → string PS: escape de strings/arrays/hashtables, `SwitchParameter` (bool → `-Flag`), suporte a tags `ps:"ParamName,switch"` |
| 1.5 | Parser de respostas | `internal/parser/*.go` | JSON → struct Go via `encoding/json`, tratamento de `$null`/arrays vazios, tipo `PSADTError` com exit code mapping |
| 1.6 | Sistema de tipos | `types/*.go` | Todos os enums, structs de opções, structs de retorno (ver seção 5) |

### Fase 2: Session Management (3 funções)

| # | Tarefa | Arquivo(s) | Descrição |
|---|--------|-----------|-----------|
| 2.1 | Client principal | `psadt.go` | `NewClient(opts ...Option) (*Client, error)` — cria runner PS, `Import-Module PSAppDeployToolkit`, valida versão, `Client.Close()`, `Client.IsElevated()` |
| 2.2 | Options pattern | `psadt.go` | `WithPSPath()`, `WithModuleVersion()`, `WithTimeout()`, `WithLogger()`, `WithPowerShell7()` |
| 2.3 | Session lifecycle | `session.go` | `Client.OpenSession(cfg SessionConfig) (*Session, error)` → `Open-ADTSession`, `Session.Close(exitCode int) error` → `Close-ADTSession`, `Session.GetProperties()` |
| 2.4 | Environment info | `environment.go` | `Client.GetEnvironment() (*EnvironmentInfo, error)` — coleta todas as ~90 variáveis PSADT (paths, OS, domain, permissions, office, etc.) e retorna como struct Go hierárquico. Funciona no nível do Client (não requer sessão aberta). |

### Fase 3: UI / Diálogos (7 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 3.1 | `Session.ShowInstallationWelcome(WelcomeOptions)` | `Show-ADTInstallationWelcome` | `error` |
| 3.2 | `Session.ShowInstallationPrompt(PromptOptions)` | `Show-ADTInstallationPrompt` | `(PromptResult, error)` |
| 3.3 | `Session.ShowInstallationProgress(ProgressOptions)` | `Show-ADTInstallationProgress` | `error` |
| 3.4 | `Session.CloseInstallationProgress()` | `Close-ADTInstallationProgress` | `error` |
| 3.5 | `Session.ShowInstallationRestartPrompt(RestartPromptOptions)` | `Show-ADTInstallationRestartPrompt` | `error` |
| 3.6 | `Session.ShowDialogBox(DialogBoxOptions)` | `Show-ADTDialogBox` | `(DialogBoxResult, error)` |
| 3.7 | `Session.ShowBalloonTip(BalloonTipOptions)` | `Show-ADTBalloonTip` | `error` |

### Fase 4: Process Execution (8 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 4.1 | `Session.StartProcess(StartProcessOptions)` | `Start-ADTProcess` | `(*ProcessResult, error)` |
| 4.2 | `Session.StartProcessAsUser(StartProcessAsUserOptions)` | `Start-ADTProcessAsUser` | `(*ProcessResult, error)` |
| 4.3 | `Session.StartMsiProcess(MsiProcessOptions)` | `Start-ADTMsiProcess` | `(*ProcessResult, error)` |
| 4.4 | `Session.StartMsiProcessAsUser(MsiProcessAsUserOptions)` | `Start-ADTMsiProcessAsUser` | `(*ProcessResult, error)` |
| 4.5 | `Session.StartMspProcess(MspProcessOptions)` | `Start-ADTMspProcess` | `(*ProcessResult, error)` |
| 4.6 | `Session.StartMspProcessAsUser(MspProcessAsUserOptions)` | `Start-ADTMspProcessAsUser` | `(*ProcessResult, error)` |
| 4.7 | `Session.BlockAppExecution([]ProcessDefinition, ...DialogPosition)` | `Block-ADTAppExecution` | `error` |
| 4.8 | `Session.UnblockAppExecution()` | `Unblock-ADTAppExecution` | `error` |

### Fase 5: Application Management (3 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 5.1 | `Session.GetApplication(GetApplicationOptions)` | `Get-ADTApplication` | `([]InstalledApplication, error)` |
| 5.2 | `Session.UninstallApplication(UninstallApplicationOptions)` | `Uninstall-ADTApplication` | `error` |
| 5.3 | `Session.GetRunningProcesses([]string)` | `Get-ADTRunningProcesses` | `([]RunningProcess, error)` |

### Fase 6: Registry (5 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 6.1 | `Session.GetRegistryKey(GetRegistryKeyOptions)` | `Get-ADTRegistryKey` | `(interface{}, error)` |
| 6.2 | `Session.SetRegistryKey(SetRegistryKeyOptions)` | `Set-ADTRegistryKey` | `error` |
| 6.3 | `Session.RemoveRegistryKey(RemoveRegistryKeyOptions)` | `Remove-ADTRegistryKey` | `error` |
| 6.4 | `Session.TestRegistryValue(TestRegistryValueOptions)` | `Test-ADTRegistryValue` | `(bool, error)` |
| 6.5 | `Session.InvokeAllUsersRegistryAction(string, ...AllUsersOptions)` | `Invoke-ADTAllUsersRegistryAction` | `error` |

### Fase 7: File System (8 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 7.1 | `Session.CopyFile(CopyFileOptions)` | `Copy-ADTFile` | `error` |
| 7.2 | `Session.CopyFileToUserProfiles(CopyFileToUserProfilesOptions)` | `Copy-ADTFileToUserProfiles` | `error` |
| 7.3 | `Session.RemoveFile(RemoveFileOptions)` | `Remove-ADTFile` | `error` |
| 7.4 | `Session.RemoveFileFromUserProfiles(RemoveFileFromUserProfilesOptions)` | `Remove-ADTFileFromUserProfiles` | `error` |
| 7.5 | `Session.NewFolder(string)` | `New-ADTFolder` | `error` |
| 7.6 | `Session.RemoveFolder(RemoveFolderOptions)` | `Remove-ADTFolder` | `error` |
| 7.7 | `Session.CopyContentToCache(CopyContentToCacheOptions)` | `Copy-ADTContentToCache` | `(string, error)` |
| 7.8 | `Session.RemoveContentFromCache(RemoveContentFromCacheOptions)` | `Remove-ADTContentFromCache` | `error` |

### Fase 8: INI Files (6 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 8.1 | `Session.GetIniValue(filePath, section, key)` | `Get-ADTIniValue` | `(string, error)` |
| 8.2 | `Session.SetIniValue(filePath, section, key, value, ...force)` | `Set-ADTIniValue` | `error` |
| 8.3 | `Session.RemoveIniValue(filePath, section, ...key)` | `Remove-ADTIniValue` | `error` |
| 8.4 | `Session.GetIniSection(filePath, section)` | `Get-ADTIniSection` | `(map[string]string, error)` |
| 8.5 | `Session.SetIniSection(filePath, section, content, ...overwrite)` | `Set-ADTIniSection` | `error` |
| 8.6 | `Session.RemoveIniSection(filePath, section)` | `Remove-ADTIniSection` | `error` |

### Fase 9: Environment Variables (3 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 9.1 | `Session.GetEnvironmentVariable(variable, ...target)` | `Get-ADTEnvironmentVariable` | `(string, error)` |
| 9.2 | `Session.SetEnvironmentVariable(variable, value, ...target)` | `Set-ADTEnvironmentVariable` | `error` |
| 9.3 | `Session.RemoveEnvironmentVariable(variable, ...target)` | `Remove-ADTEnvironmentVariable` | `error` |

### Fase 10: Shortcuts (3 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 10.1 | `Session.NewShortcut(NewShortcutOptions)` | `New-ADTShortcut` | `error` |
| 10.2 | `Session.SetShortcut(SetShortcutOptions)` | `Set-ADTShortcut` | `error` |
| 10.3 | `Session.GetShortcut(path)` | `Get-ADTShortcut` | `(*ShortcutInfo, error)` |

### Fase 11: Services (5 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 11.1 | `Session.StartServiceAndDependencies(ServiceOptions)` | `Start-ADTServiceAndDependencies` | `error` |
| 11.2 | `Session.StopServiceAndDependencies(ServiceOptions)` | `Stop-ADTServiceAndDependencies` | `error` |
| 11.3 | `Session.GetServiceStartMode(name)` | `Get-ADTServiceStartMode` | `(ServiceStartMode, error)` |
| 11.4 | `Session.SetServiceStartMode(name, mode)` | `Set-ADTServiceStartMode` | `error` |
| 11.5 | `Session.TestServiceExists(name)` | `Test-ADTServiceExists` | `(bool, error)` |

### Fase 12: WIM/ZIP (3 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 12.1 | `Session.MountWimFile(MountWimOptions)` | `Mount-ADTWimFile` | `(string, error)` |
| 12.2 | `Session.DismountWimFile(DismountWimOptions)` | `Dismount-ADTWimFile` | `error` |
| 12.3 | `Session.NewZipFile(NewZipOptions)` | `New-ADTZipFile` | `error` |

### Fase 13: System Info & Checks (17 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 13.1 | `Session.GetLoggedOnUser()` | `Get-ADTLoggedOnUser` | `([]LoggedOnUser, error)` |
| 13.2 | `Session.GetFreeDiskSpace(...drive)` | `Get-ADTFreeDiskSpace` | `(uint64, error)` |
| 13.3 | `Session.GetPendingReboot()` | `Get-ADTPendingReboot` | `(*PendingRebootInfo, error)` |
| 13.4 | `Session.GetOperatingSystemInfo()` | `Get-ADTOperatingSystemInfo` | `(*OSInfo, error)` |
| 13.5 | `Session.GetUserProfiles(...UserProfileOptions)` | `Get-ADTUserProfiles` | `([]UserProfile, error)` |
| 13.6 | `Session.GetFileVersion(filePath)` | `Get-ADTFileVersion` | `(string, error)` |
| 13.7 | `Session.GetExecutableInfo(filePath)` | `Get-ADTExecutableInfo` | `(*ExecutableInfo, error)` |
| 13.8 | `Session.GetPEFileArchitecture(filePath)` | `Get-ADTPEFileArchitecture` | `(string, error)` |
| 13.9 | `Session.GetWindowTitle(GetWindowTitleOptions)` | `Get-ADTWindowTitle` | `([]WindowTitle, error)` |
| 13.10 | `Session.TestBattery()` | `Test-ADTBattery` | `(*BatteryInfo, error)` |
| 13.11 | `Session.TestCallerIsAdmin()` | `Test-ADTCallerIsAdmin` | `(bool, error)` |
| 13.12 | `Session.TestNetworkConnection()` | `Test-ADTNetworkConnection` | `(bool, error)` |
| 13.13 | `Session.TestMutexAvailability(mutexName)` | `Test-ADTMutexAvailability` | `(bool, error)` |
| 13.14 | `Session.TestPowerPoint()` | `Test-ADTPowerPoint` | `(bool, error)` |
| 13.15 | `Session.TestMicrophoneInUse()` | `Test-ADTMicrophoneInUse` | `(bool, error)` |
| 13.16 | `Session.TestUserIsBusy()` | `Test-ADTUserIsBusy` | `(bool, error)` |
| 13.17 | `Session.TestEspActive()` | `Test-ADTEspActive` | `(bool, error)` |
| 13.18 | `Session.TestOobeCompleted()` | `Test-ADTOobeCompleted` | `(bool, error)` |

### Fase 14: DLL/COM, MSI Utils, Active Setup, Edge, System Ops (15 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 14.1 | `Session.RegisterDll(filePath)` | `Register-ADTDll` | `error` |
| 14.2 | `Session.UnregisterDll(filePath)` | `Unregister-ADTDll` | `error` |
| 14.3 | `Session.InvokeRegSvr32(RegSvr32Options)` | `Invoke-ADTRegSvr32` | `error` |
| 14.4 | `Session.GetMsiExitCodeMessage(exitCode)` | `Get-ADTMsiExitCodeMessage` | `(string, error)` |
| 14.5 | `Session.GetMsiTableProperty(MsiTableOptions)` | `Get-ADTMsiTableProperty` | `(map[string]string, error)` |
| 14.6 | `Session.SetMsiProperty(SetMsiPropertyOptions)` | `Set-ADTMsiProperty` | `error` |
| 14.7 | `Session.NewMsiTransform(MsiTransformOptions)` | `New-ADTMsiTransform` | `error` |
| 14.8 | `Session.SetActiveSetup(ActiveSetupOptions)` | `Set-ADTActiveSetup` | `error` |
| 14.9 | `Session.AddEdgeExtension(extensionID, ...EdgeExtensionOptions)` | `Add-ADTEdgeExtension` | `error` |
| 14.10 | `Session.RemoveEdgeExtension(extensionID)` | `Remove-ADTEdgeExtension` | `error` |
| 14.11 | `Session.UpdateDesktop()` | `Update-ADTDesktop` | `error` |
| 14.12 | `Session.UpdateGroupPolicy()` | `Update-ADTGroupPolicy` | `error` |
| 14.13 | `Session.InstallMSUpdates(directory)` | `Install-ADTMSUpdates` | `error` |
| 14.14 | `Session.InstallSCCMSoftwareUpdates()` | `Install-ADTSCCMSoftwareUpdates` | `error` |
| 14.15 | `Session.InvokeSCCMTask(SCCMTaskOptions)` | `Invoke-ADTSCCMTask` | `error` |

### Fase 15: Logging, Config, Utilities (10 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 15.1 | `Session.WriteLogEntry(LogEntryOptions)` | `Write-ADTLogEntry` | `error` |
| 15.2 | `Session.GetConfig()` | `Get-ADTConfig` | `(*ADTConfig, error)` |
| 15.3 | `Session.GetStringTable()` | `Get-ADTStringTable` | `(*ADTStringTable, error)` |
| 15.4 | `Session.SendKeys(SendKeysOptions)` | `Send-ADTKeys` | `error` |
| 15.5 | `Session.ConvertToNTAccountOrSID(input)` | `ConvertTo-ADTNTAccountOrSID` | `(string, error)` |
| 15.6 | `Session.GetDeferHistory()` | `Get-ADTDeferHistory` | `(*DeferHistory, error)` |
| 15.7 | `Session.SetDeferHistory(...)` | `Set-ADTDeferHistory` | `error` |
| 15.8 | `Session.ResetDeferHistory()` | `Reset-ADTDeferHistory` | `error` |
| 15.9 | `Session.SetItemPermission(ItemPermissionOptions)` | `Set-ADTItemPermission` | `error` |
| 15.10 | `Session.InvokeCommandWithRetries(RetryOptions)` | `Invoke-ADTCommandWithRetries` | `(interface{}, error)` |

### Fase 16: Terminal Server (2 funções)

| # | Método Go | Função PSADT | Retorno |
|---|-----------|-------------|---------|
| 16.1 | `Session.EnableTerminalServerInstallMode()` | `Enable-ADTTerminalServerInstallMode` | `error` |
| 16.2 | `Session.DisableTerminalServerInstallMode()` | `Disable-ADTTerminalServerInstallMode` | `error` |

**Total: ~102 funções públicas mapeadas**

---

## 7. Mapeamento Completo de Funções PSADT → Go

### Legenda de Status

- ✅ **Incluída** — será implementada como método público
- ⛔ **Excluída** — função interna do framework PSADT, não relevante para a API Go

| # | Função PSADT | Status | Método Go | Fase |
|---|-------------|--------|-----------|------|
| 1 | `Add-ADTEdgeExtension` | ✅ | `Session.AddEdgeExtension()` | 14 |
| 2 | `Add-ADTModuleCallback` | ⛔ | — | — |
| 3 | `Block-ADTAppExecution` | ✅ | `Session.BlockAppExecution()` | 4 |
| 4 | `Clear-ADTModuleCallback` | ⛔ | — | — |
| 5 | `Close-ADTInstallationProgress` | ✅ | `Session.CloseInstallationProgress()` | 3 |
| 6 | `Close-ADTSession` | ✅ | `Session.Close()` | 2 |
| 7 | `Complete-ADTFunction` | ⛔ | — | — |
| 8 | `Convert-ADTRegistryPath` | ⛔ | — | — |
| 9 | `Convert-ADTValueType` | ⛔ | — | — |
| 10 | `Convert-ADTValuesFromRemainingArguments` | ⛔ | — | — |
| 11 | `ConvertTo-ADTNTAccountOrSID` | ✅ | `Session.ConvertToNTAccountOrSID()` | 15 |
| 12 | `Copy-ADTContentToCache` | ✅ | `Session.CopyContentToCache()` | 7 |
| 13 | `Copy-ADTFile` | ✅ | `Session.CopyFile()` | 7 |
| 14 | `Copy-ADTFileToUserProfiles` | ✅ | `Session.CopyFileToUserProfiles()` | 7 |
| 15 | `Disable-ADTTerminalServerInstallMode` | ✅ | `Session.DisableTerminalServerInstallMode()` | 16 |
| 16 | `Dismount-ADTWimFile` | ✅ | `Session.DismountWimFile()` | 12 |
| 17 | `Enable-ADTTerminalServerInstallMode` | ✅ | `Session.EnableTerminalServerInstallMode()` | 16 |
| 18 | `Export-ADTEnvironmentTableToSessionState` | ✅ | `Client.GetEnvironment()` | 2 |
| 19 | `Get-ADTApplication` | ✅ | `Session.GetApplication()` | 5 |
| 20 | `Get-ADTBoundParametersAndDefaultValues` | ⛔ | — | — |
| 21 | `Get-ADTCommandTable` | ⛔ | — | — |
| 22 | `Get-ADTConfig` | ✅ | `Session.GetConfig()` | 15 |
| 23 | `Get-ADTDeferHistory` | ✅ | `Session.GetDeferHistory()` | 15 |
| 24 | `Get-ADTEnvironment` | ✅ | `Client.GetEnvironment()` (dados consumidos internamente) | 2 |
| 25 | `Get-ADTEnvironmentTable` | ✅ | `Client.GetEnvironment()` (dados consumidos internamente) | 2 |
| 26 | `Get-ADTEnvironmentVariable` | ✅ | `Session.GetEnvironmentVariable()` | 9 |
| 27 | `Get-ADTExecutableInfo` | ✅ | `Session.GetExecutableInfo()` | 13 |
| 28 | `Get-ADTFileVersion` | ✅ | `Session.GetFileVersion()` | 13 |
| 29 | `Get-ADTFreeDiskSpace` | ✅ | `Session.GetFreeDiskSpace()` | 13 |
| 30 | `Get-ADTIniSection` | ✅ | `Session.GetIniSection()` | 8 |
| 31 | `Get-ADTIniValue` | ✅ | `Session.GetIniValue()` | 8 |
| 32 | `Get-ADTLoggedOnUser` | ✅ | `Session.GetLoggedOnUser()` | 13 |
| 33 | `Get-ADTModuleCallback` | ⛔ | — | — |
| 34 | `Get-ADTMsiExitCodeMessage` | ✅ | `Session.GetMsiExitCodeMessage()` | 14 |
| 35 | `Get-ADTMsiTableProperty` | ✅ | `Session.GetMsiTableProperty()` | 14 |
| 36 | `Get-ADTObjectProperty` | ⛔ | — | — |
| 37 | `Get-ADTOperatingSystemInfo` | ✅ | `Session.GetOperatingSystemInfo()` | 13 |
| 38 | `Get-ADTPEFileArchitecture` | ✅ | `Session.GetPEFileArchitecture()` | 13 |
| 39 | `Get-ADTPendingReboot` | ✅ | `Session.GetPendingReboot()` | 13 |
| 40 | `Get-ADTPowerShellProcessPath` | ⛔ | — | — |
| 41 | `Get-ADTPresentationSettingsEnabledUsers` | ✅ | `Session.GetPresentationSettingsEnabledUsers()` | 13 |
| 42 | `Get-ADTRegistryKey` | ✅ | `Session.GetRegistryKey()` | 6 |
| 43 | `Get-ADTRunningProcesses` | ✅ | `Session.GetRunningProcesses()` | 5 |
| 44 | `Get-ADTServiceStartMode` | ✅ | `Session.GetServiceStartMode()` | 11 |
| 45 | `Get-ADTSession` | ✅ | `Session.GetProperties()` | 2 |
| 46 | `Get-ADTShortcut` | ✅ | `Session.GetShortcut()` | 10 |
| 47 | `Get-ADTStringTable` | ✅ | `Session.GetStringTable()` | 15 |
| 48 | `Get-ADTUniversalDate` | ✅ | `Session.GetUniversalDate()` | 15 |
| 49 | `Get-ADTUserNotificationState` | ✅ | `Session.GetUserNotificationState()` | 13 |
| 50 | `Get-ADTUserProfiles` | ✅ | `Session.GetUserProfiles()` | 13 |
| 51 | `Get-ADTWindowTitle` | ✅ | `Session.GetWindowTitle()` | 13 |
| 52 | `Initialize-ADTFunction` | ⛔ | — | — |
| 53 | `Initialize-ADTModule` | ⛔ | — (chamado internamente no NewClient) | — |
| 54 | `Install-ADTMSUpdates` | ✅ | `Session.InstallMSUpdates()` | 14 |
| 55 | `Install-ADTSCCMSoftwareUpdates` | ✅ | `Session.InstallSCCMSoftwareUpdates()` | 14 |
| 56 | `Invoke-ADTAllUsersRegistryAction` | ✅ | `Session.InvokeAllUsersRegistryAction()` | 6 |
| 57 | `Invoke-ADTCommandWithRetries` | ✅ | `Session.InvokeCommandWithRetries()` | 15 |
| 58 | `Invoke-ADTFunctionErrorHandler` | ⛔ | — | — |
| 59 | `Invoke-ADTObjectMethod` | ⛔ | — | — |
| 60 | `Invoke-ADTRegSvr32` | ✅ | `Session.InvokeRegSvr32()` | 14 |
| 61 | `Invoke-ADTSCCMTask` | ✅ | `Session.InvokeSCCMTask()` | 14 |
| 62 | `Mount-ADTWimFile` | ✅ | `Session.MountWimFile()` | 12 |
| 63 | `New-ADTErrorRecord` | ⛔ | — | — |
| 64 | `New-ADTFolder` | ✅ | `Session.NewFolder()` | 7 |
| 65 | `New-ADTMsiTransform` | ✅ | `Session.NewMsiTransform()` | 14 |
| 66 | `New-ADTShortcut` | ✅ | `Session.NewShortcut()` | 10 |
| 67 | `New-ADTTemplate` | ✅ | `Session.NewTemplate()` | 15 |
| 68 | `New-ADTValidateScriptErrorRecord` | ⛔ | — | — |
| 69 | `New-ADTZipFile` | ✅ | `Session.NewZipFile()` | 12 |
| 70 | `Open-ADTSession` | ✅ | `Client.OpenSession()` | 2 |
| 71 | `Out-ADTPowerShellEncodedCommand` | ✅ | `Session.OutPowerShellEncodedCommand()` | 15 |
| 72 | `Register-ADTDll` | ✅ | `Session.RegisterDll()` | 14 |
| 73 | `Remove-ADTContentFromCache` | ✅ | `Session.RemoveContentFromCache()` | 7 |
| 74 | `Remove-ADTEdgeExtension` | ✅ | `Session.RemoveEdgeExtension()` | 14 |
| 75 | `Remove-ADTEnvironmentVariable` | ✅ | `Session.RemoveEnvironmentVariable()` | 9 |
| 76 | `Remove-ADTFile` | ✅ | `Session.RemoveFile()` | 7 |
| 77 | `Remove-ADTFileFromUserProfiles` | ✅ | `Session.RemoveFileFromUserProfiles()` | 7 |
| 78 | `Remove-ADTFolder` | ✅ | `Session.RemoveFolder()` | 7 |
| 79 | `Remove-ADTHashtableNullOrEmptyValues` | ⛔ | — | — |
| 80 | `Remove-ADTIniSection` | ✅ | `Session.RemoveIniSection()` | 8 |
| 81 | `Remove-ADTIniValue` | ✅ | `Session.RemoveIniValue()` | 8 |
| 82 | `Remove-ADTInvalidFileNameChars` | ✅ | `Session.RemoveInvalidFileNameChars()` | 15 |
| 83 | `Remove-ADTModuleCallback` | ⛔ | — | — |
| 84 | `Remove-ADTRegistryKey` | ✅ | `Session.RemoveRegistryKey()` | 6 |
| 85 | `Reset-ADTDeferHistory` | ✅ | `Session.ResetDeferHistory()` | 15 |
| 86 | `Resolve-ADTErrorRecord` | ⛔ | — | — |
| 87 | `Send-ADTKeys` | ✅ | `Session.SendKeys()` | 15 |
| 88 | `Set-ADTActiveSetup` | ✅ | `Session.SetActiveSetup()` | 14 |
| 89 | `Set-ADTDeferHistory` | ✅ | `Session.SetDeferHistory()` | 15 |
| 90 | `Set-ADTEnvironmentVariable` | ✅ | `Session.SetEnvironmentVariable()` | 9 |
| 91 | `Set-ADTIniSection` | ✅ | `Session.SetIniSection()` | 8 |
| 92 | `Set-ADTIniValue` | ✅ | `Session.SetIniValue()` | 8 |
| 93 | `Set-ADTItemPermission` | ✅ | `Session.SetItemPermission()` | 15 |
| 94 | `Set-ADTMsiProperty` | ✅ | `Session.SetMsiProperty()` | 14 |
| 95 | `Set-ADTPowerShellCulture` | ✅ | `Session.SetPowerShellCulture()` | 15 |
| 96 | `Set-ADTRegistryKey` | ✅ | `Session.SetRegistryKey()` | 6 |
| 97 | `Set-ADTServiceStartMode` | ✅ | `Session.SetServiceStartMode()` | 11 |
| 98 | `Set-ADTShortcut` | ✅ | `Session.SetShortcut()` | 10 |
| 99 | `Show-ADTBalloonTip` | ✅ | `Session.ShowBalloonTip()` | 3 |
| 100 | `Show-ADTDialogBox` | ✅ | `Session.ShowDialogBox()` | 3 |
| 101 | `Show-ADTHelpConsole` | ✅ | `Session.ShowHelpConsole()` | 3 |
| 102 | `Show-ADTInstallationProgress` | ✅ | `Session.ShowInstallationProgress()` | 3 |
| 103 | `Show-ADTInstallationPrompt` | ✅ | `Session.ShowInstallationPrompt()` | 3 |
| 104 | `Show-ADTInstallationRestartPrompt` | ✅ | `Session.ShowInstallationRestartPrompt()` | 3 |
| 105 | `Show-ADTInstallationWelcome` | ✅ | `Session.ShowInstallationWelcome()` | 3 |
| 106 | `Start-ADTMsiProcess` | ✅ | `Session.StartMsiProcess()` | 4 |
| 107 | `Start-ADTMsiProcessAsUser` | ✅ | `Session.StartMsiProcessAsUser()` | 4 |
| 108 | `Start-ADTMspProcess` | ✅ | `Session.StartMspProcess()` | 4 |
| 109 | `Start-ADTMspProcessAsUser` | ✅ | `Session.StartMspProcessAsUser()` | 4 |
| 110 | `Start-ADTProcess` | ✅ | `Session.StartProcess()` | 4 |
| 111 | `Start-ADTProcessAsUser` | ✅ | `Session.StartProcessAsUser()` | 4 |
| 112 | `Start-ADTServiceAndDependencies` | ✅ | `Session.StartServiceAndDependencies()` | 11 |
| 113 | `Stop-ADTServiceAndDependencies` | ✅ | `Session.StopServiceAndDependencies()` | 11 |
| 114 | `Test-ADTBattery` | ✅ | `Session.TestBattery()` | 13 |
| 115 | `Test-ADTCallerIsAdmin` | ✅ | `Session.TestCallerIsAdmin()` | 13 |
| 116 | `Test-ADTEspActive` | ✅ | `Session.TestEspActive()` | 13 |
| 117 | `Test-ADTMSUpdates` | ✅ | `Session.TestMSUpdates()` | 13 |
| 118 | `Test-ADTMicrophoneInUse` | ✅ | `Session.TestMicrophoneInUse()` | 13 |
| 119 | `Test-ADTModuleInitialized` | ⛔ | — (usado internamente) | — |
| 120 | `Test-ADTMutexAvailability` | ✅ | `Session.TestMutexAvailability()` | 13 |
| 121 | `Test-ADTNetworkConnection` | ✅ | `Session.TestNetworkConnection()` | 13 |
| 122 | `Test-ADTOobeCompleted` | ✅ | `Session.TestOobeCompleted()` | 13 |
| 123 | `Test-ADTPowerPoint` | ✅ | `Session.TestPowerPoint()` | 13 |
| 124 | `Test-ADTRegistryValue` | ✅ | `Session.TestRegistryValue()` | 6 |
| 125 | `Test-ADTServiceExists` | ✅ | `Session.TestServiceExists()` | 11 |
| 126 | `Test-ADTSessionActive` | ⛔ | — (verificado internamente) | — |
| 127 | `Test-ADTUserIsBusy` | ✅ | `Session.TestUserIsBusy()` | 13 |
| 128 | `Unblock-ADTAppExecution` | ✅ | `Session.UnblockAppExecution()` | 4 |
| 129 | `Uninstall-ADTApplication` | ✅ | `Session.UninstallApplication()` | 5 |
| 130 | `Unregister-ADTDll` | ✅ | `Session.UnregisterDll()` | 14 |
| 131 | `Update-ADTDesktop` | ✅ | `Session.UpdateDesktop()` | 14 |
| 132 | `Update-ADTEnvironmentPsProvider` | ⛔ | — (PS-specific) | — |
| 133 | `Update-ADTGroupPolicy` | ✅ | `Session.UpdateGroupPolicy()` | 14 |
| 134 | `Write-ADTLogEntry` | ✅ | `Session.WriteLogEntry()` | 15 |

**Resumo: 135 funções PSADT → ~105 métodos Go públicos + ~30 funções internas excluídas**

---

## 8. Funções Internas Excluídas

As seguintes **33 funções** são de framework/infraestrutura interna do PSADT e **não serão expostas** na API Go. Elas são usadas internamente pelo módulo PowerShell para gerenciamento de estado, callbacks e tratamento de erros:

| Função | Motivo da Exclusão |
|--------|--------------------|
| `Add-ADTModuleCallback` | Callback PS interno para extensões |
| `Clear-ADTModuleCallback` | Callback PS interno |
| `Complete-ADTFunction` | Finalização interna de funções PS |
| `Convert-ADTRegistryPath` | Helper interno de conversão de path |
| `Convert-ADTValueType` | Helper interno de conversão de tipos |
| `Convert-ADTValuesFromRemainingArguments` | Parser interno de argumentos PS |
| `Get-ADTBoundParametersAndDefaultValues` | Introspecção PS interna |
| `Get-ADTCommandTable` | Tabela de comandos interna |

| `Get-ADTModuleCallback` | Callback PS interno |
| `Get-ADTObjectProperty` | Acesso genérico a propriedades .NET |
| `Get-ADTPowerShellProcessPath` | Helper para resolver path do PS |
| `Initialize-ADTFunction` | Inicialização interna de funções |
| `Initialize-ADTModule` | Inicialização do módulo (chamado pelo `NewClient` internamente) |
| `Invoke-ADTFunctionErrorHandler` | Handler de erro interno |
| `Invoke-ADTObjectMethod` | Invocação genérica de métodos .NET |
| `New-ADTErrorRecord` | Criação de ErrorRecord PS |
| `New-ADTValidateScriptErrorRecord` | Criação de ErrorRecord de validação |
| `Remove-ADTHashtableNullOrEmptyValues` | Limpeza interna de hashtables |
| `Remove-ADTModuleCallback` | Callback PS interno |
| `Resolve-ADTErrorRecord` | Resolução de ErrorRecord PS |
| `Test-ADTModuleInitialized` | Verificação interna de estado do módulo |
| `Test-ADTSessionActive` | Verificação interna de sessão (feita implicitamente) |
| `Update-ADTEnvironmentPsProvider` | Atualiza provider PS (PS-specific) |

---

## 9. Exemplos de Uso da API

### 9.1 Instalação Completa de Aplicação

```go
package main

import (
    "log"
    "github.com/org/go-psadt"
    "github.com/org/go-psadt/types"
)

func main() {
    // Criar client
    client, err := psadt.NewClient(
        psadt.WithMinModuleVersion("4.1.0"),
        psadt.WithTimeout(30 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Abrir sessão
    session, err := client.OpenSession(types.SessionConfig{
        AppName:        "Adobe Acrobat Reader",
        AppVersion:     "24.001.20604",
        AppVendor:      "Adobe",
        DeploymentType: types.DeployInstall,
        DeployMode:     types.DeployModeAuto,
        AppProcessesToClose: []types.ProcessDefinition{
            {Name: "AcroRd32", Description: "Adobe Acrobat Reader"},
            {Name: "Acrobat", Description: "Adobe Acrobat"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close(0)

    // == PRE-INSTALLATION ==

    // Mostrar diálogo de boas-vindas com deferral
    err = session.ShowInstallationWelcome(types.WelcomeOptions{
        CloseProcesses: []types.ProcessDefinition{
            {Name: "AcroRd32"},
            {Name: "Acrobat"},
        },
        AllowDefer:               true,
        DeferTimes:               3,
        BlockExecution:           true,
        CloseProcessesCountdown:  600,
        PersistPrompt:            true,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Mostrar progresso
    err = session.ShowInstallationProgress(types.ProgressOptions{
        StatusMessage: "Instalando Adobe Acrobat Reader...",
    })
    if err != nil {
        log.Fatal(err)
    }

    // == INSTALLATION ==

    // Desinstalar versão antiga
    err = session.UninstallApplication(types.UninstallApplicationOptions{
        Name:            []string{"Adobe Acrobat Reader"},
        ApplicationType: types.AppTypeMSI,
    })
    if err != nil {
        log.Printf("Aviso ao desinstalar versão antiga: %v", err)
    }

    // Instalar nova versão
    result, err := session.StartMsiProcess(types.MsiProcessOptions{
        Action:   types.MsiInstall,
        FilePath: "AcroRead.msi",
        Transforms: []string{"AcroRead.mst"},
        AdditionalArgumentList: []string{"ALLUSERS=1", "EULA_ACCEPT=YES"},
        PassThru: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("MSI exit code: %d", result.ExitCode)

    // == POST-INSTALLATION ==

    // Criar atalho no desktop
    err = session.NewShortcut(types.NewShortcutOptions{
        Path:       `C:\Users\Public\Desktop\Adobe Acrobat Reader.lnk`,
        TargetPath: `C:\Program Files\Adobe\Acrobat Reader\Reader\AcroRd32.exe`,
        IconIndex:  0,
        Description: "Adobe Acrobat Reader",
    })
    if err != nil {
        log.Printf("Aviso ao criar atalho: %v", err)
    }

    // Fechar progresso
    session.CloseInstallationProgress()

    // Balloon notification
    session.ShowBalloonTip(types.BalloonTipOptions{
        BalloonTipTitle: "Adobe Acrobat Reader",
        BalloonTipText:  "Instalação concluída com sucesso!",
        BalloonTipIcon:  types.BalloonInfo,
    })
}
```

### 9.2 Exibição de Diálogos UI

```go
// Prompt com input do usuário
result, err := session.ShowInstallationPrompt(types.PromptOptions{
    Title:          "Configuração",
    Message:        "Digite o nome do servidor:",
    RequestInput:   true,
    DefaultValue:   "server01.contoso.com",
    ButtonRightText: "OK",
    Icon:           types.IconQuestion,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Servidor: %s\n", result.InputText)

// Prompt Sim/Não com 3 botões
result, err = session.ShowInstallationPrompt(types.PromptOptions{
    Title:           "Confirmação",
    Message:         "Como você está se sentindo hoje?",
    ButtonLeftText:  "Bem",
    ButtonMiddleText: "Indiferente",
    ButtonRightText: "Mal",
    Icon:            types.IconQuestion,
})
fmt.Printf("Resposta: %s\n", result.ButtonClicked)

// Dialog box padrão Windows
dialogResult, err := session.ShowDialogBox(types.DialogBoxOptions{
    Title:         "Aviso",
    Text:          "A instalação levará 30 minutos. Deseja continuar?",
    Buttons:       types.ButtonsOkCancel,
    DefaultButton: "Second",
    Icon:          types.DialogIconExclamation,
    Timeout:       600,
})
fmt.Printf("Botão clicado: %s\n", dialogResult)

// Progresso com percentual
for i := 0; i <= 100; i += 10 {
    session.ShowInstallationProgress(types.ProgressOptions{
        StatusMessage:      fmt.Sprintf("Processando... %d%%", i),
        StatusBarPercentage: float64(i),
    })
    time.Sleep(1 * time.Second)
}
session.CloseInstallationProgress()
```

### 9.3 Consultas e Verificações

```go
// Listar aplicações instaladas
apps, err := session.GetApplication(types.GetApplicationOptions{
    Name:            []string{"Microsoft Office"},
    NameMatch:       types.MatchContains,
    ApplicationType: types.AppTypeMSI,
})
for _, app := range apps {
    fmt.Printf("  %s v%s (%s)\n", app.DisplayName, app.DisplayVersion, app.Publisher)
}

// Verificar espaço em disco
freeSpace, _ := session.GetFreeDiskSpace()
fmt.Printf("Espaço livre: %d MB\n", freeSpace)

// Verificar reboot pendente
reboot, _ := session.GetPendingReboot()
if reboot.IsSystemRebootPending {
    fmt.Println("ATENÇÃO: Reboot pendente!")
}

// Verificar registro
val, _ := session.GetRegistryKey(types.GetRegistryKeyOptions{
    LiteralPath: `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion`,
    Name:        "ProgramFilesDir",
})
fmt.Printf("ProgramFiles: %v\n", val)

// Verificar se é admin
isAdmin, _ := session.TestCallerIsAdmin()
fmt.Printf("É admin: %v\n", isAdmin)

// Verificar rede
hasNetwork, _ := session.TestNetworkConnection()
fmt.Printf("Rede disponível: %v\n", hasNetwork)
```

---

## 10. Estratégia de Testes

### 10.1 Testes Unitários

Cada arquivo `*_test.go` valida:

1. **Geração de comandos PS** — verifica que o `cmdbuilder` gera a string PS correta para cada combinação de parâmetros
2. **Parsing de respostas JSON** — fixtures em `testdata/responses/` com JSON mockado do PSADT
3. **Tratamento de erros** — JSON de erro PS é mapeado corretamente para `error` Go
4. **Conversão de tipos** — enums, arrays, hashtables são serializados corretamente

```go
// Exemplo: test de geração de comando
func TestCmdBuilder_ShowInstallationWelcome(t *testing.T) {
    opts := types.WelcomeOptions{
        CloseProcesses: []types.ProcessDefinition{{Name: "winword"}},
        AllowDefer:     true,
        DeferTimes:     3,
        BlockExecution: true,
    }
    cmd := cmdbuilder.Build("Show-ADTInstallationWelcome", opts)
    expected := `Show-ADTInstallationWelcome -CloseProcesses @{Name='winword'} -AllowDefer -DeferTimes 3 -BlockExecution`
    assert.Equal(t, expected, cmd)
}
```

### 10.2 Testes de Integração

- Requerem **PSADT instalado** no sistema
- Tag `//go:build integration`
- Validam ciclo completo: `NewClient` → `OpenSession` → funções → `Close`
- Executados em CI com Windows runner

### 10.3 Teste de Cobertura

Script que compara a lista de funções exportadas do PSADT vs métodos Go implementados:
```go
func TestFunctionCoverage(t *testing.T) {
    // Executa: Get-Command -Module PSAppDeployToolkit | Select Name
    // Compara com métodos públicos do Session
    // Deve ser 102/102 (excluindo as 33 internas)
}
```

### 10.4 Testes de UI (Manual)

- Validação visual de cada diálogo: Welcome, Prompt, Progress, RestartPrompt, DialogBox, BalloonTip
- Scripts de exemplo em `examples/dialog/`

### 10.5 Benchmark

```go
func BenchmarkRunnerCommand(b *testing.B) {
    // Mede overhead do runner PS por comando
    // Esperado: <10ms com processo persistente
}
```

---

## 11. Decisões Arquiteturais

### 11.1 Processo PS Persistente (vs Efêmero)

**Decisão:** Manter um único processo `powershell.exe` durante toda a vida do `Client`.

**Motivo:** Evita overhead de ~500ms por spawn de processo. A sessão ADT (`Open-ADTSession`) fica aberta e disponível para todos os comandos subsequentes, sem necessidade de re-importar o módulo.

**Trade-off:** Se o processo PS crashar, todos os comandos pendentes falham. O runner implementa detecção de crash e restart automático, mas a sessão ADT é perdida.

### 11.2 Protocolo JSON

**Decisão:** Cada comando PS é wrappado em `try/catch` e retorna JSON via `ConvertTo-Json -Depth 10`.

```powershell
# Template de execução de cada comando
try {
    $result = <COMANDO_PSADT>
    @{ Success = $true; Data = $result; Error = $null } | ConvertTo-Json -Depth 10
} catch {
    @{ Success = $false; Data = $null; Error = @{
        Message = $_.Exception.Message
        Type = $_.Exception.GetType().FullName
        StackTrace = $_.ScriptStackTrace
    }} | ConvertTo-Json -Depth 10
}
```

### 11.3 Mapeamento 1:1

**Decisão:** Cada função PSADT = 1 método Go, sem abstrações adicionais.

**Motivo:** Mantém a lib previsível e documentável. Usuários familiarizados com PSADT encontram exatamente o mesmo conceito na API Go.

### 11.4 Pré-requisito: PSADT Instalado

**Decisão:** A lib **não distribui** o módulo PSADT. Requer `Install-Module PSAppDeployToolkit` prévio.

**Motivo:** Licenciamento (LGPL-3.0), versionamento independente, e simplificação da lib Go.

### 11.5 Windows-Only

**Decisão:** Build tag `//go:build windows` em todos os arquivos. Sem suporte Linux/macOS.

**Motivo:** PSADT é um framework Windows-only. Não há caso de uso cross-platform.

### 11.6 Serialização de Chamadas

**Decisão:** Mutex interno no runner. Uma chamada PS por vez.

**Motivo:** O processo PS é single-threaded. Chamadas concorrentes precisam ser serializadas.

**Futuro:** Pool de N runners para paralelismo (cada runner com sua própria sessão ADT).

---

## 12. Considerações Adicionais

### 12.1 Detecção de Elevação

Muitas funções PSADT requerem privilégios de administrador. A lib deve:

1. Verificar elevação no `NewClient()` via `windows.GetCurrentProcessToken()` e `IsElevated()`
2. Expor `Client.IsElevated() bool` para a aplicação consumidora
3. Fail-fast se `SessionConfig.RequireAdmin = true` e o processo não está elevado

### 12.2 Versionamento do Módulo PSADT

`NewClient` deve:

1. Verificar se o módulo está instalado: `Get-Module -ListAvailable PSAppDeployToolkit`
2. Verificar versão mínima compatível (default: `4.1.0`)
3. Retornar erro claro se módulo ausente ou versão incompatível

```go
client, err := psadt.NewClient(
    psadt.WithMinModuleVersion("4.1.0"), // Falha se < 4.1.0
)
```

### 12.3 Logging

A lib deve suportar um logger configurável:

```go
client, err := psadt.NewClient(
    psadt.WithLogger(slog.Default()), // slog, zap, zerolog, etc.
)
```

Cada comando PS executado é logado com nível Debug. Erros são logados com nível Error.

### 12.4 Timeout e Cancelamento

Suporte a `context.Context` para timeout e cancelamento:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := session.StartProcess(ctx, types.StartProcessOptions{
    FilePath: "setup.exe",
    ArgumentList: []string{"/S"},
})
```

### 12.5 Segurança

- **Sanitização de input:** O `cmdbuilder` deve escapar todos os parâmetros para evitar injection de comandos PS
- **SecureArgumentList:** Parâmetros sensíveis (senhas, chaves) não são logados
- **Sem credenciais em log:** O runner PS nunca loga credenciais ou tokens

### 12.6 Compatibilidade de Versão

| Versão PSADT | Suporte na lib Go |
|-------------|-------------------|
| 4.0.x | Não (API incompatível) |
| 4.1.0 - 4.1.8 | Sim (target principal) |
| 4.2.x | A validar (provável compatibilidade) |

### 12.7 Extensions Module

**Decisão:** A lib Go **não importa** o módulo `PSAppDeployToolkit.Extensions` automaticamente, mas permite ao usuário executá-lo se necessário.

**Motivo:** Extensions é um módulo opcional com funções customizadas. O core module auto-importa Extensions se presente, então comandos via runner PS já terão acesso. A lib Go pode invocar funções de Extensions via `Session.InvokeCommand(cmdString)` genérico.

### 12.8 ADMX / Group Policy

**Decisão:** A lib Go não gerencia ADMX templates diretamente. A precedência de configuração (GPO > config.psd1 > defaults) é transparente — `Get-ADTConfig` já retorna os valores resolvidos.

**Impacto:** Se o admin definiu `UI.FluentAccentColor` via GPO, `GetConfig()` retorna esse valor sem a lib precisar saber da fonte.

### 12.9 Deployment Phases (Pre/Main/Post)

O PSADT estrutura cada deployment em 3 fases sequenciais:

| DeploymentType | Pre-Phase | Main Phase | Post-Phase |
|---|---|---|---|
| **Install** | Welcome, fechar apps, uninstall anterior | Executar instalador (MSI/EXE) | Atalhos, registry, prompts |
| **Uninstall** | Mensagem ao usuário, fechar apps | Executar desinstalação | Cleanup, restart prompt |
| **Repair** | Mensagem ao usuário, fechar apps | Executar reparo | Registry, reset settings |

**Decisão:** A lib Go **não impõe** a estrutura de fases. O consumidor organiza as chamadas como quiser. Porém:

1. `SessionProperties.InstallPhase` reflete a fase atual (`Pre-Installation`, `Installation`, `Post-Installation`)
2. Exemplos em `examples/` seguem o padrão Pre/Main/Post como referência
3. A propriedade `DeploymentType` (Install/Uninstall/Repair) já está no `SessionConfig`

**Motivo:** Forçar fases no Go seria over-engineering — o PSADT já gerencia isso internamente quando `Close-ADTSession` é chamado.

### 12.10 Invoke-AppDeployToolkit.exe

O executável wrapper do PSADT suporta flags que influenciam a execução:

| Flag | Efeito | Mapeamento Go |
|------|--------|---------------|
| `/Debug` | Modo debug com logging verboso | `WithDebug(true)` (futuro) |
| `/32` | Força execução em PowerShell x86 | `WithPSPath("...\SysWOW64\...\powershell.exe")` |
| `/Core` | Usa PowerShell 7 (pwsh.exe) | `WithPowerShell7()` |
| `-File <path>` | Executa script customizado | N/A (lib Go é o "script") |

**Nota:** A lib Go substitui o `Invoke-AppDeployToolkit.exe` — ela mesma gerencia o processo PS. As flags `/32` e `/Core` são cobertas pelas opções `WithPSPath()` e `WithPowerShell7()` existentes.

---

## 13. Referência PSADT

### 13.1 Instalação

```powershell
# Via PowerShell Gallery
Install-Module -Name PSAppDeployToolkit -Scope AllUsers

# Verificar instalação
Get-Module -ListAvailable PSAppDeployToolkit
```

### 13.2 Estrutura do Módulo PSADT

Ref: https://psappdeploytoolkit.com/docs/4.1.x/reference/module-structure

#### Core Module (`PSAppDeployToolkit/`)

> **Não modificar** — se precisar customizar, usar a pasta do deployment template ou Extensions.

```
PSAppDeployToolkit/
├── PSAppDeployToolkit.psd1          # Module Manifest (versão, GUID, dependências)
├── PSAppDeployToolkit.psm1          # Root module (auto-loaded)
├── PSAppDeployToolkit.cer           # Certificado público (verificação de assinaturas)
├── Public/                          # ~135 funções exportadas (Verb-ADTNoun)
├── Private/                         # Funções internas do framework
├── ADMX/                            # Group Policy ADMX templates (config centralizada via GPO)
├── Assets/
│   └── AppIcon.png                  # Ícone padrão do toolkit (fallback se não customizado)
├── Config/
│   └── config.psd1                  # Configuração padrão (NÃO editar — usar config do deployment)
├── Strings/
│   └── strings.psd1                 # Strings UI padrão em inglês (fallback multi-idioma)
├── FrontEnd/                        # Templates de deployment scripts
├── Lib/                             # Assemblies C# compiladas
│   ├── PSADT.dll                    # Foundation (core logic)
│   ├── PSADT.UserInterface.dll      # WPF Fluent + WinForms Classic dialogs
│   ├── PSADT.ClientServer.Server.dll # Server-side (SYSTEM → pipes)
│   └── PSADT.ClientServer.Client.exe # Client-side (user session → UI)
```

#### Extensions Module (`PSAppDeployToolkit.Extensions/`)

> Auto-importado pelo core module. Permite funções customizadas no range de exit codes `70000-79999`.

```
PSAppDeployToolkit.Extensions/
├── PSAppDeployToolkit.Extensions.psd1  # Extensions manifest
└── PSAppDeployToolkit.Extensions.psm1  # Funções e lógica customizada
```

#### Deployment Template (o que o usuário prepara)

```
<DeploymentPackage>/
├── Invoke-AppDeployToolkit.ps1      # Script principal de deployment
├── Invoke-AppDeployToolkit.exe      # Wrapper executável (lança PS com elevação)
├── Assets/
│   ├── AppIcon.png                  # Logo customizado para Fluent UI (256x256 PNG)
│   ├── AppIcon-Dark.png             # Logo para dark mode (opcional, 256x256 PNG)
│   └── Banner.Classic.png           # Banner customizado para Classic UI (450x50 px PNG)
├── Config/
│   └── config.psd1                  # Overrides de configuração (UI, logging, MSI, toolkit)
├── Strings/
│   └── strings.psd1                 # Overrides de strings UI
├── Files/                           # Arquivos de instalação (MSI, EXE, etc.)
├── SupportFiles/                    # Arquivos de suporte auxiliares
├── PSAppDeployToolkit/              # Core module (cópia ou referência)
└── PSAppDeployToolkit.Extensions/   # Extensions module (opcional)
```

#### Precedência de Configuração

```
Group Policy (ADMX) > Deployment Config/config.psd1 > Module Config/config.psd1 > Built-in Defaults
```

### 13.3 Documentação

- Introdução: https://psappdeploytoolkit.com/docs/4.1.x/introduction
- Referência: https://psappdeploytoolkit.com/docs/4.1.x/reference
- Funções: https://psappdeploytoolkit.com/docs/4.1.x/category/functions
- Exit Codes: https://psappdeploytoolkit.com/docs/4.1.x/reference/exit-codes
- ADTSession Object: https://psappdeploytoolkit.com/docs/4.1.x/reference/adtsession-object
- Variables: https://psappdeploytoolkit.com/docs/4.1.x/reference/variables
- Language Strings: https://psappdeploytoolkit.com/docs/4.1.x/reference/language-strings
- Text Formatting: https://psappdeploytoolkit.com/docs/4.1.x/reference/text-formatting
- Config Settings: https://psappdeploytoolkit.com/docs/4.1.x/reference/config-settings
- Module Structure: https://psappdeploytoolkit.com/docs/4.1.x/reference/module-structure
- Deployment Structure: https://psappdeploytoolkit.com/docs/4.1.x/deployment-concepts/deployment-structure
- Invoke-AppDeployToolkit: https://psappdeploytoolkit.com/docs/4.1.x/deployment-concepts/invoke-appdeploytoolkit
- Release Notes: https://psappdeploytoolkit.com/docs/4.1.x/getting-started/release-notes
- GitHub: https://github.com/PSAppDeployToolkit/PSAppDeployToolkit
- PowerShell Gallery: https://www.powershellgallery.com/packages/PSAppDeployToolkit/4.1.8
