# Example: query

[English](#english) | [Português](#português)

---

## English

### What This Example Does

Demonstrates **system information queries** using `go-psadt`. Unlike the other examples, this one focuses on reading system state rather than modifying it. It shows:

1. Querying environment information **without a session** (`client.GetEnvironment()`)
2. Checking administrator privileges
3. Verifying network connectivity
4. Reading free disk space
5. Listing logged-on users with their session IDs
6. Detecting a pending system reboot
7. Checking whether a Windows service exists

### File

```
examples/query/main.go
```

### Key Concepts Demonstrated

| Concept | Code element | Description |
|---|---|---|
| Session-less query | `client.GetEnvironment()` | Retrieves ~90 PSADT environment variables without `Open-ADTSession` |
| Hierarchical env type | `env.OS.Name`, `env.OS.Version`, `env.PowerShell.PSVersion` | `EnvironmentInfo` is a structured Go struct — no map lookups |
| Admin check | `session.TestCallerIsAdmin()` | Returns `bool` — `true` if the process runs with elevated privileges |
| Network check | `session.TestNetworkConnection()` | Returns `bool` — `true` if a network connection is available |
| Disk space | `session.GetFreeDiskSpace()` | Returns free space on `%SystemDrive%` in **megabytes** as `uint64` |
| Logged-on users | `session.GetLoggedOnUser()` | Returns `[]types.LoggedOnUser` — NTAccount, SID, IsAdmin, SessionID |
| Pending reboot | `session.GetPendingReboot()` | Returns `*types.PendingRebootInfo` with per-source flags |
| Service exists | `session.TestServiceExists("Spooler")` | Returns `bool` — checks Windows Service Control Manager |
| Silent mode | `types.DeployModeSilent` | Session runs with no UI — suitable for query-only scenarios |

### Code Walkthrough

```go
// 1. Client without custom options — defaults: 5 min timeout, powershell.exe
client, err := psadt.NewClient()

// 2. GetEnvironment — works WITHOUT a session (direct PS variable access)
env, err := client.GetEnvironment()
fmt.Printf("OS: %s %s\n", env.OS.Name, env.OS.Version)
fmt.Printf("Architecture: %s\n", env.OS.Architecture)
fmt.Printf("PowerShell: %s\n", env.PowerShell.PSVersion)

// 3. Open a session for operations that require ADT context
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployInstall,
    DeployMode:     types.DeployModeSilent,  // no UI popups
    AppVendor:      "Query",
    AppName:        "System Check",
    AppVersion:     "1.0",
})

// 4. Admin check
isAdmin, _ := session.TestCallerIsAdmin()
fmt.Printf("Running as admin: %v\n", isAdmin)

// 5. Network check
hasNetwork, _ := session.TestNetworkConnection()
fmt.Printf("Network connected: %v\n", hasNetwork)

// 6. Disk space — returns MB as uint64
freeSpace, _ := session.GetFreeDiskSpace()
fmt.Printf("Free disk space: %d MB\n", freeSpace)

// 7. Logged-on users — slice of strongly-typed structs
users, _ := session.GetLoggedOnUser()
fmt.Printf("Logged-on users: %d\n", len(users))
for _, u := range users {
    fmt.Printf("  - %s (Session: %d)\n", u.NTAccount, u.SessionID)
}

// 8. Pending reboot — per-source breakdown
reboot, _ := session.GetPendingReboot()
fmt.Printf("Pending reboot: %v\n", reboot.IsSystemRebootPending)
// Also available: reboot.IsCBServicing, reboot.IsWindowsUpdate,
//                 reboot.IsSCCMClientReboot, reboot.IsFileRenameOps

// 9. Service existence check
svcExists, _ := session.TestServiceExists("Spooler")
fmt.Printf("Spooler service exists: %v\n", svcExists)
```

### `EnvironmentInfo` Structure (Partial)

`client.GetEnvironment()` returns a single nested struct covering the entire PSADT environment:

```
EnvironmentInfo
├── OS
│   ├── Name            string   "Microsoft Windows 11 Pro"
│   ├── Version         string   "10.0.22631"
│   ├── Architecture    string   "x64"
│   └── ...
├── PowerShell
│   ├── PSVersion       string   "5.1.22621.4391"
│   ├── PSVersionMajor  int      5
│   └── ...
├── Machine
│   ├── ComputerName    string
│   ├── Domain          string
│   └── ...
├── Users
│   ├── CurrentUserName string
│   ├── IsLocalAdmin    bool
│   └── ...
└── Paths
    ├── Windows         string   "C:\Windows"
    ├── System32        string   "C:\Windows\System32"
    └── ...
```

### PSADT Functions Used

| Go method | PSADT cmdlet |
|---|---|
| `client.GetEnvironment` | `Get-Variable` (PSADT env vars sweep) |
| `session.TestCallerIsAdmin` | `Test-ADTCallerIsAdmin` |
| `session.TestNetworkConnection` | `Test-ADTNetworkConnection` |
| `session.GetFreeDiskSpace` | `Get-ADTFreeDiskSpace` |
| `session.GetLoggedOnUser` | `Get-ADTLoggedOnUser` |
| `session.GetPendingReboot` | `Get-ADTPendingReboot` |
| `session.TestServiceExists` | `Test-ADTServiceExists` |

### Running

```powershell
# Must run as Administrator for accurate admin/service queries
cd examples\query
go run main.go
```

Expected output (values depend on the host machine):

```
OS: Microsoft Windows 11 Pro 10.0.22631
Architecture: x64
PowerShell: 5.1.22621.4391
Running as admin: true
Network connected: true
Free disk space: 48392 MB
Logged-on users: 1
  - CONTOSO\jdoe (Session: 2)
Pending reboot: false

System query completed!
```

---

## Português

### O que Este Exemplo Faz

Demonstra **consultas de informações do sistema** usando `go-psadt`. Diferente dos outros exemplos, este foca em ler o estado do sistema em vez de modificá-lo. Mostra:

1. Consultar informações de ambiente **sem sessão** (`client.GetEnvironment()`)
2. Verificar privilégios de administrador
3. Verificar conectividade de rede
4. Ler espaço livre em disco
5. Listar usuários conectados com seus IDs de sessão
6. Detectar reboot pendente do sistema
7. Verificar se um serviço Windows existe

### Arquivo

```
examples/query/main.go
```

### Conceitos-Chave Demonstrados

| Conceito | Elemento de código | Descrição |
|---|---|---|
| Consulta sem sessão | `client.GetEnvironment()` | Recupera ~90 variáveis de ambiente do PSADT sem `Open-ADTSession` |
| Tipo de env hierárquico | `env.OS.Name`, `env.OS.Version`, `env.PowerShell.PSVersion` | `EnvironmentInfo` é uma struct Go estruturada — sem lookups em map |
| Verificação de admin | `session.TestCallerIsAdmin()` | Retorna `bool` — `true` se o processo tem privilégios elevados |
| Verificação de rede | `session.TestNetworkConnection()` | Retorna `bool` — `true` se há conexão de rede disponível |
| Espaço em disco | `session.GetFreeDiskSpace()` | Retorna espaço livre em `%SystemDrive%` em **megabytes** como `uint64` |
| Usuários conectados | `session.GetLoggedOnUser()` | Retorna `[]types.LoggedOnUser` — NTAccount, SID, IsAdmin, SessionID |
| Reboot pendente | `session.GetPendingReboot()` | Retorna `*types.PendingRebootInfo` com flags por fonte |
| Existência de serviço | `session.TestServiceExists("Spooler")` | Retorna `bool` — consulta o Service Control Manager do Windows |
| Modo silencioso | `types.DeployModeSilent` | Sessão sem UI — adequada para cenários apenas de consulta |

### Walkthrough do Código

```go
// 1. Client sem opções customizadas — padrões: timeout 5 min, powershell.exe
client, err := psadt.NewClient()

// 2. GetEnvironment — funciona SEM sessão (acesso direto às variáveis PS)
env, err := client.GetEnvironment()
fmt.Printf("SO: %s %s\n", env.OS.Name, env.OS.Version)
fmt.Printf("Arquitetura: %s\n", env.OS.Architecture)
fmt.Printf("PowerShell: %s\n", env.PowerShell.PSVersion)

// 3. Abre sessão para operações que requerem contexto ADT
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployInstall,
    DeployMode:     types.DeployModeSilent,  // sem popups de UI
    AppVendor:      "Query",
    AppName:        "System Check",
    AppVersion:     "1.0",
})

// 4. Verificação de admin
isAdmin, _ := session.TestCallerIsAdmin()
fmt.Printf("Executando como admin: %v\n", isAdmin)

// 5. Verificação de rede
hasNetwork, _ := session.TestNetworkConnection()
fmt.Printf("Rede conectada: %v\n", hasNetwork)

// 6. Espaço em disco — retorna MB como uint64
freeSpace, _ := session.GetFreeDiskSpace()
fmt.Printf("Espaço livre: %d MB\n", freeSpace)

// 7. Usuários conectados — slice de structs fortemente tipadas
users, _ := session.GetLoggedOnUser()
fmt.Printf("Usuários conectados: %d\n", len(users))
for _, u := range users {
    fmt.Printf("  - %s (Sessão: %d)\n", u.NTAccount, u.SessionID)
}

// 8. Reboot pendente — breakdown por fonte
reboot, _ := session.GetPendingReboot()
fmt.Printf("Reboot pendente: %v\n", reboot.IsSystemRebootPending)
// Também disponível: reboot.IsCBServicing, reboot.IsWindowsUpdate,
//                    reboot.IsSCCMClientReboot, reboot.IsFileRenameOps

// 9. Verificação de existência de serviço
svcExists, _ := session.TestServiceExists("Spooler")
fmt.Printf("Serviço Spooler existe: %v\n", svcExists)
```

### Estrutura de `EnvironmentInfo` (Parcial)

`client.GetEnvironment()` retorna uma única struct aninhada cobrindo todo o ambiente PSADT:

```
EnvironmentInfo
├── OS
│   ├── Name            string   "Microsoft Windows 11 Pro"
│   ├── Version         string   "10.0.22631"
│   ├── Architecture    string   "x64"
│   └── ...
├── PowerShell
│   ├── PSVersion       string   "5.1.22621.4391"
│   ├── PSVersionMajor  int      5
│   └── ...
├── Machine
│   ├── ComputerName    string
│   ├── Domain          string
│   └── ...
├── Users
│   ├── CurrentUserName string
│   ├── IsLocalAdmin    bool
│   └── ...
└── Paths
    ├── Windows         string   "C:\Windows"
    ├── System32        string   "C:\Windows\System32"
    └── ...
```

### Funções PSADT Utilizadas

| Método Go | Cmdlet PSADT |
|---|---|
| `client.GetEnvironment` | `Get-Variable` (varredura de variáveis env do PSADT) |
| `session.TestCallerIsAdmin` | `Test-ADTCallerIsAdmin` |
| `session.TestNetworkConnection` | `Test-ADTNetworkConnection` |
| `session.GetFreeDiskSpace` | `Get-ADTFreeDiskSpace` |
| `session.GetLoggedOnUser` | `Get-ADTLoggedOnUser` |
| `session.GetPendingReboot` | `Get-ADTPendingReboot` |
| `session.TestServiceExists` | `Test-ADTServiceExists` |

### Executando

```powershell
# Deve executar como Administrador para consultas precisas de admin/serviço
cd examples\query
go run main.go
```

Saída esperada (valores dependem da máquina host):

```
SO: Microsoft Windows 11 Pro 10.0.22631
Arquitetura: x64
PowerShell: 5.1.22621.4391
Executando como admin: true
Rede conectada: true
Espaço livre: 48392 MB
Usuários conectados: 1
  - CONTOSO\jdoe (Sessão: 2)
Reboot pendente: false

System query completed!
```
