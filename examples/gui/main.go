//go:build windows

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pedrostefanogv/go-psadt"
	"github.com/pedrostefanogv/go-psadt/types"
)

const (
	moduleName      = "PSAppDeployToolkit"
	defaultHTTPAddr = "127.0.0.1:17841"
)

type pageData struct {
	Message                      string
	Error                        string
	ClientReady                  bool
	ClientStatus                 string
	ClientReconnects             int
	LastClientInvalidationReason string
	LastClientInvalidationAt     string
	ToolkitInstalled             bool
	ToolkitVersion               string
	ToolkitInstallPath           string
	GeneratedConfig              string
	CurrentUIConfig              string
	LogLines                     []string
	Form                         formData
}

type formData struct {
	AppVendor      string
	AppName        string
	AppVersion     string
	DeploymentType string
	DeployMode     string

	DialogTitle   string
	DialogText    string
	DialogButtons string
	DialogIcon    string
	DialogTimeout int

	PromptTitle        string
	PromptMessage      string
	PromptAlignment    string
	PromptLeftButton   string
	PromptMiddleButton string
	PromptRightButton  string
	PromptIcon         string
	PromptTimeout      int
	PromptPos          string
	PromptAllowMove    bool
	PromptNotTopMost   bool

	BalloonTitle string
	BalloonText  string
	BalloonIcon  string
	BalloonTime  int

	ProgressMessage    string
	ProgressDetail     string
	ProgressPercent    string
	ProgressAlign      string
	ProgressPos        string
	ProgressAllowMove  bool
	ProgressNotTopMost bool

	WelcomeProcesses string
	WelcomeCountdown int
	WelcomeDiskCheck bool

	SelectedDialogStyle string
	SelectedTheme       string
	FluentAccentColor   string
	AssetAppIcon        string
	AssetAppIconDark    string
	AssetBannerClassic  string
}

type toolkitStatus struct {
	Installed bool
	Version   string
	Path      string
}

// ringBuffer armazena os últimos N logs em memória para exibição no painel.
type ringBuffer struct {
	mu  sync.Mutex
	buf []string
	cap int
}

func newRingBuffer(cap int) *ringBuffer { return &ringBuffer{cap: cap} }

func (rb *ringBuffer) append(s string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if len(rb.buf) >= rb.cap {
		rb.buf = rb.buf[1:]
	}
	rb.buf = append(rb.buf, s)
}

func (rb *ringBuffer) lines() []string {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	out := make([]string, len(rb.buf))
	copy(out, rb.buf)
	return out
}

// bufHandler é um slog.Handler que grava em um ringBuffer e também repassa para o handler pai.
type bufHandler struct {
	parent slog.Handler
	ring   *ringBuffer
}

func (h *bufHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.parent.Enabled(ctx, level)
}

func (h *bufHandler) Handle(ctx context.Context, rec slog.Record) error {
	var sb strings.Builder
	sb.WriteString(rec.Time.Format("15:04:05"))
	sb.WriteString(" ")
	sb.WriteString(rec.Level.String())
	sb.WriteString(" ")
	sb.WriteString(rec.Message)
	rec.Attrs(func(a slog.Attr) bool {
		sb.WriteString(" ")
		sb.WriteString(a.Key)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", a.Value))
		return true
	})
	h.ring.append(sb.String())
	return h.parent.Handle(ctx, rec)
}

func (h *bufHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &bufHandler{parent: h.parent.WithAttrs(attrs), ring: h.ring}
}

func (h *bufHandler) WithGroup(name string) slog.Handler {
	return &bufHandler{parent: h.parent.WithGroup(name), ring: h.ring}
}

// appState agrupa todo o estado compartilhado com proteção via mutex.
type appState struct {
	mu                  sync.RWMutex
	clientMu            sync.Mutex // protege criacao/invalidacao do client (warm-up seguro)
	sessionMu           sync.Mutex // serializa Open-ADTSession (uma sessao por vez no runner PS)
	clientConnectedOnce bool
	page                pageData
	startTime           time.Time
	reqCount            int64
	ring                *ringBuffer
	logger              *slog.Logger
	client              *psadt.Client // processo PS persistente, criado uma vez
}

var pageTmpl = template.Must(template.New("ui-lab").Parse(`<!doctype html>
<html lang="pt-br">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>go-psadt UI Lab</title>
  <style>
    :root {
      --bg: #f4f7f8;
      --bg-elev: #ffffff;
      --text: #1e2a2d;
      --muted: #5d6c70;
      --primary: #0b7a75;
      --primary-2: #2f9e9a;
      --danger: #b42318;
      --ok: #166534;
      --border: #c9d7db;
      --shadow: 0 10px 24px rgba(11, 32, 35, 0.09);
    }

    body.dark {
      --bg: #0f1719;
      --bg-elev: #162226;
      --text: #e6f1f3;
      --muted: #a7babf;
      --primary: #5cd6d0;
      --primary-2: #3fb7b2;
      --danger: #fda29b;
      --ok: #86efac;
      --border: #345158;
      --shadow: 0 10px 28px rgba(0, 0, 0, 0.35);
    }

    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Segoe UI", "Calibri", sans-serif;
      color: var(--text);
      background:
        radial-gradient(circle at 10% -10%, rgba(47, 158, 154, 0.22), transparent 45%),
        radial-gradient(circle at 100% 0%, rgba(11, 122, 117, 0.16), transparent 40%),
        var(--bg);
      min-height: 100vh;
    }

    .wrap { max-width: 1220px; margin: 0 auto; padding: 24px 18px 42px; }
    .header { display: flex; justify-content: space-between; align-items: center; gap: 16px; }
    h1 { margin: 0; font-size: 28px; }
    .sub { color: var(--muted); margin-top: 6px; }

    .top-actions { display: flex; gap: 10px; flex-wrap: wrap; }
    button, .btn {
      border: 1px solid var(--border);
      background: var(--bg-elev);
      color: var(--text);
      border-radius: 10px;
      padding: 9px 12px;
      cursor: pointer;
      font-weight: 600;
    }
    button.primary, .btn.primary {
      background: linear-gradient(130deg, var(--primary), var(--primary-2));
      border-color: transparent;
      color: #fff;
    }
    button:hover, .btn:hover { filter: brightness(1.03); }

    .notice {
      margin-top: 14px;
      border-radius: 10px;
      padding: 11px 13px;
      border: 1px solid var(--border);
      background: var(--bg-elev);
    }
    .notice.error { border-color: rgba(180, 35, 24, 0.4); color: var(--danger); }
    .notice.ok { border-color: rgba(22, 101, 52, 0.4); color: var(--ok); }

    .grid {
      margin-top: 18px;
      display: grid;
      grid-template-columns: 1.1fr 1fr;
      gap: 16px;
    }
    .card {
      background: var(--bg-elev);
      border: 1px solid var(--border);
      border-radius: 14px;
      padding: 14px;
      box-shadow: var(--shadow);
    }
    .card h2 { margin: 0 0 10px; font-size: 19px; }
    .card h3 { margin: 10px 0 8px; font-size: 15px; }

    .row {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
      margin-bottom: 10px;
    }
    .field { display: flex; flex-direction: column; gap: 4px; }
    label { font-size: 12px; color: var(--muted); }
    input, select, textarea {
      width: 100%;
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 8px 9px;
      background: var(--bg-elev);
      color: var(--text);
    }
    textarea { min-height: 72px; resize: vertical; }
    .full { grid-column: 1 / -1; }

    .toolkit-status { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; margin-bottom: 10px; }
    .pill { border: 1px solid var(--border); border-radius: 10px; padding: 8px; font-size: 13px; }
    .pill b { display: block; margin-bottom: 3px; }

    .section-actions { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 6px; }
    .mono { font-family: Consolas, "Courier New", monospace; white-space: pre-wrap; }

    @media (max-width: 990px) {
      .grid { grid-template-columns: 1fr; }
      .toolkit-status { grid-template-columns: 1fr; }
      .row { grid-template-columns: 1fr; }
    }

		.loading-overlay {
			display: none;
			position: fixed; inset: 0;
			background: rgba(0,0,0,0.45);
			z-index: 9999;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			color: #fff;
			font-size: 18px;
			gap: 16px;
		}
		.loading-overlay.active { display: flex; }
		.spinner {
			width: 44px; height: 44px;
			border: 5px solid rgba(255,255,255,0.3);
			border-top-color: #fff;
			border-radius: 50%;
			animation: spin 0.8s linear infinite;
		}
		@keyframes spin { to { transform: rotate(360deg); } }
		.client-badge {
			display: inline-block;
			border-radius: 8px;
			padding: 3px 8px;
			font-size: 12px;
			font-weight: 600;
		}
		.client-badge.init { background: rgba(180,120,24,0.14); color: #9a6700; border: 1px solid rgba(180,120,24,0.28); }
		.client-badge.ok  { background: rgba(22,101,52,0.15); color: var(--ok); border: 1px solid rgba(22,101,52,0.3); }
		.client-badge.off { background: rgba(180,35,24,0.12); color: var(--danger); border: 1px solid rgba(180,35,24,0.3); }
		.diag-grid {
			display: grid;
			grid-template-columns: repeat(3, minmax(0, 1fr));
			gap: 8px;
			margin: 10px 0 12px;
		}
		.diag-item {
			border: 1px solid var(--border);
			border-radius: 10px;
			padding: 8px;
			font-size: 13px;
		}
		.diag-item b {
			display: block;
			margin-bottom: 4px;
		}
		.diag-item .mono {
			font-size: 12px;
			white-space: normal;
			word-break: break-word;
		}
  </style>
  <script>
    function toggleTheme() {
      document.body.classList.toggle("dark");
      localStorage.setItem("ui-lab-theme", document.body.classList.contains("dark") ? "dark" : "light");
    }
    window.addEventListener("DOMContentLoaded", function() {
      const t = localStorage.getItem("ui-lab-theme");
      if (t === "dark") {
        document.body.classList.add("dark");
      }
			// Loading overlay: mostra ao enviar qualquer formulário com ação PSADT.
			var psadtForms = document.querySelectorAll('form[data-psadt]');
			psadtForms.forEach(function(f) {
				f.addEventListener('submit', function() {
					document.getElementById('loading-overlay').classList.add('active');
				});
			});
    });
  </script>
</head>
<body>
	<div id="loading-overlay" class="loading-overlay">
		<div class="spinner"></div>
		<div>Executando ação PSADT… aguarde</div>
		<div style="font-size:13px;opacity:0.75">A janela do toolkit pode aparecer atrás desta janela do navegador.</div>
	</div>
  <div class="wrap">
    <div class="header">
      <div>
        <h1>go-psadt UI Lab</h1>
        <div class="sub">Teste visual de modal, prompt, fluent/classic, alertas, assets e configuração de sessão.</div>
      </div>
      <div class="top-actions">
        <button type="button" onclick="toggleTheme()">Alternar claro/escuro (painel)</button>
      </div>
    </div>

    {{if .Error}}<div class="notice error">{{.Error}}</div>{{end}}
    {{if .Message}}<div class="notice ok">{{.Message}}</div>{{end}}

    <div class="grid">
      <div class="card">
        <h2>Toolkit (PSGallery)</h2>
        <div class="toolkit-status">
          <div class="pill"><b>Instalado</b>{{if .ToolkitInstalled}}Sim{{else}}Nao{{end}}</div>
          <div class="pill"><b>Versao</b>{{if .ToolkitVersion}}{{.ToolkitVersion}}{{else}}-{{end}}</div>
          <div class="pill"><b>Caminho</b>{{if .ToolkitInstallPath}}{{.ToolkitInstallPath}}{{else}}-{{end}}</div>
        </div>
				<div style="margin-bottom:8px">
			{{if .ClientReady}}<span class="client-badge ok">&#9679; Client PS pronto (modulo ja carregado)</span>{{else if eq .ClientStatus "initializing"}}<span class="client-badge init">&#9679; Client PS inicializando...</span>{{else if eq .ClientStatus "error"}}<span class="client-badge off">&#9679; Client PS indisponivel</span>{{else}}<span class="client-badge off">&#9679; Client PS nao iniciado</span>{{end}}
				</div>
				<div class="diag-grid">
				  <div class="diag-item"><b>Reconexoes</b>{{.ClientReconnects}}</div>
				  <div class="diag-item"><b>Ultima invalidacao</b>{{if .LastClientInvalidationAt}}{{.LastClientInvalidationAt}}{{else}}-{{end}}</div>
				  <div class="diag-item"><b>Motivo</b><span class="mono">{{if .LastClientInvalidationReason}}{{.LastClientInvalidationReason}}{{else}}-{{end}}</span></div>
				</div>
        <div class="section-actions">
          <form method="post" action="/toolkit/check"><button class="primary" type="submit">Verificar toolkit</button></form>
          <form method="post" action="/toolkit/install/user"><button type="submit">Instalar CurrentUser</button></form>
          <form method="post" action="/toolkit/install/allusers"><button type="submit">Instalar AllUsers</button></form>
        </div>
        <p class="sub">A instalacao usa Install-Module via PowerShell Gallery e tenta adicionar PSGallery como Trusted para evitar prompts.</p>

        <h3>Sessao base</h3>
				<form method="post" action="/run/welcome" data-psadt>
          <div class="row">
            <div class="field"><label>Vendor</label><input name="AppVendor" value="{{.Form.AppVendor}}"></div>
            <div class="field"><label>App</label><input name="AppName" value="{{.Form.AppName}}"></div>
            <div class="field"><label>Versao</label><input name="AppVersion" value="{{.Form.AppVersion}}"></div>
            <div class="field"><label>DeploymentType</label>
              <select name="DeploymentType">
                <option {{if eq .Form.DeploymentType "Install"}}selected{{end}}>Install</option>
                <option {{if eq .Form.DeploymentType "Uninstall"}}selected{{end}}>Uninstall</option>
                <option {{if eq .Form.DeploymentType "Repair"}}selected{{end}}>Repair</option>
              </select>
            </div>
            <div class="field"><label>DeployMode</label>
              <select name="DeployMode">
                <option {{if eq .Form.DeployMode "Interactive"}}selected{{end}}>Interactive</option>
                <option {{if eq .Form.DeployMode "Silent"}}selected{{end}}>Silent</option>
                <option {{if eq .Form.DeployMode "NonInteractive"}}selected{{end}}>NonInteractive</option>
                <option {{if eq .Form.DeployMode "Auto"}}selected{{end}}>Auto</option>
              </select>
            </div>
            <div class="field"><label>Processos para fechar (CSV)</label><input name="WelcomeProcesses" value="{{.Form.WelcomeProcesses}}"></div>
            <div class="field"><label>Countdown (s)</label><input type="number" name="WelcomeCountdown" value="{{.Form.WelcomeCountdown}}"></div>
            <div class="field"><label>CheckDiskSpace</label>
              <select name="WelcomeDiskCheck">
                <option value="true" {{if .Form.WelcomeDiskCheck}}selected{{end}}>true</option>
                <option value="false" {{if not .Form.WelcomeDiskCheck}}selected{{end}}>false</option>
              </select>
            </div>
          </div>
          <button class="primary" type="submit">Testar ShowInstallationWelcome</button>
        </form>
      </div>

      <div class="card">
        <h2>Dialogos e Alertas</h2>
		<form method="post" action="/run/dialog" data-psadt>
          <div class="row">
            <div class="field"><label>Titulo modal</label><input name="DialogTitle" value="{{.Form.DialogTitle}}"></div>
            <div class="field"><label>Icone modal</label>
              <select name="DialogIcon">
				<option {{if eq .Form.DialogIcon "None"}}selected{{end}}>None</option>
                <option {{if eq .Form.DialogIcon "Question"}}selected{{end}}>Question</option>
                <option {{if eq .Form.DialogIcon "Information"}}selected{{end}}>Information</option>
				<option {{if eq .Form.DialogIcon "Exclamation"}}selected{{end}}>Exclamation</option>
				<option {{if eq .Form.DialogIcon "Stop"}}selected{{end}}>Stop</option>
              </select>
            </div>
            <div class="field full"><label>Texto modal</label><textarea name="DialogText">{{.Form.DialogText}}</textarea></div>
            <div class="field"><label>Botoes</label>
              <select name="DialogButtons">
                <option {{if eq .Form.DialogButtons "YesNo"}}selected{{end}}>YesNo</option>
                <option {{if eq .Form.DialogButtons "Ok"}}selected{{end}}>Ok</option>
                <option {{if eq .Form.DialogButtons "OkCancel"}}selected{{end}}>OkCancel</option>
                <option {{if eq .Form.DialogButtons "YesNoCancel"}}selected{{end}}>YesNoCancel</option>
                <option {{if eq .Form.DialogButtons "RetryCancel"}}selected{{end}}>RetryCancel</option>
                <option {{if eq .Form.DialogButtons "AbortRetryIgnore"}}selected{{end}}>AbortRetryIgnore</option>
              </select>
            </div>
            <div class="field"><label>Timeout (s)</label><input type="number" name="DialogTimeout" value="{{.Form.DialogTimeout}}"></div>
          </div>
          <button class="primary" type="submit">Testar modal (ShowDialogBox)</button>
        </form>

        <h3>Prompt (modal PSADT Fluent/Classic)</h3>
		<form method="post" action="/run/prompt" data-psadt>
          <div class="row">
            <div class="field"><label>Titulo</label><input name="PromptTitle" value="{{.Form.PromptTitle}}"></div>
            <div class="field"><label>Icone</label>
              <select name="PromptIcon">
                <option {{if eq .Form.PromptIcon "Question"}}selected{{end}}>Question</option>
                <option {{if eq .Form.PromptIcon "Information"}}selected{{end}}>Information</option>
                <option {{if eq .Form.PromptIcon "Warning"}}selected{{end}}>Warning</option>
                <option {{if eq .Form.PromptIcon "Error"}}selected{{end}}>Error</option>
                <option {{if eq .Form.PromptIcon "Shield"}}selected{{end}}>Shield</option>
              </select>
            </div>
            <div class="field full"><label>Mensagem</label><textarea name="PromptMessage">{{.Form.PromptMessage}}</textarea></div>
            <div class="field"><label>Alinhamento</label>
              <select name="PromptAlignment">
                <option {{if eq .Form.PromptAlignment "Left"}}selected{{end}}>Left</option>
                <option {{if eq .Form.PromptAlignment "Center"}}selected{{end}}>Center</option>
                <option {{if eq .Form.PromptAlignment "Right"}}selected{{end}}>Right</option>
              </select>
            </div>
            <div class="field"><label>Timeout (s)</label><input type="number" name="PromptTimeout" value="{{.Form.PromptTimeout}}"></div>
            <div class="field"><label>Botao esquerdo</label><input name="PromptLeftButton" value="{{.Form.PromptLeftButton}}"></div>
            <div class="field"><label>Botao meio</label><input name="PromptMiddleButton" value="{{.Form.PromptMiddleButton}}"></div>
            <div class="field"><label>Botao direito</label><input name="PromptRightButton" value="{{.Form.PromptRightButton}}"></div>
			<div class="field"><label>Posicao da janela</label>
			  <select name="PromptPos">
				<option value="" {{if eq .Form.PromptPos ""}}selected{{end}}>Padrao do toolkit</option>
				<option {{if eq .Form.PromptPos "TopLeft"}}selected{{end}}>TopLeft</option>
				<option {{if eq .Form.PromptPos "Top"}}selected{{end}}>Top</option>
				<option {{if eq .Form.PromptPos "TopRight"}}selected{{end}}>TopRight</option>
				<option {{if eq .Form.PromptPos "TopCenter"}}selected{{end}}>TopCenter</option>
				<option {{if eq .Form.PromptPos "Center"}}selected{{end}}>Center</option>
				<option {{if eq .Form.PromptPos "BottomLeft"}}selected{{end}}>BottomLeft</option>
				<option {{if eq .Form.PromptPos "Bottom"}}selected{{end}}>Bottom</option>
				<option {{if eq .Form.PromptPos "BottomRight"}}selected{{end}}>BottomRight</option>
			  </select>
			</div>
			<div class="field"><label>Permitir mover</label>
			  <select name="PromptAllowMove">
				<option value="false" {{if not .Form.PromptAllowMove}}selected{{end}}>false</option>
				<option value="true" {{if .Form.PromptAllowMove}}selected{{end}}>true</option>
			  </select>
			</div>
			<div class="field"><label>Nao ficar sempre no topo</label>
			  <select name="PromptNotTopMost">
				<option value="false" {{if not .Form.PromptNotTopMost}}selected{{end}}>false</option>
				<option value="true" {{if .Form.PromptNotTopMost}}selected{{end}}>true</option>
			  </select>
			</div>
          </div>
          <button class="primary" type="submit">Testar prompt (ShowInstallationPrompt)</button>
        </form>
		<p class="sub">No estilo Fluent do PSADT, Icon e MessageAlignment podem nao ter efeito. Para comportamento de janela, teste Posicao, AllowMove e NotTopMost.</p>

        <h3>Balloon alert</h3>
		<form method="post" action="/run/balloon" data-psadt>
          <div class="row">
            <div class="field"><label>Titulo</label><input name="BalloonTitle" value="{{.Form.BalloonTitle}}"></div>
            <div class="field"><label>Icone</label>
              <select name="BalloonIcon">
                <option {{if eq .Form.BalloonIcon "Info"}}selected{{end}}>Info</option>
                <option {{if eq .Form.BalloonIcon "Warning"}}selected{{end}}>Warning</option>
                <option {{if eq .Form.BalloonIcon "Error"}}selected{{end}}>Error</option>
                <option {{if eq .Form.BalloonIcon "None"}}selected{{end}}>None</option>
              </select>
            </div>
            <div class="field full"><label>Texto</label><textarea name="BalloonText">{{.Form.BalloonText}}</textarea></div>
            <div class="field"><label>Tempo (ms)</label><input type="number" name="BalloonTime" value="{{.Form.BalloonTime}}"></div>
          </div>
          <button class="primary" type="submit">Testar balloon (ShowBalloonTip)</button>
        </form>
      </div>
    </div>

    <div class="grid">
      <div class="card">
        <h2>Progress + Config visual</h2>
		<form method="post" action="/run/progress" data-psadt>
          <div class="row">
            <div class="field"><label>Status</label><input name="ProgressMessage" value="{{.Form.ProgressMessage}}"></div>
            <div class="field"><label>Detalhe</label><input name="ProgressDetail" value="{{.Form.ProgressDetail}}"></div>
			<div class="field"><label>Percentual (opcional)</label><input type="number" step="0.1" name="ProgressPercent" value="{{.Form.ProgressPercent}}" placeholder="vazio = indeterminado"></div>
			<div class="field"><label>Posicao da janela</label>
			  <select name="ProgressPos">
				<option value="" {{if eq .Form.ProgressPos ""}}selected{{end}}>Padrao do toolkit</option>
				<option {{if eq .Form.ProgressPos "TopLeft"}}selected{{end}}>TopLeft</option>
				<option {{if eq .Form.ProgressPos "Top"}}selected{{end}}>Top</option>
				<option {{if eq .Form.ProgressPos "TopRight"}}selected{{end}}>TopRight</option>
				<option {{if eq .Form.ProgressPos "TopCenter"}}selected{{end}}>TopCenter</option>
				<option {{if eq .Form.ProgressPos "Center"}}selected{{end}}>Center</option>
				<option {{if eq .Form.ProgressPos "BottomLeft"}}selected{{end}}>BottomLeft</option>
				<option {{if eq .Form.ProgressPos "Bottom"}}selected{{end}}>Bottom</option>
				<option {{if eq .Form.ProgressPos "BottomRight"}}selected{{end}}>BottomRight</option>
			  </select>
			</div>
			<div class="field"><label>Alinhamento (Classic)</label>
			  <select name="ProgressAlign">
				<option value="" {{if eq .Form.ProgressAlign ""}}selected{{end}}>Padrao do toolkit</option>
				<option {{if eq .Form.ProgressAlign "Left"}}selected{{end}}>Left</option>
				<option {{if eq .Form.ProgressAlign "Center"}}selected{{end}}>Center</option>
				<option {{if eq .Form.ProgressAlign "Right"}}selected{{end}}>Right</option>
			  </select>
			</div>
			<div class="field"><label>Permitir mover</label>
			  <select name="ProgressAllowMove">
				<option value="false" {{if not .Form.ProgressAllowMove}}selected{{end}}>false</option>
				<option value="true" {{if .Form.ProgressAllowMove}}selected{{end}}>true</option>
			  </select>
			</div>
			<div class="field"><label>Nao ficar sempre no topo</label>
			  <select name="ProgressNotTopMost">
				<option value="false" {{if not .Form.ProgressNotTopMost}}selected{{end}}>false</option>
				<option value="true" {{if .Form.ProgressNotTopMost}}selected{{end}}>true</option>
			  </select>
			</div>
          </div>
          <button class="primary" type="submit">Exibir progress (6s)</button>
        </form>
		<p class="sub">No PSADT atual, o progress abre uma janela WPF separada. Isso pode aparecer como uma janela propria na barra de tarefas do Windows; a opcao NotTopMost apenas remove o always-on-top, nao transforma a UI em overlay embutido.</p>

        <h3>Assets e estilo (config.psd1)</h3>
        <form method="post" action="/config/generate">
          <div class="row">
            <div class="field"><label>DialogStyle</label>
              <select name="SelectedDialogStyle">
                <option {{if eq .Form.SelectedDialogStyle "Fluent"}}selected{{end}}>Fluent</option>
                <option {{if eq .Form.SelectedDialogStyle "Classic"}}selected{{end}}>Classic</option>
              </select>
            </div>
            <div class="field"><label>Theme sugerido</label>
              <select name="SelectedTheme">
                <option {{if eq .Form.SelectedTheme "Light"}}selected{{end}}>Light</option>
                <option {{if eq .Form.SelectedTheme "Dark"}}selected{{end}}>Dark</option>
              </select>
            </div>
            <div class="field"><label>Fluent accent color (hex)</label><input name="FluentAccentColor" value="{{.Form.FluentAccentColor}}" placeholder="#0B7A75"></div>
            <div class="field"><label>AppIcon (png)</label><input name="AssetAppIcon" value="{{.Form.AssetAppIcon}}" placeholder="C:\\Deploy\\Assets\\AppIcon.png"></div>
            <div class="field"><label>AppIconDark (png)</label><input name="AssetAppIconDark" value="{{.Form.AssetAppIconDark}}" placeholder="C:\\Deploy\\Assets\\AppIcon-Dark.png"></div>
            <div class="field"><label>BannerClassic (png)</label><input name="AssetBannerClassic" value="{{.Form.AssetBannerClassic}}" placeholder="C:\\Deploy\\Assets\\Banner.Classic.png"></div>
          </div>
          <button class="primary" type="submit">Gerar snippet de config</button>
        </form>

        {{if .GeneratedConfig}}
        <h3>Snippet gerado</h3>
        <div class="pill mono">{{.GeneratedConfig}}</div>
        {{end}}
      </div>

      <div class="card">
        <h2>Diagnostico</h2>
        <div class="section-actions">
		  <form method="post" action="/config/read" data-psadt><button class="primary" type="submit">Ler GetConfig atual</button></form>
		  <form method="post" action="/query/environment" data-psadt><button type="submit">Rodar query de ambiente</button></form>
        </div>
        {{if .CurrentUIConfig}}
        <h3>Config atual (Get-ADTConfig)</h3>
        <div class="pill mono">{{.CurrentUIConfig}}</div>
        {{end}}
				<p class="sub">O endpoint /diag tambem expone client_reconnect_count, last_client_invalidation_reason e last_client_invalidation_at para acompanhar quando o runner e recriado em background.</p>
        <p class="sub">O estilo Fluent/Classic, assets e cores dependem da configuracao ativa do toolkit. O app gera snippet para facilitar teste rapido em laboratorio.</p>
      </div>
    </div>

    <div class="card" style="margin-top:16px">
      <h2>Log de diagnóstico <span style="font-size:13px;font-weight:400;color:var(--muted)">(últimas 100 entradas · <a href="/diag" target="_blank" style="color:var(--primary)">/diag JSON</a>)</span></h2>
      <div id="logpanel" style="font-family:Consolas,'Courier New',monospace;font-size:12px;line-height:1.6;max-height:260px;overflow-y:auto;background:var(--bg);border:1px solid var(--border);border-radius:8px;padding:10px;white-space:pre-wrap;">
        {{if .LogLines}}{{range .LogLines}}{{.}}&#10;{{end}}{{else}}<span style="color:var(--muted)">Nenhum log ainda. Execute uma ação para ver os logs aqui.</span>{{end}}
      </div>
      <div style="margin-top:8px;display:flex;gap:8px;align-items:center">
        <form method="get" action="/"><button type="submit">Atualizar logs</button></form>
        <span style="font-size:12px;color:var(--muted)">Ou acesse <a href="/diag" target="_blank" style="color:var(--primary)">/diag</a> para JSON completo com uptime, versões e último erro.</span>
      </div>
    </div>
  </div>
</body>
</html>`))

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	ring := newRingBuffer(100)
	textHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(&bufHandler{parent: textHandler, ring: ring})
	slog.SetDefault(logger)

	status, err := checkToolkitInstalled(context.Background())
	if err != nil {
		logger.Warn("verificacao inicial do toolkit falhou", "err", err)
	}

	app := &appState{
		startTime: time.Now(),
		ring:      ring,
		logger:    logger,
		page: pageData{
			ClientStatus:       "initializing",
			ToolkitInstalled:   status.Installed,
			ToolkitVersion:     status.Version,
			ToolkitInstallPath: status.Path,
			Form:               defaultFormData(),
		},
	}

	// Warm-up em background: nao bloqueia o startup da interface HTTP.
	// Usa apenas clientMu internamente — nunca bloqueia sessionMu.
	logger.Info("agendando inicializacao do client PSADT persistente em background...")
	go func() {
		if err := app.ensureClient(); err != nil {
			logger.Warn("warm-up do client PSADT falhou (sera tentado sob demanda)", "err", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		app.mu.RLock()
		snap := app.page
		app.mu.RUnlock()
		snap.LogLines = ring.lines()
		renderPage(w, &snap)
	})))

	mux.Handle("/toolkit/check", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		logger.Info("verificando toolkit instalado")
		status, err := checkToolkitInstalled(r.Context())
		app.mu.Lock()
		if err != nil {
			logger.Error("falha ao verificar toolkit", "err", err)
			app.page.Error = fmt.Sprintf("Falha ao verificar toolkit: %v", err)
			app.page.Message = ""
		} else {
			app.page.ToolkitInstalled = status.Installed
			app.page.ToolkitVersion = status.Version
			app.page.ToolkitInstallPath = status.Path
			app.page.Error = ""
			if status.Installed {
				app.page.Message = fmt.Sprintf("Toolkit encontrado: versao %s", status.Version)
				logger.Info("toolkit encontrado", "version", status.Version, "path", status.Path)
			} else {
				app.page.Message = "Toolkit nao encontrado. Use os botoes de instalacao via PSGallery."
				logger.Warn("toolkit nao encontrado")
			}
		}
		app.mu.Unlock()
		redirectHome(w, r)
	})))

	mux.Handle("/toolkit/install/user", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInstallToolkit(w, r, app, "CurrentUser")
	})))
	mux.Handle("/toolkit/install/allusers", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInstallToolkit(w, r, app, "AllUsers")
	})))

	mux.Handle("/run/welcome", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, app, func(s *psadt.Session, f formData) (string, error) {
			procs := parseProcesses(f.WelcomeProcesses)
			err := s.ShowInstallationWelcome(types.WelcomeOptions{
				CloseProcesses:          procs,
				CloseProcessesCountdown: f.WelcomeCountdown,
				CheckDiskSpace:          f.WelcomeDiskCheck,
			})
			if err != nil {
				return "", err
			}
			return "ShowInstallationWelcome executado com sucesso.", nil
		})
	})))

	mux.Handle("/run/dialog", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, app, func(s *psadt.Session, f formData) (string, error) {
			result, err := s.ShowDialogBox(types.DialogBoxOptions{
				Title:   f.DialogTitle,
				Text:    f.DialogText,
				Buttons: types.DialogBoxButtons(f.DialogButtons),
				Icon:    types.DialogSystemIcon(f.DialogIcon),
				Timeout: f.DialogTimeout,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("ShowDialogBox concluido. Botao selecionado: %s", result), nil
		})
	})))

	mux.Handle("/run/prompt", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, app, func(s *psadt.Session, f formData) (string, error) {
			res, err := s.ShowInstallationPrompt(types.PromptOptions{
				Title:            f.PromptTitle,
				Message:          f.PromptMessage,
				MessageAlignment: types.MessageAlignment(f.PromptAlignment),
				ButtonLeftText:   f.PromptLeftButton,
				ButtonMiddleText: f.PromptMiddleButton,
				ButtonRightText:  f.PromptRightButton,
				Icon:             types.DialogSystemIcon(f.PromptIcon),
				Timeout:          f.PromptTimeout,
				WindowLocation:   types.DialogPosition(f.PromptPos),
				AllowMove:        f.PromptAllowMove,
				NotTopMost:       f.PromptNotTopMost,
			})
			if err != nil {
				return "", err
			}
			if res == nil {
				return "ShowInstallationPrompt executado sem retorno de botao.", nil
			}
			return fmt.Sprintf("ShowInstallationPrompt concluido. ButtonClicked=%s", res.ButtonClicked), nil
		})
	})))

	mux.Handle("/run/balloon", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, app, func(s *psadt.Session, f formData) (string, error) {
			err := s.ShowBalloonTip(types.BalloonTipOptions{
				BalloonTipTitle: f.BalloonTitle,
				BalloonTipText:  f.BalloonText,
				BalloonTipIcon:  types.BalloonTipIcon(f.BalloonIcon),
				BalloonTipTime:  f.BalloonTime,
			})
			if err != nil {
				return "", err
			}
			return "ShowBalloonTip enviado com sucesso.", nil
		})
	})))

	mux.Handle("/run/progress", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, app, func(s *psadt.Session, f formData) (string, error) {
			err := s.ShowInstallationProgress(types.ProgressOptions{
				StatusMessage:       f.ProgressMessage,
				StatusMessageDetail: f.ProgressDetail,
				StatusBarPercentage: parseFloatOrDefault(f.ProgressPercent, 0),
				MessageAlignment:    types.MessageAlignment(f.ProgressAlign),
				WindowLocation:      types.DialogPosition(f.ProgressPos),
				AllowMove:           f.ProgressAllowMove,
				NotTopMost:          f.ProgressNotTopMost,
			})
			if err != nil {
				return "", err
			}
			time.Sleep(6 * time.Second)
			if cerr := s.CloseInstallationProgress(); cerr != nil {
				return "", cerr
			}
			return "ShowInstallationProgress exibido por 6 segundos e fechado.", nil
		})
	})))

	mux.Handle("/config/generate", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = r.ParseForm()
		app.mu.Lock()
		app.page.Form = mergeForm(app.page.Form, r)
		app.page.GeneratedConfig = buildToolkitConfigSnippet(app.page.Form)
		app.page.Error = ""
		app.page.Message = "Snippet de configuracao gerado."
		app.mu.Unlock()
		logger.Info("snippet de config gerado", "style", r.FormValue("SelectedDialogStyle"))
		redirectHome(w, r)
	})))

	mux.Handle("/config/read", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, app, func(s *psadt.Session, _ formData) (string, error) {
			cfg, err := s.GetConfig()
			if err != nil {
				return "", err
			}
			encoded, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return "", err
			}
			app.mu.Lock()
			app.page.CurrentUIConfig = string(encoded)
			app.mu.Unlock()
			return "GetConfig carregado para validar estilo/modal/cores atuais do toolkit.", nil
		})
	})))

	mux.Handle("/query/environment", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		logger.Info("query de ambiente via client persistente")
		if err := app.ensureClient(); err != nil {
			logger.Error("client PSADT indisponivel para query de ambiente", "err", err)
			app.mu.Lock()
			app.page.Error = fmt.Sprintf("Client PSADT indisponivel: %v", err)
			app.page.Message = ""
			app.mu.Unlock()
			redirectHome(w, r)
			return
		}

		// sessionMu apos ensureClient para evitar deadlock com warm-up.
		app.sessionMu.Lock()
		defer app.sessionMu.Unlock()

		// Re-verifica dentro do lock (client pode ter sido invalidado no intervalo).
		if err := app.ensureClient(); err != nil {
			logger.Error("client PSADT indisponivel para query de ambiente (re-check)", "err", err)
			app.mu.Lock()
			app.page.Error = fmt.Sprintf("Client PSADT indisponivel: %v", err)
			app.page.Message = ""
			app.mu.Unlock()
			redirectHome(w, r)
			return
		}

		logger.Info("chamando GetEnvironment")
		env, err := app.client.GetEnvironment()
		if err != nil {
			logger.Error("GetEnvironment falhou", "err", err)
			app.invalidateClient()
			app.mu.Lock()
			app.page.Error = fmt.Sprintf("GetEnvironment falhou: %v", err)
			app.page.Message = ""
			app.mu.Unlock()
			redirectHome(w, r)
			return
		}

		result := fmt.Sprintf(
			"OS=%s %s\nPowerShell=%s\nIsAdmin=%v\nToolkit=%s %s",
			env.OS.Name,
			env.OS.Version,
			env.PowerShell.PSVersion,
			env.Permissions.IsAdmin,
			env.Toolkit.ShortName,
			env.Toolkit.Version,
		)
		logger.Info("GetEnvironment concluido", "os", env.OS.Name, "psVersion", env.PowerShell.PSVersion, "isAdmin", env.Permissions.IsAdmin)
		app.mu.Lock()
		app.page.CurrentUIConfig = result
		app.page.Error = ""
		app.page.Message = "Consulta de ambiente concluida."
		app.mu.Unlock()
		redirectHome(w, r)
	})))

	// /diag — endpoint JSON com estado completo para diagnóstico externo.
	mux.Handle("/diag", logMiddleware(logger, app, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		app.mu.RLock()
		snap := app.page
		reqCount := app.reqCount
		uptime := time.Since(app.startTime).Round(time.Second).String()
		app.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uptime":                          uptime,
			"request_count":                   reqCount,
			"go_version":                      runtime.Version(),
			"os":                              runtime.GOOS,
			"arch":                            runtime.GOARCH,
			"client_ready":                    snap.ClientReady,
			"client_status":                   snap.ClientStatus,
			"client_reconnect_count":          snap.ClientReconnects,
			"last_client_invalidation_reason": snap.LastClientInvalidationReason,
			"last_client_invalidation_at":     snap.LastClientInvalidationAt,
			"toolkit_installed":               snap.ToolkitInstalled,
			"toolkit_version":                 snap.ToolkitVersion,
			"toolkit_path":                    snap.ToolkitInstallPath,
			"last_error":                      snap.Error,
			"last_message":                    snap.Message,
			"log_lines":                       ring.lines(),
		})
	})))

	logger.Info("go-psadt UI Lab iniciado", "addr", "http://"+defaultHTTPAddr, "diag", "http://"+defaultHTTPAddr+"/diag")
	log.Printf("go-psadt UI Lab em http://%s  |  diagnostico: http://%s/diag", defaultHTTPAddr, defaultHTTPAddr)
	if err := http.ListenAndServe(defaultHTTPAddr, mux); err != nil {
		log.Fatal(err)
	}
}

// logMiddleware loga método, path, duração e incrementa o contador de requisições.
func logMiddleware(logger *slog.Logger, app *appState, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		app.mu.Lock()
		app.reqCount++
		app.mu.Unlock()
		logger.Info("req", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
		logger.Debug("req done", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}

func runWithSession(
	w http.ResponseWriter,
	r *http.Request,
	app *appState,
	action func(s *psadt.Session, f formData) (string, error),
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	app.mu.Lock()
	f := mergeForm(app.page.Form, r)
	app.page.Form = f
	app.mu.Unlock()

	// Garante que o client persistente existe; reconecta se necessario.
	// ensureClient usa clientMu internamente — nao bloqueia sessionMu.
	if err := app.ensureClient(); err != nil {
		app.logger.Error("client PSADT indisponivel", "path", r.URL.Path, "err", err)
		app.mu.Lock()
		app.page.Error = fmt.Sprintf("Client PSADT indisponivel. Toolkit instalado? Erro: %v", err)
		app.page.Message = ""
		app.mu.Unlock()
		redirectHome(w, r)
		return
	}

	// sessionMu serializa o Open-ADTSession (runner PS nao suporta sessoes paralelas).
	// ensureClient() foi chamado antes para aguardar warm-up sem bloquear sessionMu.
	app.sessionMu.Lock()
	defer app.sessionMu.Unlock()

	// Re-verifica o client dentro do lock: pode ter sido invalidado entre a chamada
	// acima e a aquisicao do sessionMu (ex: reinstalacao de toolkit concorrente).
	if err := app.ensureClient(); err != nil {
		app.logger.Error("client PSADT indisponivel (re-check)", "path", r.URL.Path, "err", err)
		app.mu.Lock()
		app.page.Error = fmt.Sprintf("Client PSADT indisponivel. Toolkit instalado? Erro: %v", err)
		app.page.Message = ""
		app.mu.Unlock()
		redirectHome(w, r)
		return
	}

	app.logger.Info("abrindo sessao ADT", "path", r.URL.Path, "app", f.AppName, "type", f.DeploymentType, "mode", f.DeployMode)
	session, err := app.client.OpenSession(types.SessionConfig{
		DeploymentType: types.DeploymentType(f.DeploymentType),
		DeployMode:     types.DeployMode(f.DeployMode),
		AppVendor:      f.AppVendor,
		AppName:        f.AppName,
		AppVersion:     f.AppVersion,
	})
	if err != nil {
		app.logger.Error("falha ao abrir sessao ADT", "err", err)
		// Sessão falhou: invalida o client para forçar reconexão na próxima tentativa.
		app.invalidateClient()
		app.mu.Lock()
		app.page.Error = fmt.Sprintf("Falha ao abrir sessao ADT: %v", err)
		app.page.Message = ""
		app.mu.Unlock()
		redirectHome(w, r)
		return
	}
	defer func() {
		needReconnect := false
		if cerr := session.Close(0); cerr != nil {
			app.logger.Warn("falha ao fechar sessao ADT", "path", r.URL.Path, "err", cerr)
			app.invalidateClient()
			needReconnect = true
		} else if app.client != nil && !app.client.IsAlive() {
			app.invalidateClientWithLevel(slog.LevelInfo, "Close-ADTSession encerrou o runner; o client sera reconectado em background")
			needReconnect = true
		}
		if needReconnect {
			go func() {
				if werr := app.ensureClient(); werr != nil {
					app.logger.Warn("reconexao do client apos fechamento falhou", "path", r.URL.Path, "err", werr)
				}
			}()
		}
	}()

	app.logger.Info("executando acao", "path", r.URL.Path)
	msg, err := action(session, f)
	if err != nil {
		app.logger.Error("acao falhou", "path", r.URL.Path, "err", err)
		app.mu.Lock()
		app.page.Error = fmt.Sprintf("Acao falhou: %v", err)
		app.page.Message = ""
		app.mu.Unlock()
		redirectHome(w, r)
		return
	}

	app.logger.Info("acao concluida", "path", r.URL.Path, "msg", msg)
	status, _ := checkToolkitInstalled(r.Context())
	app.mu.Lock()
	app.page.ToolkitInstalled = status.Installed
	app.page.ToolkitVersion = status.Version
	app.page.ToolkitInstallPath = status.Path
	app.page.Error = ""
	app.page.Message = msg
	app.mu.Unlock()
	redirectHome(w, r)
}

// ensureClient garante que app.client esta pronto.
// Seguro para chamadas concorrentes — usa clientMu internamente.
func (app *appState) ensureClient() error {
	app.clientMu.Lock()
	defer app.clientMu.Unlock()
	if app.client != nil {
		if app.client.IsAlive() {
			return nil
		}
		app.invalidateClientLocked(slog.LevelInfo, "client PSADT existente nao esta mais ativo; recriando runner")
	}
	app.mu.Lock()
	app.page.ClientReady = false
	app.page.ClientStatus = "initializing"
	app.mu.Unlock()
	app.logger.Info("reconectando client PSADT (importando modulo PS)...")
	c, err := psadt.NewClient(
		psadt.WithTimeout(3*time.Minute),
		psadt.WithLogger(app.logger),
	)
	if err != nil {
		app.mu.Lock()
		app.page.ClientReady = false
		app.page.ClientStatus = "error"
		app.mu.Unlock()
		return err
	}
	isReconnect := app.clientConnectedOnce
	app.clientConnectedOnce = true
	app.client = c
	app.recordClientConnected(isReconnect)
	app.logger.Info("client PSADT reconectado com sucesso")
	return nil
}

// invalidateClient descarta o client atual para forcar reconexao.
// Seguro para chamadas concorrentes — usa clientMu internamente.
func (app *appState) invalidateClient() {
	app.invalidateClientWithLevel(slog.LevelWarn, "client PSADT invalidado — sera reconectado na proxima acao")
}

func (app *appState) invalidateClientWithLevel(level slog.Level, msg string) {
	app.clientMu.Lock()
	defer app.clientMu.Unlock()
	app.invalidateClientLocked(level, msg)
}

func (app *appState) invalidateClientLocked(level slog.Level, msg string) {
	if app.client != nil {
		_ = app.client.Close()
		app.client = nil
	}
	app.recordClientInvalidation(msg)
	app.logger.Log(context.Background(), level, msg)
}

func (app *appState) recordClientConnected(isReconnect bool) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.page.ClientReady = true
	app.page.ClientStatus = "ready"
	if isReconnect {
		app.page.ClientReconnects++
	}
}

func (app *appState) recordClientInvalidation(reason string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.page.ClientReady = false
	app.page.ClientStatus = "error"
	app.page.LastClientInvalidationReason = reason
	app.page.LastClientInvalidationAt = time.Now().Format(time.RFC3339)
}

func defaultFormData() formData {
	return formData{
		AppVendor:           "Contoso",
		AppName:             "UI Lab",
		AppVersion:          "1.0.0",
		DeploymentType:      string(types.DeployInstall),
		DeployMode:          string(types.DeployModeInteractive),
		DialogTitle:         "Teste modal",
		DialogText:          "Deseja continuar com o teste do modal?",
		DialogButtons:       string(types.ButtonsYesNo),
		DialogIcon:          string(types.IconQuestion),
		DialogTimeout:       45,
		PromptTitle:         "Teste prompt",
		PromptMessage:       "Clique em uma opcao para validar o comportamento do prompt.",
		PromptAlignment:     string(types.AlignCenter),
		PromptLeftButton:    "Prosseguir",
		PromptMiddleButton:  "Adiar",
		PromptRightButton:   "Cancelar",
		PromptIcon:          string(types.IconShield),
		PromptTimeout:       60,
		PromptPos:           "",
		BalloonTitle:        "go-psadt UI Lab",
		BalloonText:         "Notificacao de teste enviada com sucesso.",
		BalloonIcon:         string(types.BalloonInfo),
		BalloonTime:         8000,
		ProgressMessage:     "Executando teste de progresso...",
		ProgressDetail:      "Aguarde alguns segundos para fechamento automatico.",
		ProgressAlign:       "",
		ProgressPos:         "",
		WelcomeProcesses:    "msiexec,setup,installer",
		WelcomeCountdown:    120,
		WelcomeDiskCheck:    true,
		SelectedDialogStyle: string(types.DialogStyleFluent),
		SelectedTheme:       "Light",
		FluentAccentColor:   "#0B7A75",
	}
}

func mergeForm(base formData, r *http.Request) formData {
	out := base
	out.AppVendor = formOrDefault(r, "AppVendor", out.AppVendor)
	out.AppName = formOrDefault(r, "AppName", out.AppName)
	out.AppVersion = formOrDefault(r, "AppVersion", out.AppVersion)
	out.DeploymentType = formOrDefault(r, "DeploymentType", out.DeploymentType)
	out.DeployMode = formOrDefault(r, "DeployMode", out.DeployMode)

	out.DialogTitle = formOrDefault(r, "DialogTitle", out.DialogTitle)
	out.DialogText = formOrDefault(r, "DialogText", out.DialogText)
	out.DialogButtons = formOrDefault(r, "DialogButtons", out.DialogButtons)
	out.DialogIcon = formOrDefault(r, "DialogIcon", out.DialogIcon)
	out.DialogTimeout = parseIntOrDefault(r.FormValue("DialogTimeout"), out.DialogTimeout)

	out.PromptTitle = formOrDefault(r, "PromptTitle", out.PromptTitle)
	out.PromptMessage = formOrDefault(r, "PromptMessage", out.PromptMessage)
	out.PromptAlignment = formOrDefault(r, "PromptAlignment", out.PromptAlignment)
	out.PromptLeftButton = formOrDefault(r, "PromptLeftButton", out.PromptLeftButton)
	out.PromptMiddleButton = formOrDefault(r, "PromptMiddleButton", out.PromptMiddleButton)
	out.PromptRightButton = formOrDefault(r, "PromptRightButton", out.PromptRightButton)
	out.PromptIcon = formOrDefault(r, "PromptIcon", out.PromptIcon)
	out.PromptTimeout = parseIntOrDefault(r.FormValue("PromptTimeout"), out.PromptTimeout)
	out.PromptPos = formOrDefault(r, "PromptPos", out.PromptPos)
	out.PromptAllowMove = parseBoolOrDefault(r.FormValue("PromptAllowMove"), out.PromptAllowMove)
	out.PromptNotTopMost = parseBoolOrDefault(r.FormValue("PromptNotTopMost"), out.PromptNotTopMost)

	out.BalloonTitle = formOrDefault(r, "BalloonTitle", out.BalloonTitle)
	out.BalloonText = formOrDefault(r, "BalloonText", out.BalloonText)
	out.BalloonIcon = formOrDefault(r, "BalloonIcon", out.BalloonIcon)
	out.BalloonTime = parseIntOrDefault(r.FormValue("BalloonTime"), out.BalloonTime)

	out.ProgressMessage = formOrDefault(r, "ProgressMessage", out.ProgressMessage)
	out.ProgressDetail = formOrDefault(r, "ProgressDetail", out.ProgressDetail)
	out.ProgressPercent = formOrDefault(r, "ProgressPercent", out.ProgressPercent)
	out.ProgressAlign = formOrDefault(r, "ProgressAlign", out.ProgressAlign)
	out.ProgressPos = formOrDefault(r, "ProgressPos", out.ProgressPos)
	out.ProgressAllowMove = parseBoolOrDefault(r.FormValue("ProgressAllowMove"), out.ProgressAllowMove)
	out.ProgressNotTopMost = parseBoolOrDefault(r.FormValue("ProgressNotTopMost"), out.ProgressNotTopMost)

	out.WelcomeProcesses = formOrDefault(r, "WelcomeProcesses", out.WelcomeProcesses)
	out.WelcomeCountdown = parseIntOrDefault(r.FormValue("WelcomeCountdown"), out.WelcomeCountdown)
	out.WelcomeDiskCheck = parseBoolOrDefault(r.FormValue("WelcomeDiskCheck"), out.WelcomeDiskCheck)

	out.SelectedDialogStyle = formOrDefault(r, "SelectedDialogStyle", out.SelectedDialogStyle)
	out.SelectedTheme = formOrDefault(r, "SelectedTheme", out.SelectedTheme)
	out.FluentAccentColor = formOrDefault(r, "FluentAccentColor", out.FluentAccentColor)
	out.AssetAppIcon = formOrDefault(r, "AssetAppIcon", out.AssetAppIcon)
	out.AssetAppIconDark = formOrDefault(r, "AssetAppIconDark", out.AssetAppIconDark)
	out.AssetBannerClassic = formOrDefault(r, "AssetBannerClassic", out.AssetBannerClassic)

	return out
}

func handleInstallToolkit(w http.ResponseWriter, r *http.Request, app *appState, scope string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	app.logger.Info("instalando toolkit via PSGallery", "scope", scope)
	err := installToolkitFromPSGallery(r.Context(), scope)
	if err != nil {
		app.logger.Error("falha na instalacao do toolkit", "scope", scope, "err", err)
		app.mu.Lock()
		app.page.Error = fmt.Sprintf("Falha na instalacao (%s): %v", scope, err)
		app.page.Message = ""
		app.mu.Unlock()
		redirectHome(w, r)
		return
	}

	status, serr := checkToolkitInstalled(r.Context())
	if serr != nil {
		app.logger.Error("instalou mas falhou ao verificar", "err", serr)
		app.mu.Lock()
		app.page.Error = fmt.Sprintf("Instalou, mas falhou ao verificar: %v", serr)
		app.page.Message = ""
		app.mu.Unlock()
		redirectHome(w, r)
		return
	}

	app.logger.Info("toolkit instalado", "scope", scope, "version", status.Version)
	app.mu.Lock()
	app.page.ToolkitInstalled = status.Installed
	app.page.ToolkitVersion = status.Version
	app.page.ToolkitInstallPath = status.Path
	app.page.Error = ""
	app.page.Message = fmt.Sprintf("Toolkit instalado com sucesso no escopo %s.", scope)
	app.mu.Unlock()

	// Reconecta o client persistente com o modulo recem-instalado.
	// Adquire sessionMu para garantir que nenhuma sessao ativa usa app.client
	// enquanto o ponteiro e zerado (invariante: invalidar exige sessionMu).
	app.sessionMu.Lock()
	app.invalidateClient()
	app.sessionMu.Unlock()
	// ensureClient em background — nao bloqueia a resposta ao usuario.
	go func() {
		if cerr := app.ensureClient(); cerr != nil {
			app.logger.Warn("toolkit instalado mas client nao reconectou ainda", "err", cerr)
		}
	}()

	redirectHome(w, r)
}

func checkToolkitInstalled(ctx context.Context) (toolkitStatus, error) {
	cmd := `$m = Get-Module -Name 'PSAppDeployToolkit' -ListAvailable | Sort-Object Version -Descending | Select-Object -First 1; if ($null -eq $m) { @{ Installed = $false; Version = ''; Path = '' } | ConvertTo-Json -Compress } else { @{ Installed = $true; Version = $m.Version.ToString(); Path = $m.Path } | ConvertTo-Json -Compress }`
	out, err := runPowerShell(ctx, cmd)
	if err != nil {
		return toolkitStatus{}, err
	}

	var parsed toolkitStatus
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return toolkitStatus{}, fmt.Errorf("falha ao parsear status do toolkit: %w", err)
	}
	return parsed, nil
}

func installToolkitFromPSGallery(ctx context.Context, scope string) error {
	cmd := fmt.Sprintf(`$ErrorActionPreference = 'Stop'; [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; if (-not (Get-PSRepository -Name PSGallery -ErrorAction SilentlyContinue)) { Register-PSRepository -Default }; Set-PSRepository -Name PSGallery -InstallationPolicy Trusted; Install-Module -Name %s -Scope %s -Force -AllowClobber -Repository PSGallery`, moduleName, scope)
	_, err := runPowerShell(ctx, cmd)
	return err
}

func runPowerShell(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func parseProcesses(csv string) []types.ProcessDefinition {
	parts := strings.Split(csv, ",")
	out := make([]types.ProcessDefinition, 0, len(parts))
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name == "" {
			continue
		}
		out = append(out, types.ProcessDefinition{Name: name})
	}
	return out
}

func formOrDefault(r *http.Request, key, fallback string) string {
	v := strings.TrimSpace(r.FormValue(key))
	if v == "" {
		return fallback
	}
	return v
}

func parseIntOrDefault(v string, fallback int) int {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseFloatOrDefault(v string, fallback float64) float64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return n
}

func parseBoolOrDefault(v string, fallback bool) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func buildToolkitConfigSnippet(f formData) string {
	// This snippet can be copied to PSADT config.psd1 for Fluent/Classic asset testing.
	return fmt.Sprintf(`@{
    UI = @{
        DialogStyle = '%s'
        FluentAccentColor = '%s'
    }
    Assets = @{
        AppIcon = '%s'
        AppIconDark = '%s'
        BannerClassic = '%s'
    }
    # Theme sugerido para execucao do laboratorio: %s
}`,
		f.SelectedDialogStyle,
		f.FluentAccentColor,
		escapePS(f.AssetAppIcon),
		escapePS(f.AssetAppIconDark),
		escapePS(f.AssetBannerClassic),
		f.SelectedTheme,
	)
}

func escapePS(v string) string {
	return strings.ReplaceAll(v, "'", "''")
}

func renderPage(w http.ResponseWriter, data *pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTmpl.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
	}
}

func redirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
