# Plano de Melhoria e Otimização — go-psadt

> Análise completa do projeto `go-psadt` (dePara Go ↔ PSAppDeployToolkit v4.1.x)
> Data: 2026-07-06 · Status: **Fase 1 concluída ✅ | Fases 2-4 pendentes de aprovação**

---

## 1. Sumário Executivo

O projeto `go-psadt` é um bridge Go ↔ PowerShell que mantém um processo `powershell.exe` persistente e expõe ~127 métodos tipados cobrindo ~90% das 141 funções públicas do PSAppDeployToolkit v4.1.x. A arquitetura em três camadas (public API → internal cmdbuilder/runner/parser → runtime OS) é sólida e bem documentada.

**No entanto, a análise revelou:**

- **3 bugs críticos** ✅ Corrigidos (recursão infinita, goroutine leak, race condition no canal)
- **5 bugs de correção** ✅ Corrigidos (Reconnect perde config, envCacheTTL não configurável, parêntese extra falso positivo, correção dos registry typed getters, proteção do closed)
- **12 funções PSADT públicas sem wrapper Go** (internas/deprecated — baixa prioridade)
- **8 melhorias arquiteturais** (testes de integração, CI, benchmarks, pprof, etc.)
- **6 melhorias de API** (erros tipados, options fluentes, validação em compile-time)

O plano está organizado em **4 fases** com prioridade decrescente.

---

## 2. Bugs Críticos (Corrigir imediatamente)

### BC-1 — Recursão infinita em `Session.getContext()`

**Arquivo:** `session.go` linhas 168-173

```go
func (s *Session) getContext() (context.Context, context.CancelFunc) {
	if s.ctx != nil {
		return context.WithCancel(s.ctx)
	}
	return s.getContext()  // ← BUG: chama a si mesma quando s.ctx == nil
}
```

Quando `s.ctx == nil` (o caso padrão — sessão sem `WithContext`), a função entra em recursão infinita e causa stack overflow. **Todos os métodos de sessão que usam `getContext()` travam imediatamente** a menos que o caller use `WithContext`.

**Correção:**
```go
return s.client.defaultContext()
```

**Impacto:** Sem essa correção, o Quick Start do README não funciona — `session.ShowInstallationWelcome(...)` sem `WithContext` entra em stack overflow.

---

### BC-2 — Goroutine leak em `readResponse()`

**Arquivo:** `internal/runner/command.go` linhas 60-110

A cada iteração do loop `for`, uma **nova goroutine** é criada para ler do scanner:

```go
for {
	go func() {
		if scanner.Scan() {
			lineCh <- scanResult{line: scanner.Text()}
		} else {
			// ...
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()   // ← goroutine acima fica órfã, bloqueada no scanner
	case <-time.After(timeout):
		return nil, fmt.Errorf(...)  // ← mesma coisa
	}
}
```

Em caso de timeout ou cancelamento, a goroutine anterior permanece bloqueada em `scanner.Scan()`. Após N timeouts, há N goroutines acumuladas. Como o `bufio.Scanner` compartilhado não é thread-safe, a próxima iteração cria outra goroutine que compete pelo mesmo scanner — comportamento indefinido.

**Correção:** Criar a goroutine de leitura **uma única vez** fora do loop, ou usar um canal persistente com um único reader:

```go
func (r *Runner) readResponse(ctx context.Context) ([]byte, error) {
	// ... setup ...
	lineCh := make(chan scanResult, 1)
	go func() {
		for scanner.Scan() {
			lineCh <- scanResult{line: scanner.Text()}
		}
		// scanner parou — envia EOF ou erro
		if err := scanner.Err(); err != nil {
			lineCh <- scanResult{err: err}
		} else {
			lineCh <- scanResult{eof: true}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(timeout):
			return nil, fmt.Errorf("timeout...")
		case result := <-lineCh:
			// processar linha
		}
	}
}
```

---

### BC-3 — Race condition no `liveOutputCh` durante `Stop()`

**Arquivo:** `internal/runner/runner.go` linhas 130-155

`Stop()` fecha `liveOutputCh` sob o mutex, mas `drainStderr()` (goroutine de fundo) e `emitOutput()` (chamado de `readResponse`) escrevem no canal **sem** segurar o mutex:

```go
func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// ...
	close(r.liveOutputCh)  // ← sem proteção contra escritas concorrentes
}

func (r *Runner) emitOutput(line string) {
	select {
	case r.liveOutputCh <- line:  // ← pode panicar se canal foi fechado
	default:
	}
}
```

Se `Stop()` fecha o canal enquanto `drainStderr` ou `readResponse` está escrevendo, ocorre `panic: send on closed channel`.

**Correção:** Usar um flag atômico `closing` ou um `sync.Once` para garantir que escritas não ocorram após o close, ou nunca fechar o canal e deixar o GC coletar.

---

## 3. Bugs de Correção (Alta prioridade)

### BC-4 — `Reconnect()` perde configuração original

**Arquivo:** `psadt.go` linhas 113-135

```go
func (c *Client) Reconnect(ctx context.Context) error {
	// ...
	cfg := runner.Config{
		Timeout: c.timeout,
	}
	r, err := runner.New(cfg)  // ← perde PSPath e UsePowerShell7
```

Se o cliente foi criado com `WithPowerShell7()` ou `WithPSPath(...)`, o `Reconnect` ignora essas opções e volta ao auto-detect default.

**Correção:** Armazenar `psPath` e `usePowerShell7` no `Client` e reutilizá-los.

---

### BC-5 — `WithEnvCacheTTL` documentado mas não implementado

**Arquivo:** `README.md` lista `WithEnvCacheTTL(5 * time.Minute)` como option, mas **não existe** em `psadt.go`. O `Client` usa `defaultEnvCacheTTL` fixo de 5 minutos.

**Correção:** Adicionar a option e aplicar `cfg.envCacheTTL` no construtor.

---

### BC-6 — Parêntese extra em `formatMapParam`

**Arquivo:** `internal/cmdbuilder/builder.go` linha ~190

```go
pairs = append(pairs, fmt.Sprintf("%s=%s", EscapeString(...), EscapeString(...)))  // ← ) extra
```

Isso é um erro de sintaxe que **deveria** falhar a compilação. Se compila, pode ser um bug silencioso de parsing. Verificar com `go vet` (passou, então pode ser que o arquivo esteja diferente do indexado).

---

### BC-7 — `GetRegistryKeyString/DWord/MultiString/Binary/QWord` usam `.Name` incorretamente

**Arquivo:** `registry.go` e `batch.go`

```go
cmd := fmt.Sprintf("(Get-ADTRegistryKey -Key %s -Name %s).%s",
	escapeArg(key), escapeArg(name), escapeArg(name))
```

`Get-ADTRegistryKey` retorna um objeto `RegistryValue` cuja propriedade com o valor é tipicamente `.Value`, não `.<name>`. O `.Name` acessa o nome do valor, não o valor em si. Isso retorna o próprio nome em vez do valor armazenado.

**Correção:** Usar `.Value` ou validar contra a documentação do PSADT v4.1.

---

### BC-8 — `Session.closed` sem proteção contra acesso concorrente

**Arquivo:** `session.go`

`closed` é lido/escrito em `CloseWithContext` sem mutex. Se duas goroutines chamam `Close` simultaneamente, ambas podem passar o check `if s.closed` antes de setar `s.closed = true`.

**Correção:** Usar `sync.Once` ou `atomic.Bool`.

---

## 4. Lacunas de Cobertura PSADT

### Funções públicas PSADT v4.1.x **sem** wrapper Go:

| Função PSADT | Categoria | Prioridade |
|---|---|---|
| `Complete-ADTFunction` | Lifecycle | Baixa (interno) |
| `Convert-ADTValuesFromRemainingArguments` | Interno | Baixa |
| `Convert-ADTValueType` | Interno | Baixa |
| `Get-ADTBoundParametersAndDefaultValues` | Interno | Baixa |
| `Get-ADTCommandTable` | Interno | Baixa |
| `Get-ADTObjectProperty` | Deprecated | Não envolver |
| `Initialize-ADTFunction` | Lifecycle | Baixa |
| `Initialize-ADTModule` | Lifecycle | Média |
| `Initialize-ADTModuleIfUninitialized` | Lifecycle | Média |
| `Invoke-ADTFunctionErrorHandler` | Interno | Baixa |
| `Invoke-ADTObjectMethod` | Deprecated | Não envolver |
| `New-ADTErrorRecord` | Interno | Baixa |
| `New-ADTValidateScriptErrorRecord` | Interno | Baixa |
| `Remove-ADTHashtableNullOrEmptyValues` | Utility | Baixa |
| `Resolve-ADTErrorRecord` | Interno | Baixa |
| `Select-ADTUniqueObject` | Utility | Baixa |

**Recomendação:** As funções marcadas "Interno" são usadas internamente pelo PSADT e não fazem sentido como API pública Go. As de "Lifecycle" (`Initialize-ADTModule`, `Initialize-ADTModuleIfUninitialized`) podem ser úteis para controle fino. As "Deprecated" não devem ser envolvidas.

---

## 5. Melhorias Arquiteturais

### MA-1 — Testes de integração com PowerShell real

Atualmente os testes são unitários e não validam o protocolo completo stdin/stdout/marcadores. Criar:

- `internal/runner/runner_integration_test.go` — testa `Execute` com comandos PS reais (`$true`, `Get-Date`, etc.)
- `psadt_integration_test.go` — testa `NewClient` → `OpenSession` → `Close` com PSADT instalado
- Usar `testing.Short()` para pular em CI sem Windows/PSADT

### MA-2 — Pipeline de CI (GitHub Actions)

Criar `.github/workflows/ci.yml`:
- `go vet ./...`
- `go test ./...` (apenas Windows runner)
- `golangci-lint run`
- Coverage report

### MA-3 — Benchmarks

Criar `internal/runner/bench_test.go`:
- Benchmark da latência de round-trip (`Execute("$true")`)
- Benchmark do `cmdbuilder.Build` com structs grandes
- Comparar overhead vs `powershell.exe -Command` por chamada

### MA-4 — Suporte a PowerShell 7 como default opcional

`detectPowerShell` prefere `powershell.exe` (5.1). PSADT v4.1 funciona melhor com PS 7. Adicionar `WithAutoPreferPS7()` que tenta `pwsh.exe` primeiro se disponível.

### MA-5 — Health check automático antes de comandos

Adicionar option `WithAutoReconnect()` que, se `!r.IsAlive()` antes de `executeWrapped`, chama `Reconnect` automaticamente. Útil para RMM agents de longa duração.

### MA-6 — Métricas de observabilidade

Adicionar ao `Client`:
- `CommandCount()` — total de comandos executados
- `LastError()` — último erro não-nil
- `Uptime()` — tempo desde `NewClient`
- Hook `OnCommand(func(cmd string, duration time.Duration, err error))`

### MA-7 — Suporte a `context.Context` explícito em todos os métodos

Embora `WithContext` evite duplicação, muitos usuários Go esperam `ctx` como primeiro parâmetro. Adicionar variants `*WithContext` para os ~100 métodos (pode ser gerado por codegen).

### MA-8 — Documentação gerada por exemplo

Adicionar `go doc -all` output como arquivo `API.md` no CI, garantindo que a doc está sempre atualizada.

---

## 6. Melhorias de API

### API-1 — Erros tipados por categoria

Além de `IsRebootRequired`/`IsUserCancelled`, adicionar:

```go
parser.IsAccessDenied(err)      // ExitCode 5
parser.IsNetworkError(err)      // tipo System.Net.WebException
parser.IsFileNotFound(err)      // tipo System.IO.FileNotFoundException
parser.IsTimeout(err)           // wrapper de context.DeadlineExceeded
```

### API-2 — Validação compile-time de enums

Usar tipos genéricos (Go 1.18+) ou `go:generate` para validar que valores de `DeploymentType`, `DeployMode`, etc. são válidos em compile-time.

### API-3 — `SessionConfig` builder fluente

```go
session, _ := client.OpenSession(
    psadt.NewSessionConfig().
        App("Contoso", "Widget Pro", "2.0.0").
        Install().
        Interactive(),
)
```

### API-4 — Suporte a `io.Reader` para scripts grandes

`ExecuteRawScript` aceita `string`. Para scripts muito grandes, adicionar `ExecuteRawScriptReader(ctx, io.Reader)` para streaming.

### API-5 — Pool de clientes para deploy paralelo

Como cada `Client` tem seu próprio processo PS, deploy paralelo de N apps precisa de N clients. Criar `ClientPool` com N runners pré-aquecidos.

### API-6 — Hooks de lifecycle de sessão

```go
session.OnClose(func(exitCode int) {
    log.Printf("Session closed with exit %d", exitCode)
})
```

---

## 7. Plano de Implementação por Fases

### Fase 1 — Correções Críticas (1-2 dias)

| ID | Item | Arquivo | Esforço |
|---|---|---|---|
| BC-1 | Corrigir recursão infinita em `getContext()` | `session.go` | 5 min |
| BC-2 | Corrigir goroutine leak em `readResponse()` | `internal/runner/command.go` | 1h |
| BC-3 | Corrigir race no `liveOutputCh` | `internal/runner/runner.go` | 1h |
| BC-4 | `Reconnect()` preserva config original | `psadt.go` | 30 min |
| BC-5 | Implementar `WithEnvCacheTTL` | `psadt.go` | 15 min |
| BC-7 | Corrigir `GetRegistryKey*` (`.Value` em vez de `.Name`) | `registry.go`, `batch.go` | 30 min |
| BC-8 | Proteger `Session.closed` com `sync.Once` | `session.go` | 15 min |

**Validação:** `go vet`, `go test ./...`, testar Quick Start do README.

---

### Fase 2 — Robustez e Testes (3-5 dias)

| ID | Item | Esforço |
|---|---|---|
| MA-1 | Testes de integração com PS real | 1 dia |
| MA-2 | Pipeline CI GitHub Actions | 2h |
| MA-3 | Benchmarks de latência | 2h |
| BC-6 | Validar/corrigir `formatMapParam` | 30 min |
| — | Adicionar testes para `getContext`, `Reconnect`, `envCacheTTL` | 3h |
| — | Fuzzing do `cmdbuilder.Build` e `EscapeString` | 2h |

---

### Fase 3 — Melhorias Arquiteturais (1 semana)

| ID | Item | Esforço |
|---|---|---|
| MA-4 | `WithAutoPreferPS7()` | 1h |
| MA-5 | `WithAutoReconnect()` | 3h |
| MA-6 | Métricas de observabilidade | 4h |
| MA-8 | `API.md` gerada no CI | 1h |
| API-1 | Erros tipados por categoria | 3h |
| API-5 | `ClientPool` para deploy paralelo | 1 dia |

---

### Fase 4 — API e DX (1-2 semanas)

| ID | Item | Esforço |
|---|---|---|
| MA-7 | Variants `*WithContext` (codegen) | 2 dias |
| API-2 | Validação compile-time de enums | 1 dia |
| API-3 | `SessionConfig` builder fluente | 4h |
| API-4 | `ExecuteRawScriptReader` | 2h |
| API-6 | Hooks de lifecycle | 3h |
| — | Cobertura das funções PSADT faltantes (lifecycle) | 1 dia |
| — | Revisão completa da documentação | 1 dia |

---

## 8. Riscos e Considerações

1. **BC-1 é bloqueante:** Sem a correção, o projeto não funciona para o caso de uso principal (Quick Start). Prioridade absoluta.

2. **BC-2 pode causar OOM em produção:** RMM agents de longa duração com timeouts frequentes acumularão goroutines. Corrigir antes de qualquer uso em produção.

3. **BC-3 é intermitente:** O panic só ocorre se `Stop` e `emitOutput` correrem exatamente ao mesmo tempo. Difícil de reproduzir, mas causará crashes em produção.

4. **BC-7 retorna dados errados silenciosamente:** O caller recebe o nome do registry value em vez do valor. Pode causar bugs sutis em automação que depende de valores de registro.

5. **Compatibilidade PSADT:** As correções BC-7 precisam ser validadas contra a documentação do PSADT v4.1 — o comportamento de `Get-ADTRegistryKey` pode variar conforme o tipo de valor.

6. **Quebra de API:** As Fases 1-2 não quebram a API. A Fase 3 (MA-7) adiciona métodos, não quebra. A Fase 4 (API-3) pode mudar `OpenSession` — manter assinatura antiga como deprecated.

---

## 9. Próximos Passos

1. **Revisar este plano** com o time/usuário
2. **Aprovar Fase 1** para implementação imediata
3. **Criar branch** `fix/critical-bugs` para Fase 1
4. **Implementar e testar** cada item da Fase 1
5. **Validar** com Quick Start do README + testes de integração
6. **Decidir** sobre Fases 2-4 após validação

---

_Apresente este plano para revisão antes de iniciar a implementação._
