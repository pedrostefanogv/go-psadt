//go:build windows

package types

// LogEntryOptions options for Write-ADTLogEntry.
type LogEntryOptions struct {
	Message          string      `ps:"Message"`
	Severity         LogSeverity `ps:"Severity"`
	Source           string      `ps:"Source"`
	ScriptSection    string      `ps:"ScriptSection"`
	LogType          string      `ps:"LogType"`
	LogFileDirectory string      `ps:"LogFileDirectory"`
	LogFileName      string      `ps:"LogFileName"`
	PassThru         bool        `ps:"PassThru,switch"`
	DebugMessage     bool        `ps:"DebugMessage,switch"`
}

// LogFileNameOptions options for New-ADTLogFileName.
type LogFileNameOptions struct {
	AppName    string `ps:"AppName"`
	AppVersion string `ps:"AppVersion"`
	Extension  string `ps:"Extension"`
	UseDate    bool   `ps:"UseDate,switch"`
}
