//go:build windows

package cmdbuilder

import "testing"

type benchWelcome struct {
	Title                   string   `ps:"Title"`
	CloseProcessesCountdown int      `ps:"CloseProcessesCountdown"`
	AllowDefer              bool     `ps:"AllowDefer,switch"`
	PromptToSave            bool     `ps:"PromptToSave,switch"`
	BlockExecution          bool     `ps:"BlockExecution,switch"`
	ArgList                 []string `ps:"ArgList"`
}

func BenchmarkBuild_SmallStruct(b *testing.B) {
	opts := benchWelcome{Title: "Firefox", CloseProcessesCountdown: 30, AllowDefer: true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Build("Show-ADTInstallationWelcome", opts)
	}
}

func BenchmarkBuild_LargeStruct(b *testing.B) {
	opts := benchWelcome{
		Title:                   "Firefox",
		CloseProcessesCountdown: 30,
		AllowDefer:              true,
		PromptToSave:            true,
		ArgList:                 []string{"/S", "--quiet", "--no-restart", "--log=c:\\temp\\x.log"},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Build("Show-ADTInstallationWelcome", opts)
	}
}

func BenchmarkEscapeString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = EscapeString("C:\\Program Files\\Mozilla Firefox\\firefox.exe 'quoted'")
	}
}
