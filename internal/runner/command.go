//go:build windows

package runner

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"
)

// nextID returns the next unique command id for marker tagging.
func (r *Runner) nextID() uint64 {
	return r.cmdSeq.Add(1)
}

// beginMarkerFor returns the id-tagged begin marker of a command.
func beginMarkerFor(id uint64) string {
	return fmt.Sprintf("<<<PSADT_BEGIN:%d>>>", id)
}

// outLine is a single line delivered by the stdout pump.
type outLine struct {
	line string
	err  error
	eof  bool
}

// Execute sends a PowerShell command and returns the JSON response bytes.
// The command is automatically wrapped with try/catch and delimiters.
func (r *Runner) Execute(ctx context.Context, psCommand string) ([]byte, error) {
	id := r.nextID()
	return r.executeWrapped(ctx, WrapCommandWithID(psCommand, id), beginMarkerFor(id))
}

// ExecuteVoid sends a PowerShell command that returns no data.
func (r *Runner) ExecuteVoid(ctx context.Context, psCommand string) ([]byte, error) {
	id := r.nextID()
	return r.executeWrapped(ctx, WrapVoidCommandWithID(psCommand, id), beginMarkerFor(id))
}

// ExecuteRaw runs an already-wrapped PowerShell command and returns the raw
// JSON response bytes. The caller is responsible for providing properly wrapped
// commands with markers. This is the escape hatch for custom scripting.
func (r *Runner) ExecuteRaw(ctx context.Context, rawWrappedCmd string) ([]byte, error) {
	return r.executeWrapped(ctx, rawWrappedCmd, BeginMarker)
}

// ExecuteRawVoid runs an already-wrapped PowerShell command that produces
// no meaningful data (e.g., void commands).
func (r *Runner) ExecuteRawVoid(ctx context.Context, rawWrappedCmd string) error {
	_, err := r.executeWrapped(ctx, rawWrappedCmd, BeginMarker)
	return err
}

// ExecuteBatch runs multiple PowerShell commands in a single round-trip.
// All commands are joined with semicolons and wrapped once, reducing latency
// for multi-step operations. Returns raw JSON response bytes.
func (r *Runner) ExecuteBatch(ctx context.Context, commands []string) ([]byte, error) {
	if len(commands) == 0 {
		return nil, nil
	}
	joined := strings.Join(commands, "; ")
	id := r.nextID()
	return r.executeWrapped(ctx, WrapCommandWithID(joined, id), beginMarkerFor(id))
}

// executeWrapped sends a raw PowerShell command string (already wrapped) and reads the response.
func (r *Runner) executeWrapped(ctx context.Context, wrappedCmd, expectedBegin string) ([]byte, error) {
	start := time.Now()

	resp, err := r.dispatch(ctx, wrappedCmd, expectedBegin)

	// The metrics hook runs outside the dispatch mutex — it must never call
	// back into the Runner or it will deadlock.
	r.onCommandMu.RLock()
	hook := r.onCommand
	r.onCommandMu.RUnlock()
	if hook != nil {
		hook(wrappedCmd, time.Since(start), err)
	}

	return resp, err
}

// dispatch performs the mutex-protected command exchange with PowerShell.
func (r *Runner) dispatch(ctx context.Context, wrappedCmd, expectedBegin string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil, fmt.Errorf("PowerShell runner is not running")
	}

	// Verifica ctx antes de escrever — evita escrever em stdin de um processo
	// que nunca será lido se o ctx já estiver cancelado.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Write command to stdin
	if _, err := fmt.Fprintln(r.stdin, wrappedCmd); err != nil {
		r.running = false
		return nil, fmt.Errorf("failed to write command to PowerShell: %w", err)
	}

	// Read response between markers
	return r.readResponse(ctx, expectedBegin)
}

// ensurePump starts the single persistent stdout pump goroutine. The pump is
// started once for the lifetime of the runner: a per-call reader goroutine
// would race with the next call over the shared scanner and steal its lines.
func (r *Runner) ensurePump() {
	r.pumpOnce.Do(func() {
		go r.pump()
	})
}

// pump continuously reads stdout and delivers marker-delimited lines to
// outCh. Lines outside a response (PSADT logs, banners) are forwarded to the
// live output stream instead. Marker lines and response payloads are never
// dropped — only live-output lines are discardable under backpressure.
func (r *Runner) pump() {
	defer close(r.outCh)

	scanner := r.stdoutScanner
	if scanner == nil {
		return
	}

	inResponse := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "<<<PSADT_BEGIN") {
			inResponse = true
			select {
			case r.outCh <- outLine{line: line}:
			case <-r.dead:
				return
			}
			continue
		}
		if strings.HasPrefix(line, "<<<PSADT_END") {
			inResponse = false
			select {
			case r.outCh <- outLine{line: line}:
			case <-r.dead:
				return
			}
			continue
		}
		if inResponse {
			// Response payload — must not be dropped.
			select {
			case r.outCh <- outLine{line: line}:
			case <-r.dead:
				return
			}
			continue
		}

		// Live output (PSADT logs etc.): emit and drop if nobody keeps up.
		r.emitOutput(line)
	}

	if err := scanner.Err(); err != nil {
		if err == bufio.ErrTooLong {
			err = fmt.Errorf("PowerShell response too large: %w", err)
		}
		select {
		case r.outCh <- outLine{err: err}:
		case <-r.dead:
		}
		return
	}

	select {
	case r.outCh <- outLine{eof: true}:
	case <-r.dead:
	}
}

// readResponse reads stdout until it finds the begin/end markers, extracting
// the JSON between them.
func (r *Runner) readResponse(ctx context.Context, expectedBegin string) ([]byte, error) {
	r.ensurePump()

	// Calcula o timeout: usa deadline do ctx se houver, senão r.timeout.
	// Se r.timeout for 0, usa defaultTimeout (30s) como fallback.
	timeout := r.timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining < timeout {
			timeout = remaining
		}
	}
	// Garante timeout positivo mesmo se ctx não tiver deadline.
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	var jsonLines []string
	inResponse := false

	// Timer é criado uma única vez e resetado a cada iteração. Usar time.After
	// dentro deste loop alocava um timer novo (não-cancelável) por linha, que
	// se acumulava até disparar em agents RMM de longa duração.
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			// O comando em execução ainda pode emitir a resposta completa depois.
			// Marca o stream como dessincronizado para que o próximo readResponse
			// descarte a resposta remanescente em vez de atribuí-la ao comando errado.
			r.desynced = true
			return nil, ctx.Err()
		case <-timer.C:
			r.desynced = true
			return nil, fmt.Errorf("timeout waiting for PowerShell response after %v", timeout)
		case result, ok := <-r.outCh:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(timeout)

			if !ok {
				r.running = false
				return nil, fmt.Errorf("PowerShell process ended before completing response")
			}
			if result.eof {
				r.running = false
				return nil, fmt.Errorf("PowerShell process ended before completing response")
			}
			if result.err != nil {
				r.running = false
				return nil, fmt.Errorf("error reading PowerShell output: %w", result.err)
			}

			line := result.line

			// Após um timeout/cancelamento anterior, o stream pode conter a
			// resposta remanescente (marcadores inclusive) do comando anterior.
			// Descarta tudo até o marcador Begin COM O ID DESTE COMANDO — o
			// primeiro BeginMarker genérico pode ainda pertencer à resposta antiga.
			if r.desynced {
				if expectedBegin != "" && line == expectedBegin {
					r.desynced = false
					inResponse = true
					jsonLines = nil
				}
				continue
			}

			if line == expectedBegin || (expectedBegin != BeginMarker && line == BeginMarker) {
				inResponse = true
				jsonLines = nil
				continue
			}

			if line == EndMarker || strings.HasPrefix(line, "<<<PSADT_END:") {
				if !inResponse {
					continue
				}
				return []byte(strings.Join(jsonLines, "\n")), nil
			}

			if inResponse {
				jsonLines = append(jsonLines, line)
			}
		}
	}
}

// emitOutput sends a line to both the live output channel and the OnOutput callback.
func (r *Runner) emitOutput(line string) {
	if line == "" {
		return
	}
	if r.closing.Load() {
		return
	}
	select {
	case r.liveOutputCh <- line:
	default:
	}
	if r.onOutput != nil {
		r.onOutput(line)
	}
}
