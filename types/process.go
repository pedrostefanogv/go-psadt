//go:build windows

package types

// ProcessResult is the return of StartProcess when PassThru=true.
type ProcessResult struct {
	ExitCode    int    `json:"ExitCode"`
	StdOut      string `json:"StdOut"`
	StdErr      string `json:"StdErr"`
	Interleaved string `json:"Interleaved"`
}

// ProcessDefinition defines a process for CloseProcesses/BlockExecution.
type ProcessDefinition struct {
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
}

// StartProcessOptions options for Start-ADTProcess.
type StartProcessOptions struct {
	FilePath           string             `ps:"FilePath"`
	ArgumentList       []string           `ps:"ArgumentList"`
	SecureArgumentList bool               `ps:"SecureArgumentList,switch"`
	WindowStyle        ProcessWindowStyle `ps:"WindowStyle"`
	CreateNoWindow     bool               `ps:"CreateNoWindow,switch"`
	WorkingDirectory   string             `ps:"WorkingDirectory"`
	NoWait             bool               `ps:"NoWait,switch"`
	PassThru           bool               `ps:"PassThru,switch"`
	WaitForMsiExec     bool               `ps:"WaitForMsiExec,switch"`
	MsiExecWaitTime    int                `ps:"MsiExecWaitTime"`
	IgnoreExitCodes    []int              `ps:"IgnoreExitCodes"`
	PriorityClass      string             `ps:"PriorityClass"`
	UseShellExecute    bool               `ps:"UseShellExecute,switch"`
}

// StartProcessAsUserOptions options for Start-ADTProcessAsUser.
type StartProcessAsUserOptions struct {
	FilePath           string             `ps:"FilePath"`
	ArgumentList       []string           `ps:"ArgumentList"`
	SecureArgumentList bool               `ps:"SecureArgumentList,switch"`
	WindowStyle        ProcessWindowStyle `ps:"WindowStyle"`
	RunLevel           string             `ps:"RunLevel"`
	Wait               bool               `ps:"Wait,switch"`
	PassThru           bool               `ps:"PassThru,switch"`
	WorkingDirectory   string             `ps:"WorkingDirectory"`
}

// MsiProcessOptions options for Start-ADTMsiProcess.
type MsiProcessOptions struct {
	Action                  MsiAction `ps:"Action"`
	FilePath                string    `ps:"FilePath"`
	Transforms              []string  `ps:"Transforms"`
	Patches                 []string  `ps:"Patches"`
	AdditionalArgumentList  []string  `ps:"AdditionalArgumentList"`
	SecureArgumentList      bool      `ps:"SecureArgumentList,switch"`
	PassThru                bool      `ps:"PassThru,switch"`
	SkipMSIAlreadyInstalled bool      `ps:"SkipMSIAlreadyInstalled,switch"`
	RepairFromSource        bool      `ps:"RepairFromSource,switch"`
	LoggingOptions          string    `ps:"LoggingOptions"`
	LogFileName             string    `ps:"LogFileName"`
	WorkingDirectory        string    `ps:"WorkingDirectory"`
}

// MsiProcessAsUserOptions options for Start-ADTMsiProcessAsUser.
type MsiProcessAsUserOptions struct {
	Action                 MsiAction `ps:"Action"`
	FilePath               string    `ps:"FilePath"`
	Transforms             []string  `ps:"Transforms"`
	Patches                []string  `ps:"Patches"`
	AdditionalArgumentList []string  `ps:"AdditionalArgumentList"`
	SecureArgumentList     bool      `ps:"SecureArgumentList,switch"`
	PassThru               bool      `ps:"PassThru,switch"`
	Wait                   bool      `ps:"Wait,switch"`
	WorkingDirectory       string    `ps:"WorkingDirectory"`
}

// MspProcessOptions options for Start-ADTMspProcess.
type MspProcessOptions struct {
	FilePath               string   `ps:"FilePath"`
	AdditionalArgumentList []string `ps:"AdditionalArgumentList"`
	SecureArgumentList     bool     `ps:"SecureArgumentList,switch"`
	PassThru               bool     `ps:"PassThru,switch"`
}

// MspProcessAsUserOptions options for Start-ADTMspProcessAsUser.
type MspProcessAsUserOptions struct {
	FilePath               string   `ps:"FilePath"`
	AdditionalArgumentList []string `ps:"AdditionalArgumentList"`
	SecureArgumentList     bool     `ps:"SecureArgumentList,switch"`
	PassThru               bool     `ps:"PassThru,switch"`
	Wait                   bool     `ps:"Wait,switch"`
}
