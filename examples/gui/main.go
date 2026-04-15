//go:build windows

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pedrostefanogv/go-psadt"
	"github.com/pedrostefanogv/go-psadt/types"
)

const (
	moduleName      = "PSAppDeployToolkit"
	minimumVersion  = "4.1.0"
	defaultHTTPAddr = "127.0.0.1:17841"
)

type pageData struct {
	Message            string
	Error              string
	ToolkitInstalled   bool
	ToolkitVersion     string
	ToolkitInstallPath string
	GeneratedConfig    string
	CurrentUIConfig    string
	Form               formData
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

	BalloonTitle string
	BalloonText  string
	BalloonIcon  string
	BalloonTime  int

	ProgressMessage string
	ProgressDetail  string

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
    });
  </script>
</head>
<body>
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
        <div class="section-actions">
          <form method="post" action="/toolkit/check"><button class="primary" type="submit">Verificar toolkit</button></form>
          <form method="post" action="/toolkit/install/user"><button type="submit">Instalar CurrentUser</button></form>
          <form method="post" action="/toolkit/install/allusers"><button type="submit">Instalar AllUsers</button></form>
        </div>
        <p class="sub">A instalacao usa Install-Module via PowerShell Gallery e tenta adicionar PSGallery como Trusted para evitar prompts.</p>

        <h3>Sessao base</h3>
        <form method="post" action="/run/welcome">
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
        <form method="post" action="/run/dialog">
          <div class="row">
            <div class="field"><label>Titulo modal</label><input name="DialogTitle" value="{{.Form.DialogTitle}}"></div>
            <div class="field"><label>Icone modal</label>
              <select name="DialogIcon">
                <option {{if eq .Form.DialogIcon "Question"}}selected{{end}}>Question</option>
                <option {{if eq .Form.DialogIcon "Information"}}selected{{end}}>Information</option>
                <option {{if eq .Form.DialogIcon "Warning"}}selected{{end}}>Warning</option>
                <option {{if eq .Form.DialogIcon "Error"}}selected{{end}}>Error</option>
                <option {{if eq .Form.DialogIcon "Shield"}}selected{{end}}>Shield</option>
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
        <form method="post" action="/run/prompt">
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
          </div>
          <button class="primary" type="submit">Testar prompt (ShowInstallationPrompt)</button>
        </form>

        <h3>Balloon alert</h3>
        <form method="post" action="/run/balloon">
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
        <form method="post" action="/run/progress">
          <div class="row">
            <div class="field"><label>Status</label><input name="ProgressMessage" value="{{.Form.ProgressMessage}}"></div>
            <div class="field"><label>Detalhe</label><input name="ProgressDetail" value="{{.Form.ProgressDetail}}"></div>
          </div>
          <button class="primary" type="submit">Exibir progress (6s)</button>
        </form>

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
          <form method="post" action="/config/read"><button class="primary" type="submit">Ler GetConfig atual</button></form>
          <form method="post" action="/query/environment"><button type="submit">Rodar query de ambiente</button></form>
        </div>
        {{if .CurrentUIConfig}}
        <h3>Config atual (Get-ADTConfig)</h3>
        <div class="pill mono">{{.CurrentUIConfig}}</div>
        {{end}}
        <p class="sub">O estilo Fluent/Classic, assets e cores dependem da configuracao ativa do toolkit. O app gera snippet para facilitar teste rapido em laboratorio.</p>
      </div>
    </div>
  </div>
</body>
</html>`))

func main() {
	status, _ := checkToolkitInstalled(context.Background())
	state := &pageData{
		ToolkitInstalled:   status.Installed,
		ToolkitVersion:     status.Version,
		ToolkitInstallPath: status.Path,
		Form:               defaultFormData(),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		renderPage(w, state)
	})

	http.HandleFunc("/toolkit/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		status, err := checkToolkitInstalled(r.Context())
		if err != nil {
			state.Error = fmt.Sprintf("Falha ao verificar toolkit: %v", err)
			state.Message = ""
			redirectHome(w, r)
			return
		}
		state.ToolkitInstalled = status.Installed
		state.ToolkitVersion = status.Version
		state.ToolkitInstallPath = status.Path
		state.Error = ""
		if status.Installed {
			state.Message = fmt.Sprintf("Toolkit encontrado: versao %s", status.Version)
		} else {
			state.Message = "Toolkit nao encontrado. Use os botoes de instalacao via PSGallery."
		}
		redirectHome(w, r)
	})

	http.HandleFunc("/toolkit/install/user", func(w http.ResponseWriter, r *http.Request) {
		handleInstallToolkit(w, r, state, "CurrentUser")
	})
	http.HandleFunc("/toolkit/install/allusers", func(w http.ResponseWriter, r *http.Request) {
		handleInstallToolkit(w, r, state, "AllUsers")
	})

	http.HandleFunc("/run/welcome", func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, state, func(s *psadt.Session, f formData) (string, error) {
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
	})

	http.HandleFunc("/run/dialog", func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, state, func(s *psadt.Session, f formData) (string, error) {
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
	})

	http.HandleFunc("/run/prompt", func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, state, func(s *psadt.Session, f formData) (string, error) {
			res, err := s.ShowInstallationPrompt(types.PromptOptions{
				Title:            f.PromptTitle,
				Message:          f.PromptMessage,
				MessageAlignment: types.MessageAlignment(f.PromptAlignment),
				ButtonLeftText:   f.PromptLeftButton,
				ButtonMiddleText: f.PromptMiddleButton,
				ButtonRightText:  f.PromptRightButton,
				Icon:             types.DialogSystemIcon(f.PromptIcon),
				Timeout:          f.PromptTimeout,
			})
			if err != nil {
				return "", err
			}
			if res == nil {
				return "ShowInstallationPrompt executado sem retorno de botao.", nil
			}
			return fmt.Sprintf("ShowInstallationPrompt concluido. ButtonClicked=%s", res.ButtonClicked), nil
		})
	})

	http.HandleFunc("/run/balloon", func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, state, func(s *psadt.Session, f formData) (string, error) {
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
	})

	http.HandleFunc("/run/progress", func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, state, func(s *psadt.Session, f formData) (string, error) {
			err := s.ShowInstallationProgress(types.ProgressOptions{
				StatusMessage:       f.ProgressMessage,
				StatusMessageDetail: f.ProgressDetail,
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
	})

	http.HandleFunc("/config/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = r.ParseForm()
		state.Form = mergeForm(state.Form, r)
		state.GeneratedConfig = buildToolkitConfigSnippet(state.Form)
		state.Error = ""
		state.Message = "Snippet de configuracao gerado."
		redirectHome(w, r)
	})

	http.HandleFunc("/config/read", func(w http.ResponseWriter, r *http.Request) {
		runWithSession(w, r, state, func(s *psadt.Session, _ formData) (string, error) {
			cfg, err := s.GetConfig()
			if err != nil {
				return "", err
			}
			encoded, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return "", err
			}
			state.CurrentUIConfig = string(encoded)
			return "GetConfig carregado para validar estilo/modal/cores atuais do toolkit.", nil
		})
	})

	http.HandleFunc("/query/environment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		client, err := psadt.NewClient(psadt.WithTimeout(2 * time.Minute))
		if err != nil {
			state.Error = fmt.Sprintf("Falha ao criar client: %v", err)
			state.Message = ""
			redirectHome(w, r)
			return
		}
		defer client.Close()

		env, err := client.GetEnvironment()
		if err != nil {
			state.Error = fmt.Sprintf("GetEnvironment falhou: %v", err)
			state.Message = ""
			redirectHome(w, r)
			return
		}

		state.CurrentUIConfig = fmt.Sprintf(
			"OS=%s %s\\nPowerShell=%s\\nIsAdmin=%v\\nToolkit=%s %s",
			env.OS.Name,
			env.OS.Version,
			env.PowerShell.PSVersion,
			env.Permissions.IsAdmin,
			env.Toolkit.ShortName,
			env.Toolkit.Version,
		)
		state.Error = ""
		state.Message = "Consulta de ambiente concluida."
		redirectHome(w, r)
	})

	log.Printf("go-psadt UI Lab em http://%s", defaultHTTPAddr)
	if err := http.ListenAndServe(defaultHTTPAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func runWithSession(
	w http.ResponseWriter,
	r *http.Request,
	state *pageData,
	action func(s *psadt.Session, f formData) (string, error),
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	f := mergeForm(state.Form, r)
	state.Form = f

	client, err := psadt.NewClient(psadt.WithTimeout(3 * time.Minute))
	if err != nil {
		state.Error = fmt.Sprintf("Falha ao criar client. Verifique se o toolkit esta instalado: %v", err)
		state.Message = ""
		redirectHome(w, r)
		return
	}
	defer client.Close()

	session, err := client.OpenSession(types.SessionConfig{
		DeploymentType: types.DeploymentType(f.DeploymentType),
		DeployMode:     types.DeployMode(f.DeployMode),
		AppVendor:      f.AppVendor,
		AppName:        f.AppName,
		AppVersion:     f.AppVersion,
	})
	if err != nil {
		state.Error = fmt.Sprintf("Falha ao abrir sessao ADT: %v", err)
		state.Message = ""
		redirectHome(w, r)
		return
	}
	defer session.Close(0)

	msg, err := action(session, f)
	if err != nil {
		state.Error = fmt.Sprintf("Acao falhou: %v", err)
		state.Message = ""
		redirectHome(w, r)
		return
	}

	status, _ := checkToolkitInstalled(r.Context())
	state.ToolkitInstalled = status.Installed
	state.ToolkitVersion = status.Version
	state.ToolkitInstallPath = status.Path
	state.Error = ""
	state.Message = msg
	redirectHome(w, r)
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
		BalloonTitle:        "go-psadt UI Lab",
		BalloonText:         "Notificacao de teste enviada com sucesso.",
		BalloonIcon:         string(types.BalloonInfo),
		BalloonTime:         8000,
		ProgressMessage:     "Executando teste de progresso...",
		ProgressDetail:      "Aguarde alguns segundos para fechamento automatico.",
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

	out.BalloonTitle = formOrDefault(r, "BalloonTitle", out.BalloonTitle)
	out.BalloonText = formOrDefault(r, "BalloonText", out.BalloonText)
	out.BalloonIcon = formOrDefault(r, "BalloonIcon", out.BalloonIcon)
	out.BalloonTime = parseIntOrDefault(r.FormValue("BalloonTime"), out.BalloonTime)

	out.ProgressMessage = formOrDefault(r, "ProgressMessage", out.ProgressMessage)
	out.ProgressDetail = formOrDefault(r, "ProgressDetail", out.ProgressDetail)

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

func handleInstallToolkit(w http.ResponseWriter, r *http.Request, state *pageData, scope string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := installToolkitFromPSGallery(r.Context(), scope)
	if err != nil {
		state.Error = fmt.Sprintf("Falha na instalacao (%s): %v", scope, err)
		state.Message = ""
		redirectHome(w, r)
		return
	}

	status, serr := checkToolkitInstalled(r.Context())
	if serr != nil {
		state.Error = fmt.Sprintf("Instalou, mas falhou ao verificar: %v", serr)
		state.Message = ""
		redirectHome(w, r)
		return
	}

	state.ToolkitInstalled = status.Installed
	state.ToolkitVersion = status.Version
	state.ToolkitInstallPath = status.Path
	state.Error = ""
	state.Message = fmt.Sprintf("Toolkit instalado com sucesso no escopo %s.", scope)
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
