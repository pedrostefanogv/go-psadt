# Example: gui (UI Lab)

[English](#english) | [Português](#português)

---

## English

### Overview

This example provides a **graphical test lab** for PSADT UI features through a local web interface.

It lets you test:

- Modal dialog options (`ShowDialogBox`): icons, button types, timeout.
- PSADT modal prompt (`ShowInstallationPrompt`): alignment, custom button labels, icon, timeout.
- Alerts (`ShowBalloonTip`): title, message, icon, display time.
- Progress modal (`ShowInstallationProgress`) with automatic close.
- Welcome modal (`ShowInstallationWelcome`) with process list and countdown.
- Toolkit installation checks and installation via **PowerShell Gallery** (`CurrentUser` or `AllUsers`).
- Visual config snippet generation for Fluent/Classic style and assets (icon/banner).
- Runtime diagnostics via `/diag`, including client reconnect count and last invalidation reason.

### Run

```powershell
cd examples\gui
go run main.go
```

Open:

```text
http://127.0.0.1:17841
```

### Notes

- The panel has its own light/dark mode toggle for easier testing.
- Fluent/Classic, accent color, and asset files are controlled by PSADT configuration. The app generates a ready-to-use snippet for `config.psd1`.
- `Show-ADTInstallationProgress` opens a separate WPF window, so it can appear in the Windows taskbar as its own window.
- `/diag` exposes `client_reconnect_count`, `last_client_invalidation_reason`, and `last_client_invalidation_at` to make reconnect behavior observable.
- For `AllUsers` installation, open terminal as Administrator.

---

## Português

### Visão geral

Este exemplo entrega um **laboratório gráfico de testes** para as funcionalidades de UI do PSADT via interface web local.

Ele permite testar:

- Opções de modal (`ShowDialogBox`): ícones, tipos de botões, timeout.
- Prompt modal do PSADT (`ShowInstallationPrompt`): alinhamento, textos de botões customizados, ícone, timeout.
- Alertas (`ShowBalloonTip`): título, mensagem, ícone e tempo.
- Modal de progresso (`ShowInstallationProgress`) com fechamento automático.
- Modal de boas-vindas (`ShowInstallationWelcome`) com lista de processos e contagem regressiva.
- Verificação e instalação do toolkit via **PowerShell Gallery** (`CurrentUser` ou `AllUsers`).
- Geração de snippet de configuração visual para Fluent/Classic e assets (ícone/banner).
- Diagnóstico de runtime via `/diag`, incluindo contagem de reconexões do client e motivo da última invalidação.

### Executar

```powershell
cd examples\gui
go run main.go
```

Abra:

```text
http://127.0.0.1:17841
```

### Observações

- O painel possui alternância claro/escuro para facilitar os testes.
- Fluent/Classic, cor de destaque e arquivos de assets dependem da configuração ativa do PSADT. O app gera snippet pronto para `config.psd1`.
- `Show-ADTInstallationProgress` abre uma janela WPF separada, então pode aparecer como janela própria na barra de tarefas do Windows.
- O endpoint `/diag` expõe `client_reconnect_count`, `last_client_invalidation_reason` e `last_client_invalidation_at` para tornar o comportamento de reconexão observável.
- Para instalação `AllUsers`, execute o terminal como Administrador.
