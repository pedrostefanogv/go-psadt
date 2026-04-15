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
10. [Type System](#10-type-system)
11. [Concurrency Model](#11-concurrency-model)
12. [Error Handling Chain](#12-error-handling-chain)
13. [Build Constraints and Platform Scope](#13-build-constraints-and-platform-scope)
14. [Design Decisions](#14-design-decisions)

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
│  ~105 methods across 20 .go files                               │
│  types/ — 22 strongly-typed option and result structs           │
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
other line    │  Discard (PS informational output, Write-Host, etc.)
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
    stderr        io.Reader         // StderrPipe — captured but not parsed
    mu            sync.Mutex        // serializes all command dispatches
    running       bool              // false after Stop() or process death
    timeout       time.Duration     // per-command default timeout
    psPath        string            // resolved PS executable path
}
```

### 7.2 `detectPowerShell()`

Auto-detection order:
1. If `UsePowerShell7 == true` → `pwsh.exe`
2. Otherwise → `powershell.exe` (guaranteed present on all supported Windows versions)

The path is resolved by the OS via `exec.LookPath` at process startup.

### 7.3 `Execute()` vs `ExecuteVoid()`

Both methods delegate to `executeWrapped(ctx, wrappedCmd)`:

```go
func (r *Runner) Execute(ctx, psCommand)     { executeWrapped(ctx, WrapCommand(psCommand)) }
func (r *Runner) ExecuteVoid(ctx, psCommand) { executeWrapped(ctx, WrapVoidCommand(psCommand)) }
```

`executeWrapped`:
1. Acquires `r.mu` (serializes concurrent calls)
2. Writes wrapped command to `r.stdin` via `fmt.Fprintln`
3. Calls `readResponse(ctx)` → returns `[]byte` JSON

Return value is `[]byte` in both cases; the `parser` layer handles the distinction.

### 7.4 `IsAlive()` / `Heartbeat()`

```go
func (r *Runner) IsAlive() bool
func (r *Runner) Heartbeat(ctx) error
```

`IsAlive()` returns `r.running` (non-blocking). `Heartbeat()` executes `$true` through the full round-trip to confirm the process is responsive.

### 7.5 `ImportModule()` / `CheckModuleVersion()`

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
    runner     *runner.Runner   // owns the PS process
    logger     *slog.Logger     // structured logging (log/slog, Go 1.21+)
    moduleName string           // "PSAppDeployToolkit"
    minVersion string           // "4.1.0"
    timeout    time.Duration    // default per-command timeout
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
    closed bool                // prevents double-close
}
```

`Session` holds a reference to the **same** `runner.Runner` as `Client`. All commands in a session go through the same persistent process, preserving the PSADT session state (open log file, app name, etc.) between calls.

### 9.4 `execute()` / `executeVoid()` helpers

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
```

`defaultContext()` on `Client` provides a context pre-configured with the client's default timeout:

```go
func (c *Client) defaultContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), c.timeout)
}
```

---

## 10. Type System

### 10.1 Struct Tags

The type system uses **two** struct tag families:

| Tag | Package | Purpose |
|---|---|---|
| `` `ps:"Name"` `` | `cmdbuilder` | Maps Go field to PS parameter name |
| `` `ps:"Name,switch"` `` | `cmdbuilder` | Maps Go bool to PS switch parameter |
| `` `json:"Name"` `` | `parser` | Maps PS JSON response field to Go field |

Many structs have **both** tags on the same field (bidirectional types used as both input and output), but most separate input structs (`XxxOptions`) from output structs (`XxxResult`, `XxxInfo`).

### 10.2 Input Structs (Options)

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

### 10.3 Output Structs (Results/Info)

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

### 10.4 Enums as Typed Strings

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

### 10.5 `EnvironmentInfo` — Hierarchical Snapshot

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

This is built by a single `Get-Variable` sweep in `environment.go` rather than 90 individual `Execute` calls.

---

## 11. Concurrency Model

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

## 12. Error Handling Chain

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

## 13. Build Constraints and Platform Scope

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

## 14. Design Decisions

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
