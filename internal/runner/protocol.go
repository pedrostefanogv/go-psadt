//go:build windows

package runner

import "fmt"

const (
	// BeginMarker marks the start of a command response.
	BeginMarker = "<<<PSADT_BEGIN>>>"
	// EndMarker marks the end of a command response.
	EndMarker = "<<<PSADT_END>>>"
)

// WrapCommand wraps a PowerShell command with try/catch error handling and JSON output.
// It wraps the result in a {Success, Data, Error} JSON envelope delimited by markers.
func WrapCommand(psCommand string) string {
	return fmt.Sprintf(`
try {
    $result = %s
    $__out = @{ Success = $true; Data = $result; Error = $null } | ConvertTo-Json -Depth 10 -Compress
} catch {
    $__out = @{ Success = $false; Data = $null; Error = @{
        Message = $_.Exception.Message
        Type = $_.Exception.GetType().FullName
        StackTrace = $_.ScriptStackTrace
    }} | ConvertTo-Json -Depth 10 -Compress
}
Write-Output '%s'
Write-Output $__out
Write-Output '%s'
`, psCommand, BeginMarker, EndMarker)
}

// WrapVoidCommand wraps a PowerShell command that returns no value.
func WrapVoidCommand(psCommand string) string {
	return fmt.Sprintf(`
try {
    %s
    $__out = @{ Success = $true; Data = $null; Error = $null } | ConvertTo-Json -Depth 10 -Compress
} catch {
    $__out = @{ Success = $false; Data = $null; Error = @{
        Message = $_.Exception.Message
        Type = $_.Exception.GetType().FullName
        StackTrace = $_.ScriptStackTrace
    }} | ConvertTo-Json -Depth 10 -Compress
}
Write-Output '%s'
Write-Output $__out
Write-Output '%s'
`, psCommand, BeginMarker, EndMarker)
}

// ImportModuleCommand returns the PS command to import the PSADT module.
func ImportModuleCommand(moduleName string) string {
	return fmt.Sprintf("Import-Module -Name '%s' -Force -ErrorAction Stop", moduleName)
}

// CheckModuleVersionCommand returns a PS command to check the PSADT module version.
func CheckModuleVersionCommand(moduleName, minVersion string) string {
	return fmt.Sprintf(`
$m = Get-Module -Name '%s' -ListAvailable | Sort-Object Version -Descending | Select-Object -First 1
if (-not $m) { throw "Module '%s' is not installed" }
if ($m.Version -lt [version]'%s') { throw "Module '%s' version $($m.Version) is below minimum required %s" }
$m.Version.ToString()
`, moduleName, moduleName, minVersion, moduleName, minVersion)
}

// HeartbeatCommand returns a simple PS command to verify the process is alive.
func HeartbeatCommand() string {
	return "$true"
}
