# Example: uninstall

[English](#english) | [Português](#português)

---

## English

### What This Example Does

Demonstrates a complete **application uninstallation workflow** using `go-psadt`. It shows how to:

1. Start a PSADT client configured for uninstallation
2. Display a welcome dialog that requests the user to close running processes
3. Search for an installed application by name
4. Uninstall it using PSADT's `Uninstall-ADTApplication`
5. Clean up leftover registry keys

### File

```
examples/uninstall/main.go
```

### Key Concepts Demonstrated

| Concept | Code element | Description |
|---|---|---|
| Uninstall session | `types.DeployUninstall` | Sets `DeploymentType` so PSADT uses uninstall-specific behavior and logging |
| Welcome on uninstall | `session.ShowInstallationWelcome(types.WelcomeOptions{...})` | Prompts to close `widget` and `widgethelper` before removal |
| Application search | `session.GetApplication(types.GetApplicationOptions{Name: [...]})` | Queries the registry for installed applications matching a name pattern |
| Typed result | `len(apps)` | Returns `[]types.InstalledApplication` — strongly typed, no string parsing |
| Uninstall | `session.UninstallApplication(types.UninstallApplicationOptions{...})` | Invokes the application's uninstaller through PSADT |
| Registry cleanup | `session.RemoveRegistryKey(types.RemoveRegistryKeyOptions{...})` | Deletes the entire `HKLM\SOFTWARE\Contoso\WidgetPro` key tree |

### Code Walkthrough

```go
// 1. Client with 5-minute timeout — uninstalls are usually faster
client, err := psadt.NewClient(
    psadt.WithTimeout(5 * time.Minute),
)

// 2. Session for uninstallation
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployUninstall,
    DeployMode:     types.DeployModeInteractive,
    AppVendor:      "Contoso",
    AppName:        "Widget Pro",
    AppVersion:     "2.0.0",
})

// 3. Welcome dialog — close running processes before removal
session.ShowInstallationWelcome(types.WelcomeOptions{
    CloseProcesses: []types.ProcessDefinition{
        {Name: "widget"},
        {Name: "widgethelper"},
    },
})

// 4. Search — verifies the app is actually installed
apps, err := session.GetApplication(types.GetApplicationOptions{
    Name: []string{"Widget Pro"},
})
fmt.Printf("Found %d matching application(s)\n", len(apps))

// 5. Uninstall — PSADT locates and runs the correct uninstaller
session.UninstallApplication(types.UninstallApplicationOptions{
    Name: []string{"Widget Pro"},
})

// 6. Registry cleanup
session.RemoveRegistryKey(types.RemoveRegistryKeyOptions{
    Key: `HKLM\SOFTWARE\Contoso\WidgetPro`,
})
```

### Comparison with `install` Example

| Aspect | `install` | `uninstall` |
|---|---|---|
| `DeploymentType` | `DeployInstall` | `DeployUninstall` |
| Timeout | 10 minutes | 5 minutes |
| Process step | `StartMsiProcess` | `UninstallApplication` |
| Registry step | `SetRegistryKey` (create) | `RemoveRegistryKey` (delete) |
| Progress bar | Yes | No |

### PSADT Functions Used

| Go method | PSADT cmdlet |
|---|---|
| `client.OpenSession` | `Open-ADTSession` |
| `session.ShowInstallationWelcome` | `Show-ADTInstallationWelcome` |
| `session.GetApplication` | `Get-ADTApplication` |
| `session.UninstallApplication` | `Uninstall-ADTApplication` |
| `session.RemoveRegistryKey` | `Remove-ADTRegistryKey` |
| `session.Close` | `Close-ADTSession` |

### Running

```powershell
# Must run as Administrator
cd examples\uninstall
go run main.go
```

---

## Português

### O que Este Exemplo Faz

Demonstra um **fluxo completo de desinstalação de aplicação** usando `go-psadt`. Mostra como:

1. Iniciar um cliente PSADT configurado para desinstalação
2. Exibir um diálogo de boas-vindas solicitando fechar processos em execução
3. Buscar uma aplicação instalada pelo nome
4. Desinstalá-la usando o `Uninstall-ADTApplication` do PSADT
5. Limpar chaves de registro remanescentes

### Arquivo

```
examples/uninstall/main.go
```

### Conceitos-Chave Demonstrados

| Conceito | Elemento de código | Descrição |
|---|---|---|
| Sessão de desinstalação | `types.DeployUninstall` | Define `DeploymentType` para comportamento e logging específicos de desinstalação |
| Welcome na desinstalação | `session.ShowInstallationWelcome(types.WelcomeOptions{...})` | Solicita fechar `widget` e `widgethelper` antes da remoção |
| Busca de aplicação | `session.GetApplication(types.GetApplicationOptions{Name: [...]})` | Consulta o registro por aplicações instaladas correspondendo a um padrão de nome |
| Resultado tipado | `len(apps)` | Retorna `[]types.InstalledApplication` — fortemente tipado, sem parsing de strings |
| Desinstalação | `session.UninstallApplication(types.UninstallApplicationOptions{...})` | Invoca o desinstalador da aplicação via PSADT |
| Limpeza do registro | `session.RemoveRegistryKey(types.RemoveRegistryKeyOptions{...})` | Remove toda a árvore de chaves `HKLM\SOFTWARE\Contoso\WidgetPro` |

### Walkthrough do Código

```go
// 1. Client com timeout de 5 minutos — desinstalações costumam ser mais rápidas
client, err := psadt.NewClient(
    psadt.WithTimeout(5 * time.Minute),
)

// 2. Sessão para desinstalação
session, err := client.OpenSession(types.SessionConfig{
    DeploymentType: types.DeployUninstall,
    DeployMode:     types.DeployModeInteractive,
    AppVendor:      "Contoso",
    AppName:        "Widget Pro",
    AppVersion:     "2.0.0",
})

// 3. Diálogo de boas-vindas — fecha processos antes da remoção
session.ShowInstallationWelcome(types.WelcomeOptions{
    CloseProcesses: []types.ProcessDefinition{
        {Name: "widget"},
        {Name: "widgethelper"},
    },
})

// 4. Busca — verifica se o app está realmente instalado
apps, err := session.GetApplication(types.GetApplicationOptions{
    Name: []string{"Widget Pro"},
})
fmt.Printf("Encontradas %d aplicação(ões) correspondentes\n", len(apps))

// 5. Desinstala — PSADT localiza e executa o desinstalador correto
session.UninstallApplication(types.UninstallApplicationOptions{
    Name: []string{"Widget Pro"},
})

// 6. Limpeza do registro
session.RemoveRegistryKey(types.RemoveRegistryKeyOptions{
    Key: `HKLM\SOFTWARE\Contoso\WidgetPro`,
})
```

### Comparação com o Exemplo `install`

| Aspecto | `install` | `uninstall` |
|---|---|---|
| `DeploymentType` | `DeployInstall` | `DeployUninstall` |
| Timeout | 10 minutos | 5 minutos |
| Etapa principal | `StartMsiProcess` | `UninstallApplication` |
| Etapa de registro | `SetRegistryKey` (criar) | `RemoveRegistryKey` (remover) |
| Barra de progresso | Sim | Não |

### Funções PSADT Utilizadas

| Método Go | Cmdlet PSADT |
|---|---|
| `client.OpenSession` | `Open-ADTSession` |
| `session.ShowInstallationWelcome` | `Show-ADTInstallationWelcome` |
| `session.GetApplication` | `Get-ADTApplication` |
| `session.UninstallApplication` | `Uninstall-ADTApplication` |
| `session.RemoveRegistryKey` | `Remove-ADTRegistryKey` |
| `session.Close` | `Close-ADTSession` |

### Executando

```powershell
# Deve executar como Administrador
cd examples\uninstall
go run main.go
```
