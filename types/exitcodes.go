//go:build windows

package types

const (
	ExitCodeSuccess                   = 0
	ExitCodeReboot3010                = 3010
	ExitCodeReboot1641                = 1641
	ExitCodeUserCancelled             = 1602
	ExitCodeScriptError               = 60001
	ExitCodeModuleImportError         = 60008
	ExitCodeExePreLaunchError         = 60010
	ExitCodeExeLaunchError            = 60011
	ExitCodeBuiltInRangeStart         = 60000
	ExitCodeBuiltInRangeEnd           = 68999
	ExitCodeUserCustomRangeStart      = 69000
	ExitCodeUserCustomRangeEnd        = 69999
	ExitCodeExtensionCustomRangeStart = 70000
	ExitCodeExtensionCustomRangeEnd   = 79999
)
