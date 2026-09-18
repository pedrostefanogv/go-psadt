# Integração RMM — go-psadt

> Guia para agents RMM (ex.: Discovery) usando go-psadt para instalar/atualizar
> aplicações com notificação ao usuário, contagem regressiva, adiamento (defer)
> e popups informativos.

## Pré-requisitos

- Windows 10/11 ou Server 2016+
- PowerShell 5.1 (built-in) ou 7+ — use `psadt.WithAutoPreferPS7()`
- PSAppDeployToolkit >= 4.1.0 instalado na máquina:

```powershell
Install-Module -Name PSAppDeployToolkit -Scope AllUsers
```

- O agent normalmente roda como **SYSTEM** (necessário para instalações
  machine-wide; os diálogos do PSADT aparecem na sessão do usuário ativo).

## Cenário 1 — "O Firefox será atualizado em 30 segundos, salve seu trabalho"

Notificação com contagem regressiva, pedido para salvar trabalho e adiamento:

```go
client, _ := psadt.NewClient(
    psadt.WithAutoPreferPS7(),
    psadt.WithAutoReconnect(),            // sobrevive a agents de longa duração
    psadt.WithTimeout(15 * time.Minute),
)
defer client.Close()

session, _ := client.OpenSession(types.NewSessionConfig().
    App("Mozilla", "Firefox", "128.0").
    Install().
    Interactive().
    CloseProcesses(types.ProcessDefinition{Name: "firefox", Description: "Mozilla Firefox"}).
    Build())

err := session.NotifyUpdate(psadt.UpdateNotification{
    AppName:          "Firefox",
    Processes:        []types.ProcessDefinition{{Name: "firefox", Description: "Mozilla Firefox"}},
    CountdownSeconds: 30,            // "...será atualizado em 30 segundos"
    PromptToSave:     boolPtr(true), // "salve seu trabalho"
    AllowDefer:       true,          // usuário pode adiar
    DeferTimes:       3,             // até 3 adiamentos
    BlockExecution:   true,          // impede reabrir o Firefox durante o update
})

// NotifyUpdate retorna quando o usuário aceita, adia ou o countdown expira.
// Prossiga com a instalação:
result, err := session.StartMsiProcess(types.MsiProcessOptions{
    Action:   types.MsiInstall,
    FilePath: "firefox.msi",
    PassThru: true,
})
```

Variações de controle via `ShowInstallationWelcome` + `types.WelcomeOptions`:

| Opção | Efeito |
|---|---|
| `CloseProcessesCountdown` | Countdown visível antes de fechar os processos |
| `ForceCloseProcessesCountdown` | Countdown **incondicional**, ignora defer |
| `ForceCountdown` | Countdown de continuar/adiar conforme processos abertos |
| `DeferDays` / `DeferDeadline` | Limite de adiamento por dias ou data absoluta |
| `DeferRunInterval` | Intervalo mínimo entre prompts de adiamento |
| `CheckDiskSpace` / `RequiredDiskSpace` | Validação de espaço em disco |

## Cenário 2 — Notificação silenciosa (só progresso)

```go
session.NotifyProgress("Firefox", "Atualizando o Firefox para 128.0...")
result, _ := session.StartMsiProcess(/* ... */)
session.CloseInstallationProgress()
```

## Cenário 3 — Informar sem bloquear (toast/balão)

```go
session.NotifyInfo("Atualização concluída", "O Firefox foi atualizado para 128.0.")
```

## Cenário 4 — Perguntar ao usuário (adiar ou não)

```go
resp, _ := session.ShowInstallationPrompt(types.PromptOptions{
    Title:           "Atualização disponível",
    Message:         "Deseja atualizar o Firefox agora?",
    ButtonLeftText:  "Atualizar agora",
    ButtonRightText: "Adiar",
    Icon:            types.IconQuestion,
})
if resp.ButtonClicked == "Right" { /* agendar para depois */ }
```

## Cenário 5 — Reinício com contagem regressiva

```go
session.ShowInstallationRestartPrompt(types.RestartPromptOptions{
    CountdownSeconds:       300,
    CountdownNoHideSeconds: 60,
})
```

## Cancelamento de instalação travada

```go
// Ao receber cancelamento do servidor RMM:
if err := client.Abort(); err != nil { /* log */ }
// Mata a árvore inteira: PowerShell + instalador + filhos (taskkill /T /F).
```

## Observabilidade para o painel RMM

```go
client, _ := psadt.NewClient(
    psadt.OnCommand(func(cmd string, d time.Duration, err error) {
        metrics.Record("psadt_command", d, err) // latência/erros por comando
    }),
)
// client.CommandCount(), client.LastError(), client.Uptime()

liveCh := session.LiveOutput() // stream do log PSADT em tempo real
```

## Deploy paralelo

Cada `Client` = 1 processo PowerShell. Para N deploys simultâneos use o pool:

```go
pool, _ := psadt.NewClientPool(ctx, 4, psadt.WithAutoReconnect())
defer pool.Close()
client, _ := pool.Acquire(ctx)
defer pool.Release(client)
```

## Erros tipados (classificação no servidor)

```go
switch {
case psadt.IsRebootRequired(err):  // exit 3010/1641 → agendar reboot
case psadt.IsUserCancelled(err):   // exit 1602 → usuário cancelou
case psadt.IsAccessDenied(err):
case psadt.IsTimeout(err):
case psadt.IsFileNotFound(err):
case psadt.IsNetworkError(err):
}
```

## Deferimento persistido (auditoria)

```go
hist, _ := session.GetDeferHistory() // DeferTimesRemaining, DeferDeadline
_ = session.ResetDeferHistory()
```

## Notas de produção

1. **Sessão de usuário**: os diálogos do PSADT precisam da sessão interativa;
   rodando como SYSTEM o PSADT v4.1 cria as UIs na sessão do usuário ativo.
2. **`WithAutoReconnect()`** re-importa o módulo automaticamente se o PS morrer.
3. **Timeouts por diálogo**: `ShowInstallationWelcome` bloqueia até o usuário
   decidir ou o countdown acabar — use `session.WithContext(ctx)` com o
   deadline da tarefa RMM.
4. **Winget como SYSTEM** é historicamente frágil em versões antigas do
   Windows — prefira instaladores diretos (MSI/EXE) quando possível.
