# Architecture — go-psadt

> Detailed technical reference for the internal design of the `go-psadt` library.

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Layer Model](#2-layer-model)
3. [Package Dependency Graph](#3-package-dependency-graph)
4. [PowerShell Process Lifecycle](#4-powershell-process-lifecycle)
5. [Communication Protocol](#5-communication-protocol)
6. [internal/cmdbuilder — Command Construction](#6-internalcmdbuilder--command-construction)
7. [internal/runner — Process Management](#7-internalrunner--process-management)
8. [internal/parser — Response Parsing](#8-internalparser--response-parsing)
9. [Client and Session Design](#9-client-and-session-design)
10. [Batch Execution & Context Propagation](#10-batch-execution--context-propagation)
11. [Live Output Streaming](#11-live-output-streaming)
12. [Type System](#12-type-system)
13. [Concurrency Model](#13-concurrency-model)
14. [Error Handling Chain](#14-error-handling-chain)
15. [Build Constraints and Platform Scope](#15-build-constraints-and-platform-scope)
16. [Design Decisions](#16-design-decisions)

---

## 1. System Overview

`go-psadt` is a **thin, statically-typed bridge** between Go and the [PSAppDeployToolkit (PSADT)](https://psappdeploytoolkit.com/) PowerShell module. Instead of invoking a new `powershell.exe` process per command (high latency, stateless), the library maintains a **single persistent PowerShell process** for the entire lifetime of a `Client`. Commands are sent via stdin and responses are read from stdout using a structured JSON protocol.

```
┌───────────────────────────────────────────────────────────────────┐
│  Go Application                                                   │
│                                                                   │
│  client, _ := psadt.NewClient(opts...)        ← Options pattern  │
│  session, _ := client.OpenSession(cfg)        ← Session config   │
│  err := session.StartMsiProcess(opts)         ← Typed method     │
│  session.Close(0)                             ← Exit code        │
└──────────────────────────────┬────────────────────────────────────┘
                               │  Compile-time-safe Go API
                               │
┌──────────────────────────────▼────────────────────────────────────┐
│  go-psadt (package psadt)                                         │
│                                                                   │
│  ┌────────────────┐  ┌───────────────────┐  ┌──────────────────┐ │
│  │  cmdbuilder    │  │  runner           │  │  parser          │ │
│  │                │  │                  │  │                  │ │
│  │  Struct + ps:  │  │  os.exec + pipes │  │  JSON → struct   │ │
│  │  tags          │  │  + sync.Mutex    │  │  + error types   │ │
│  │  → PS string   │  │  + bufio.Scanner │  │                  │ │
│  └───────┬────────┘  └────────┬──────────┘  └────────┬─────────┘ │
│          │                    │                       │           │
└──────────┼────────────────────┼───────────────────────┼───────────┘
           │                    │  stdin / stdout        │
           │    ┌───────────────▼──────────────────┐     │
           └───►│  powershell.exe  (persistent)    │◄────┘
                │                                  │
                │  -NoProfile -NonInteractive       │
                │  -ExecutionPolicy Bypass          │
                │  $ErrorActionPreference = 'Stop'  │
                │                                  │
                │  Import-Module PSAppDeployToolkit │
                │  Open-ADTSession ...              │
                │                                  │
                │  ─── per-command: ─────────────  │
                │  try { $r = <cmd>                 │
                │    @{Success=$true;Data=$r}       │
                │      | ConvertTo-Json             │
                │  } catch { @{Success=$false;...}  │
                │      | ConvertTo-Json }           │
                │  Write-Output '<<<PSADT_BEGIN>>>'  │
                │  Write-Output $__out              │
                │  Write-Output '<<<PSADT_END>>>'   │
                └──────────────────────────────────┘
```

---

## 2. Layer Model

The library is organized in three horizontal layers:

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer 3 — Public API                                           │
│                                                                 │
│  package psadt                                                  │
│  Client, Session, Option                                        │
│  ~110 methods across 25 .go files                               │
│  types/ — 24 strongly-typed option and result structs           │
└────────────────────────────────┬────────────────────────────────┘
                                 │  depends on
┌────────────────────────────────▼────────────────────────────────┐
│  Layer 2 — Internal Infrastructure                              │
│                                                                 │
│  internal/cmdbuilder   Build(), EscapeString(), FormatHashtable()│
│  internal/runner       Runner.Execute(), Runner.ExecuteVoid()   │
│  internal/parser       ParseResponse(), PSADTError              │
└────────────────────────────────┬────────────────────────────────┘
                                 │  OS primitives
┌────────────────────────────────▼────────────────────────────────┐
│  Layer 1 — Runtime                                              │
│                                                                 │
│  os/exec — PowerShell process                                   │
│  io.Pipe — stdin/stdout communication                           │
│  encoding/json — serialization                                  │
│  sync.Mutex — concurrency guard                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Rules enforced by this layering:**
- Layer 3 (`psadt`) may import `internal/*` and `types/`, never the reverse.
- `internal/*` packages are invisible to external consumers (Go module boundary).
- `types/` has **zero internal imports** — it is a pure data package.

---

## 3. Package Dependency Graph

```
github.com/pedrostefanogv/go-psadt   (root package)
├── imports types/
├── imports internal/cmdbuilder
├── imports internal/parser
└── imports internal/runner
        └── (no internal imports)

internal/cmdbuilder
├── imports reflect
├── imports fmt
└── imports strings

internal/runner
├── imports bufio
├── imports context
├── imports os/exec
├── imports sync
└── imports time

internal/parser
├── imports encoding/json
└── imports fmt

types/
└── (no imports — pure data definitions)
```

All internal packages are fully isolated. There are **no circular dependencies**.

---

## 4. PowerShell Process Lifecycle

### 4.1 Startup Sequence

`NewClient()` triggers the following ordered steps:

```
NewClient(opts...)
  │
  ├─ 1. Resolve options (PSPath, timeout, module name, min version)
  │
  ├─ 2. runner.New(cfg)
  │     ├─ detectPowerShell() — prefers pwsh.exe if WithPowerShell7(),
  │     │   otherwise auto-detects powershell.exe (Windows PowerShell 5.1)
  │     ├─ exec.Command(psPath, "-NoProfile", "-NonInteractive",
  │     │               "-ExecutionPolicy", "Bypass",
  │     │               "-OutputFormat", "Text", "-Command", "-")
  │     ├─ StdinPipe()  → r.stdin  (io.WriteCloser)
  │     ├─ StdoutPipe() → bufio.Scanner (10 MB buffer)
  │     ├─ StderrPipe() → r.stderr (captured but not surfaced directly)
  │     ├─ cmd.Start()
  │     ├─ Write: [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
  │     └─ Write: $ErrorActionPreference = 'Stop'
  │
  ├─ 3. runner.ImportModule("PSAppDeployToolkit")
  │     └─ Execute: Import-Module -Name 'PSAppDeployToolkit' -Force -ErrorAction Stop
  │
  └─ 4. runner.CheckModuleVersion("PSAppDeployToolkit", minVersion)
        └─ Execute: Get-Module -ListAvailable → compare [version] objects
```

The `-Command -` flag tells PowerShell to read commands from stdin as a REPL, keeping the process alive indefinitely.

### 4.2 Session Lifecycle

```
client.OpenSession(cfg)
  └─ Execute: Open-ADTSession -AppName ... -DeploymentType ... (all ps: tagged fields)

session.Method(opts)   ← many calls, all going to the same PS process
  └─ Execute / ExecuteVoid: <PSADT-verb>-ADT<Noun> -Param value ...

session.Close(exitCode)
  └─ Execute: Close-ADTSession -ExitCode <n>

client.Close()
  └─ runner.Stop()
        ├─ Write: exit   (graceful PS shutdown)
        └─ cmd.Wait() with 5-second kill timeout
```

### 4.3 Shutdown

`runner.Stop()` sends `exit` to stdin, then waits for the process to terminate. If it does not exit within 5 seconds, `cmd.Process.Kill()` is called. The `running` flag is set to `false` before the wait, so any concurrent call will receive an error immediately rather than hanging.

---

## 5. Communication Protocol

Every command follows a strict request/response cycle over the process's stdin/stdout pipes.

### 5.1 Command Wrapping

`WrapCommand` (for commands returning a value) and `WrapVoidCommand` (for side-effect commands) produce the following PowerShell fragment:

```powershell
# WrapCommand(psCommand)
try {
    $result = <psCommand>
    $__out = @{
        Success = $true
        Data    = $result
        Error   = $null
    } | ConvertTo-Json -Depth 10 -Compress
} catch {
    $__out = @{
        Success = $false
        Data    = $null
        Error   = @{
            Message    = $_.Exception.Message
            Type       = $_.Exception.GetType().FullName
            StackTrace = $_.ScriptStackTrace
        }
    } | ConvertTo-Json -Depth 10 -Compress
}
Write-Output '<<<PSADT_BEGIN>>>'
Write-Output $__out
Write-Output '<<<PSADT_END>>>'
```

`WrapVoidCommand` is identical except the command is executed as a statement (no `$result =` assignment), and `Data` is always `$null` on success.

### 5.2 Marker Extraction

`runner.readResponse()` uses a `bufio.Scanner` (10 MB line buffer) to read stdout line by line:

```
Line read     │  State machine action
──────────────┼────────────────────────────────────────────────────
<<<PSADT_BEGIN>>> │  Set inResponse = true
<json line>   │  Append to jsonLines (while inResponse == true)
<<<PSADT_END>>>   │  Set inResponse = false; join jsonLines → return bytes
other line    │  Forward to live output stream (see §11 Live Output Streaming)
EOF / nil     │  Return scanner error (process died)
```

The scanner uses `strings.Join(jsonLines, "")` so multi-line JSON (if any) is reassembled correctly.

### 5.3 JSON Envelope

Every response is a single-line compressed JSON object conforming to:

```json
{
  "Success": true,
  "Data": <arbitrary JSON value>,
  "Error": null
}
```

or on failure:

```json
{
  "Success": false,
  "Data": null,
  "Error": {
    "Message":    "string",
    "Type":       "fully.qualified.DotNetExceptionType",
    "StackTrace": "At line:... in ..."
  }
}
```

`-Depth 10` ensures nested objects (e.g., `EnvironmentInfo`) are fully serialized. `-Compress` keeps the output as a single line, which is essential for the line-by-line scanner.

### 5.4 Timeout Handling

Each `Execute` call derives a timeout from the context:

```go
timeout := r.timeout              // client-level default
if deadline, ok := ctx.Deadline(); ok {
    timeout = time.Until(deadline) // context deadline overrides
}
```

A goroutine reads the next scanner line and sends it to a channel. The `select` statement races that channel against `ctx.Done()` and a `time.After(timeout)` channel, whichever fires first wins. On timeout, the next write to stdin will fail, triggering `r.running = false`.

---

## 6. internal/cmdbuilder — Command Construction

### 6.1 Role

`cmdbuilder` translates a Go struct into a PowerShell parameter string:

```
types.MsiProcessOptions{
    Action:   "Install",
    FilePath: "C:\\setup.msi",
    PassThru: true,
}
→  "Start-ADTMsiProcess -Action 'Install' -FilePath 'C:\setup.msi' -PassThru"
```

### 6.2 `Build()` — Reflection Engine

```go
func Build(cmdName string, opts interface{}) string
```

Uses `reflect` to iterate over all exported struct fields. For each field:

1. Read the `ps` struct tag: `ps:"ParamName"` or `ps:"ParamName,switch"`
2. Skip fields with `ps:""` or `ps:"-"`
3. Call `formatParam(paramName, fieldValue, isSwitch)`

Result is `strings.Join([]string{cmdName, param1, param2, ...}, " ")`.

### 6.3 Type Dispatch in `formatParam()`

| Go type | PowerShell output | Notes |
|---|---|---|
| `string` | `-Name 'value'` | Escaped via `EscapeString()` |
| `int`, `int64`, ... | `-Count 42` | Zero value is omitted |
| `uint`, `uint64`, ... | `-Size 1024` | Zero value is omitted |
| `float64` | `-Rate 1.5` | Zero value is omitted |
| `bool` (non-switch) | `-Flag $true` | `false` is omitted |
| `bool` (switch) | `-PassThru` | Only emitted when `true` |
| `[]string` | `-Names 'a','b'` | Quoted and comma-joined |
| `[]int` | `-Codes 0,3010` | Comma-joined |
| `[]struct` | `-Procs @{Name='x'}` | Each elem → `FormatHashtable()` |
| `map[string]string` | `-Props @{k='v'}` | `FormatHashtable()` on map |
| `interface{}` | delegates to inner type | Unwraps to concrete value |
| `struct` | delegates to `formatStructParam` | Emits as hashtable |

Zero/nil/empty values are **always omitted**, which maps naturally to optional PowerShell parameters.

### 6.4 `EscapeString()`

Strings are wrapped in **single quotes** (safest PowerShell quoting — no variable expansion). Embedded single quotes are doubled: `'it''s'`. Special literals (`$true`, `$false`, `$null`, `$variable`, pure numbers) are passed **unquoted** to avoid double-wrapping.

### 6.5 `FormatHashtable()`

Struct values passed as parameters (e.g., `ProcessDefinition`) are converted to PS hashtable syntax:

```
@{Name='widget'; Description='Widget process'}
```

Uses the `ps:` tag on each field of the nested struct to determine the key name.

### 6.6 `FormatScriptBlock()`

Used by methods that accept a script block string (e.g., `InvokeAllUsersRegistryAction`). Wraps the string in `{ }`:

```
{ Set-ItemProperty -Path ... }
```

---

## 7. internal/runner — Process Management

### 7.1 `Runner` struct

```go
type Runner struct {
    cmd           *exec.Cmd         // the powershell.exe process
    stdin         io.Writer         // StdinPipe — command input
    stdoutScanner *bufio.Scanner    // line-by-line stdout reader (10 MB buffer)
    stderr        io.Reader         // StderrPipe — forwarded to live stream
    mu            sync.Mutex        // serializes all command dispatches
    running       bool              // false after Stop() or process death
    timeout       time.Duration     // per-command default timeout
    psPath        string            // resolved PS executable path
    liveOutputCh  chan string       // buffered channel for live stdout/stderr lines
    onOutput      OutputLineCallback // optional synchronous output callback
}
```

### 7.2 `detectPowerShell()`

Auto-detection order:
1. If `UsePowerShell7 == true` → `pwsh.exe`
2. Otherwise → `powershell.exe` (guaranteed present on all supported Windows versions)

The path is resolved by the OS via `exec.LookPath` at process startup.

### 7.3 `Execute()` variants

```go
func (r *Runner) Execute(ctx, psCommand)       // values → WrapCommand + executeWrapped
func (r *Runner) ExecuteVoid(ctx, psCommand)   // void → WrapVoidCommand + executeWrapped
func (r *Runner) ExecuteBatch(ctx, commands)   // multiple → joined + WrapCommand
func (r *Runner) ExecuteRaw(ctx, wrappedCmd)   // already wrapped → executeWrapped
func (r *Runner) ExecuteRawVoid(ctx, wrappedCmd) // already wrapped void → executeWrapped
```

`executeWrapped`:
1. Acquires `r.mu` (serializes concurrent calls)
2. Writes wrapped command to `r.stdin` via `fmt.Fprintln`
3. Calls `readResponse(ctx)` → returns `[]byte` JSON

`ExecuteBatch` joins multiple PS commands with `; ` and wraps them once, reducing
round-trips for multi-step operations. `ExecuteRaw`/`ExecuteRawVoid` accept
already-wrapped commands — useful for custom scripts that callers construct manually.

### 7.4 `IsAlive()` / `Heartbeat()`

```go
func (r *Runner) IsAlive() bool
func (r *Runner) Heartbeat(ctx) error
```

`IsAlive()` returns `r.running` (non-blocking). `Heartbeat()` executes `$true` through the full round-trip to confirm the process is responsive.

### 7.5 `LiveOutput()` / `drainStderr()`

The runner starts a background goroutine (`drainStderr`) that continuously reads
stderr and forwards every line to `liveOutputCh`. Non-marker stdout lines are
also forwarded via `emitOutput()` in `readResponse()`. Both channels eventually
arrive at the same `liveOutputCh` channel, giving the caller a unified stream.

See §11 Live Output Streaming for the full design.

### 7.6 `ImportModule()` / `CheckModuleVersion()`

```go
func (r *Runner) ImportModule(ctx, moduleName) error
func (r *Runner) CheckModuleVersion(ctx, moduleName, minVersion) (string, error)
```

Both use `Execute()` internally. `CheckModuleVersion` returns the actual installed version string, which `NewClient` validates against the configured minimum.

---

## 8. internal/parser — Response Parsing

### 8.1 `Response` struct

```go
type Response struct {
    Success bool            `json:"Success"`
    Data    json.RawMessage `json:"Data"`  // deferred parse
    Error   *ErrorDetail    `json:"Error"`
}

type ErrorDetail struct {
    Message    string `json:"Message"`
    Type       string `json:"Type"`
    StackTrace string `json:"StackTrace"`
}
```

`Data` is kept as `json.RawMessage` so the outer envelope can be parsed cheaply, and `Data` is only deserialized into the concrete target type when needed.

### 8.2 Parse Pipeline

```
[]byte (raw JSON from stdout)
    │
    ▼  Parse()
Response{Success, Data, Error}
    │
    ▼  ParseInto(resp, &target)
    │    ├─ if !Success → NewPSADTError(resp.Error)
    │    ├─ if Data == null → return nil (void success)
    │    └─ json.Unmarshal(resp.Data, target)
    │
    ▼
*types.XxxResult or primitive
```

`ParseResponse(data, &target)` combines `Parse` + `ParseInto` as a single call. All 105 public methods use either `ParseResponse`, `ParseBool`, `ParseString`, `ParseUint64`, or `CheckSuccess` (void commands).

### 8.3 `PSADTError`

```go
type PSADTError struct {
    Message    string
    Type       string  // e.g. "System.UnauthorizedAccessException"
    StackTrace string  // PS call stack
    ExitCode   int     // populated for process-level errors
}
```

Implements `error`. Callers can inspect with:

```go
if psErr, ok := parser.IsPSADTError(err); ok {
    // access psErr.Type, psErr.StackTrace
}
parser.IsRebootRequired(err)  // ExitCode 3010 or 1641
parser.IsUserCancelled(err)   // ExitCode 1602
```

---

## 9. Client and Session Design

### 9.1 `Client`

```go
type Client struct {
    runner     *runner.Runner         // owns the PS process
    logger     *slog.Logger           // structured logging (log/slog, Go 1.21+)
    moduleName string                 // "PSAppDeployToolkit"
    minVersion string                 // "4.1.0"
    timeout    time.Duration          // default per-command timeout
    envMu      sync.Mutex             // guards envCache
    envCache   *types.EnvironmentInfo // cached result of GetEnvironment()
    envCached  bool                   // true when envCache is valid
}
```

`Client` is **not** safe for concurrent `OpenSession` calls from multiple goroutines. The underlying `runner.Mutex` ensures command serialization, but `OpenSession` / `CloseSession` must be sequenced by the caller.

### 9.2 Options Pattern

```go
type Option func(*clientConfig)
```

Configuration is accumulated in a private `clientConfig` struct during `NewClient`, then applied in a single step. This avoids partially-initialized `Client` states and makes options composable:

```go
client, _ := psadt.NewClient(
    psadt.WithTimeout(5 * time.Minute),
    psadt.WithPowerShell7(),
    psadt.WithMinModuleVersion("4.2.0"),
    psadt.WithLogger(slog.Default()),
)
```

### 9.3 `Session`

```go
type Session struct {
    client *Client             // back-pointer for options and logger
    runner *runner.Runner      // shared with client (same PS process)
    config types.SessionConfig // the configuration passed to Open-ADTSession
    ctx    context.Context     // embedded context; nil means use default
    closed bool                // prevents double-close
}
```

`Session` holds a reference to the **same** `runner.Runner` as `Client`. All commands in a session go through the same persistent process, preserving the PSADT session state (open log file, app name, etc.) between calls.

The embedded `ctx` field enables **context propagation without method duplication** (see §10).

### 9.4 `execute()` / `executeVoid()` / `getContext()` helpers

Session methods call private helpers instead of accessing `runner` directly:

```go
func (s *Session) execute(ctx, cmd) ([]byte, error) {
    return s.runner.Execute(ctx, cmd)
}
func (s *Session) executeVoid(ctx, cmd) error {
    data, err := s.runner.ExecuteVoid(ctx, cmd)
    if err != nil { return err }
    return parser.CheckSuccess(data)
}
func (s *Session) getContext() (context.Context, context.CancelFunc) {
    if s.ctx != nil {
        return context.WithCancel(s.ctx)
    }
    return s.client.defaultContext()
}
```

`getContext()` is the **single context resolution point** used by all ~100 session methods. If the session has an embedded context (set via `WithContext()`), it returns a child context with cancel propagation. Otherwise it falls back to the client's default timeout. This eliminates the need to pass `ctx` to every method individually.

### 9.5 Environment Caching

`Client.GetEnvironment()` caches the ~90 PSADT environment variables after the first call. Subsequent calls return the cached snapshot without a PowerShell round-trip. Call `InvalidateEnvCache()` to force a fresh fetch (e.g., after changing machine state).

### 9.6 Client Reconnection

For long-running RMM agents, `Client.Reconnect(ctx)` tears down the current runner and starts a fresh PowerShell process, re-importing the PSADT module. `IsAlive()` allows health checks before every operation.

---

## 10. Batch Execution & Context Propagation

### 10.1 `ExecuteBatch` — Multi-Command Round-Trip

Multiple PowerShell commands can be joined and sent in a single stdin write, reducing latency for multi-step operations:

```go
// Without batch: 3 round-trips (3 × pipe latency)
name, _ := session.GetRegistryKeyString(key, "DisplayName")
ver, _  := session.GetRegistryKeyString(key, "DisplayVersion")
ok, _   := session.TestNetworkConnection()

// With batch: 1 round-trip
data, _ := session.ExecuteBatch(ctx, []string{
    "(Get-ADTRegistryKey -Key ...).DisplayName",
    "(Get-ADTRegistryKey -Key ...).DisplayVersion",
    "Test-ADTNetworkConnection",
})
```

Implementation in `runner.ExecuteBatch()`:

```go
func (r *Runner) ExecuteBatch(ctx, commands) ([]byte, error) {
    joined := strings.Join(commands, "; ")
    return r.executeWrapped(ctx, WrapCommand(joined))
}
```

All commands run inside a single `try/catch` wrapper. The return value is the JSON output of the **last** command in the sequence.

### 10.2 `WithContext` — Zero-Duplication Context Propagation

Instead of creating `XxxWithContext` variants for every method, the `Session` struct carries an optional embedded context:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
result, _ := session.WithContext(ctx).GetApplication(opts)
```

`WithContext()` returns a **shallow copy** of the session with the context set. All subsequent method calls on that copy automatically use the propagated context via `getContext()` — no method duplication needed. The original session is not modified.

### 10.3 `ExecuteRawScript` — Escape Hatch

For PSADT operations not yet wrapped by the Go API, callers can execute arbitrary PowerShell scripts in the session context:

```go
err := session.ExecuteRawVoidScript(ctx, `
    Write-ADTLogEntry -Message "Custom logic" -Source "rmm-agent" -Severity 1
`)
```

`ExecuteRawScript` returns raw JSON bytes; `ExecuteRawVoidScript` checks for errors only. Both are also available at the `Client` level for scripts that don't require an open session.

### 10.4 Typed Registry Access

To avoid `interface{}` return types, `GetRegistryKeyString` and `GetRegistryKeyDWord` provide typed convenience wrappers:

```go
version, _ := session.GetRegistryKeyString(`HKLM\SOFTWARE\Contoso`, "Version")
config, _  := session.GetRegistryKeyDWord(`HKLM\SOFTWARE\Contoso`, "ConfigFlag")
```

---

## 11. Live Output Streaming

### 11.1 Architecture

The runner exposes real-time output streaming through two mechanisms:

**Channel-based (non-blocking):**
```go
type Runner struct {
    // ...
    liveOutputCh chan string  // buffered channel (256 lines)
}
func (r *Runner) LiveOutput() <-chan string
```

**Callback-based (synchronous):**
```go
type OutputLineCallback func(line string)
type Config struct {
    // ...
    OnOutput OutputLineCallback
}
```

### 11.2 Data Flow

```
powershell.exe
    │
    ├─ stdout ──→ bufio.Scanner ──→ readResponse()
    │                                  │
    │                    (non-marker lines)
    │                                  │
    │                                  ▼
    │                            emitOutput(line)
    │                                  │
    ├─ stderr ──→ drainStderr() goroutine ─┘
    │
    ▼
liveOutputCh (chan string, 256-buffered)
    │
    ├─ LiveOutput() → caller goroutine reads
    └─ onOutput callback → synchronous notification
```

### 11.3 `drainStderr()`

A background goroutine launched during `start()` continuously reads stderr and forwards every line to the shared `liveOutputCh`. If the channel is full, lines are silently dropped to avoid blocking the runner.

### 11.4 `emitOutput()`

Called from `readResponse()` for every stdout line that falls **outside** the `<<<PSADT_BEGIN>>>` / `<<<PSADT_END>>>` markers. This captures PSADT log output (Write-ADTLogEntry, Write-Host, etc.) in real time.

### 11.5 Usage Pattern (RMM Agent)

```go
// Start streaming goroutine
liveCh := session.LiveOutput()
go func() {
    for line := range liveCh {
        fmt.Printf("[PSADT] %s\n", line)
    }
}()

// Run a long installation — logs stream in real time
session.StartMsiProcess(types.MsiProcessOptions{
    Action:   types.MsiInstall,
    FilePath: "setup.msi",
})
// The channel closes when client.Close() is called
```

---

## 12. Type System

### 12.1 Struct Tags

The type system uses **two** struct tag families:

| Tag | Package | Purpose |
|---|---|---|
| `` `ps:"Name"` `` | `cmdbuilder` | Maps Go field to PS parameter name |
| `` `ps:"Name,switch"` `` | `cmdbuilder` | Maps Go bool to PS switch parameter |
| `` `json:"Name"` `` | `parser` | Maps PS JSON response field to Go field |

Many structs have **both** tags on the same field (bidirectional types used as both input and output), but most separate input structs (`XxxOptions`) from output structs (`XxxResult`, `XxxInfo`).

### 12.2 Input Structs (Options)

Pattern: all-optional fields, zero values are omitted from the command.

```go
// types/process.go
type StartProcessOptions struct {
    FilePath         string             `ps:"FilePath"`
    ArgumentList     []string           `ps:"ArgumentList"`
    WindowStyle      ProcessWindowStyle `ps:"WindowStyle"`
    NoWait           bool               `ps:"NoWait,switch"`
    PassThru         bool               `ps:"PassThru,switch"`
    // ...
}
```

### 12.3 Output Structs (Results/Info)

Pattern: JSON tags matching PSADT's PowerShell output property names.

```go
// types/process.go
type ProcessResult struct {
    ExitCode    int    `json:"ExitCode"`
    StdOut      string `json:"StdOut"`
    StdErr      string `json:"StdErr"`
    Interleaved string `json:"Interleaved"`
}
```

### 12.4 Enums as Typed Strings

PSADT parameters that accept a fixed set of values are typed as Go string aliases:

```go
type DeploymentType string
const (
    DeployInstall   DeploymentType = "Install"
    DeployUninstall DeploymentType = "Uninstall"
    DeployRepair    DeploymentType = "Repair"
)
```

This provides IDE autocompletion and compile-time validation without runtime overhead. The `cmdbuilder` formats them as quoted strings.

### 12.5 `EnvironmentInfo` — Hierarchical Snapshot

`Client.GetEnvironment()` returns a single `types.EnvironmentInfo` that aggregates ~90 PSADT variables into a structured hierarchy:

```go
type EnvironmentInfo struct {
    Toolkit    ToolkitInfo    // PSADT version, module path, deployment type
    OS         OSEnvironment  // Windows version, architecture, service pack
    Machine    MachineInfo    // Computer name, domain, is 64-bit
    PowerShell PSVersionInfo  // PS version, CLR version
    Users      UsersInfo      // current user, logged-on users, SYSTEM check
    Paths      SystemPaths    // Windows, System32, ProgramFiles, temp dirs
    Office     OfficeInfo     // Office version and path if installed
    // ...
}
```

Built by a single PowerShell hashtable serialization in `environment.go` rather than 90 individual `Execute` calls. The result is **cached** after the first call; `InvalidateEnvCache()` forces a refresh.

---

## 13. Concurrency Model

```
Goroutine A          Goroutine B
    │                    │
    │  session.CopyFile  │  session.SetRegistryKey
    │                    │
    ▼                    ▼
runner.Execute()    runner.Execute()
    │                    │
    └──────┬─────────────┘
           │
     sync.Mutex.Lock()
           │
     (only one goroutine proceeds)
           │
     Write to stdin
     Read from stdout
           │
     sync.Mutex.Unlock()
```

**Key constraints:**

- `sync.Mutex` inside `Runner.executeWrapped()` ensures only **one command runs at a time** on a single `Client`.
- A `Client` (and its `Session`) must not be shared across goroutines without external synchronization, or commands will be serialized unpredictably.
- For parallel deployments, create **separate `Client` instances** — each gets its own PowerShell process.
- The timeout goroutine in `readResponse()` is short-lived and does not hold the mutex.

---

## 14. Error Handling Chain

```
PSADT throws a terminating error in PowerShell
        │
        ▼
catch { $_ } block captures it
        │
        ▼
Serialized to JSON: { Success:false, Error:{ Message, Type, StackTrace } }
        │
        ▼
Delimited by <<<PSADT_BEGIN>>> / <<<PSADT_END>>>
        │
        ▼
runner.readResponse() returns the raw JSON bytes
        │
        ▼
parser.Parse() → Response{Success:false, Error:*ErrorDetail}
        │
        ▼
parser.ParseInto() / ParseResponse()
    └─ !Success → NewPSADTError(resp.Error) → return *PSADTError
        │
        ▼
Session method returns (nil, *PSADTError) to the Go caller
        │
        ▼
Go caller can:
    ├─ Check IsPSADTError(err)     → inspect Type, StackTrace
    ├─ Check IsRebootRequired(err) → exit code 3010 / 1641
    ├─ Check IsUserCancelled(err)  → exit code 1602
    └─ fmt.Errorf("... %w", err)  → wrap and propagate
```

**Transport-level errors** (pipe broken, process died, timeout) are returned as plain Go errors wrapping the OS error — they are distinct from `*PSADTError` and can be differentiated with `IsPSADTError`.

---

## 15. Build Constraints and Platform Scope

Every `.go` file in the library (except `go.mod`) carries:

```go
//go:build windows
```

This directive:
- Prevents compilation on Linux/macOS, where `os/exec` with `powershell.exe` would be meaningless.
- Allows the library to be listed as a dependency in cross-platform projects without causing build failures on non-Windows targets.
- Is applied to `types/` files as well, because some types reference Windows-specific concepts (registry hives, service start modes, etc.).

The `examples/` programs also carry the constraint. There are no `_test.go` files in the root package yet; the three internal packages have unit tests that compile on Windows.

---

## 16. Design Decisions

### D1 — Persistent process over per-command invocation

Starting `powershell.exe -Command "..."` per call costs ~300–500 ms on Windows. A persistent process amortizes startup to a one-time cost at `NewClient()`. More importantly, PSADT **requires** a session to be open for most operations (`Open-ADTSession` sets global module state), which is only feasible within a single PS process lifetime.

### D2 — Delimited markers over named pipes or sockets

Using `<<<PSADT_BEGIN>>>` / `<<<PSADT_END>>>` markers on stdout avoids the need for a separate IPC channel, server socket, or named pipe setup. The tradeoff is that `Write-Host` output inside a command silently appears between commands and is discarded by the marker scanner — which is acceptable because PSADT logs to its own log file rather than stdout.

### D3 — `try/catch` wrapper per command

Rather than relying on PowerShell's global `$ErrorActionPreference = 'Stop'` alone, each command is individually wrapped in `try/catch`. This ensures errors from every command path (including deep PSADT internals) are caught and converted to the JSON error envelope, not leaked as unstructured stderr output.

### D4 — Reflection-based `cmdbuilder` over code generation

Using `reflect` avoids maintaining a code generator or 105 hand-written formatting functions. The `ps:` tag convention mirrors the well-known `json:` and `xml:` patterns, making the struct definitions self-documenting. The reflection cost is negligible compared to the stdin write and PS execution time.

### D5 — `json.RawMessage` for deferred Data parsing

The outer `Response{Success, Data, Error}` envelope is parsed eagerly to determine success/failure. `Data` is kept as `json.RawMessage` and only deserialized into the caller-provided target type. This avoids a double-unmarshal and means the infrastructure code never needs to know the shape of the data.

### D6 — `*slog.Logger` as the logging interface

`log/slog` (Go 1.21 standard library) provides structured logging without external dependencies. Callers can inject any `slog.Handler` (JSON, text, third-party) via `WithLogger()`. If no logger is provided, `slog.Default()` is used, respecting the application's global logging configuration.

### D7 — `types/` as a zero-dependency leaf package

Keeping all public types in a separate package with no internal imports allows consumers to import only `types/` for struct definitions (e.g., in shared configuration code) without pulling in the process management infrastructure.

### D8 — Embedded context over method duplication

Adding an explicit `ctx` parameter to every session method would double the API surface (~200 methods). Instead, the `Session` struct carries an optional embedded context field set by `WithContext()`. A single `getContext()` helper resolves the effective context for every call: the session's embedded context if present, otherwise the client's default timeout. This provides fine-grained deadline/cancellation control at zero duplication cost.

### D9 — Buffered channel + dual-source live output

Live output streaming merges two sources — stdout (non-marker lines from `readResponse`) and stderr (`drainStderr` goroutine) — into a single buffered channel. The 256-line buffer prevents the PS process from blocking on write while the consumer reads. An optional synchronous `OnOutput` callback runs in the same goroutine as the scanner/stderr reader, suitable for write-through logging. The channel is closed on `Stop()`, providing a clean termination signal for consumer goroutines.

### D10 — Batch execution over individual calls

Joining multiple PS commands with `; ` and wrapping them once under a single `try/catch` reduces round-trips for multi-step operations (e.g., validation: check app name + version + network in one call). The trade-off is that error handling is all-or-nothing — a failure in any command aborts the batch. For critical sequences, separate `Execute` calls remain the safer option.
