//go:build windows

package psadt

import (
	"context"
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// GetEnvironment collects all PSADT environment variables and returns them
// as a structured EnvironmentInfo. This works at the Client level and does
// not require an open session. Results are cached after first call; call
// InvalidateEnvCache() to force fresh data.
func (c *Client) GetEnvironment() (*types.EnvironmentInfo, error) {
	ctx, cancel := c.defaultContext()
	defer cancel()
	return c.GetEnvironmentWithContext(ctx)
}

// GetEnvironmentWithContext collects environment variables with an explicit context.
func (c *Client) GetEnvironmentWithContext(ctx context.Context) (*types.EnvironmentInfo, error) {
	c.envMu.Lock()
	if c.envCached && c.envCache != nil {
		env := c.envCache
		c.envMu.Unlock()
		return env, nil
	}
	c.envMu.Unlock()

	// Build a PS command that collects all PSADT env vars into a structured hashtable
	cmd := `
@{
    Toolkit = @{
        FriendlyName = $appDeployMainScriptFriendlyName
        ShortName = $appDeployToolkitName
        Version = if ($appDeployMainScriptVersion) { $appDeployMainScriptVersion.ToString() } else { '' }
    }
    Culture = @{
        Language = $currentLanguage
        UILanguage = $currentUILanguage
    }
    Paths = @{
        ProgramFiles = $envProgramFiles
        ProgramFilesX86 = $envProgramFilesX86
        ProgramData = $envProgramData
        SystemRoot = $envSystemRoot
        SystemDrive = $envSystemDrive
        System32Directory = $envSystem32Directory
        WinDir = $envWinDir
        Temp = $envTemp
        CommonProgramFiles = $envCommonProgramFiles
        CommonProgramFilesX86 = $envCommonProgramFilesX86
        Public = $envPublic
        UserProfile = $envUserProfile
        AppData = $envAppData
        LocalAppData = $envLocalAppData
        UserDesktop = $envUserDesktop
        UserDocuments = $envUserMyDocuments
        UserStartMenu = $envUserStartMenu
        UserStartMenuPrograms = $envUserStartMenuPrograms
        UserStartUp = $envUserStartUp
        AllUsersProfile = $envAllUsersProfile
        CommonDesktop = $envCommonDesktop
        CommonDocuments = $envCommonDocuments
        CommonStartMenu = $envCommonStartMenu
        CommonStartMenuPrograms = $envCommonStartMenuPrograms
        CommonStartUp = $envCommonStartUp
        CommonTemplates = $envCommonTemplates
        HomeDrive = $envHomeDrive
        HomePath = $envHomePath
        HomeShare = $envHomeShare
        ComputerName = $envComputerName
        ComputerNameFQDN = $envComputerNameFQDN
        UserName = $envUserName
        LogicalDrives = @($envLogicalDrives)
        SystemRAM = if ($envSystemRAM) { $envSystemRAM } else { 0 }
    }
    Domain = @{
        IsMachinePartOfDomain = [bool]$IsMachinePartOfDomain
        MachineADDomain = $envMachineADDomain
        MachineDNSDomain = $envMachineDNSDomain
        MachineWorkgroup = $envMachineWorkgroup
        MachineDomainController = $MachineDomainController
        UserDNSDomain = $envUserDNSDomain
        UserDomain = $envUserDomain
        LogonServer = $envLogonServer
    }
    OS = @{
        Name = $envOSName
        Version = if ($envOSVersion) { $envOSVersion.ToString() } else { '' }
        VersionMajor = if ($envOSVersionMajor) { $envOSVersionMajor } else { 0 }
        VersionMinor = if ($envOSVersionMinor) { $envOSVersionMinor } else { 0 }
        VersionBuild = if ($envOSVersionBuild) { $envOSVersionBuild } else { 0 }
        VersionRevision = if ($envOSVersionRevision) { $envOSVersionRevision } else { 0 }
        Architecture = $envOSArchitecture
        ServicePack = $envOSServicePack
        ProductType = if ($envOSProductType) { $envOSProductType } else { 0 }
        ProductTypeName = $envOSProductTypeName
        Is64Bit = [bool]$Is64Bit
        IsServerOS = [bool]$IsServerOS
        IsWorkStationOS = [bool]$IsWorkStationOS
        IsDomainControllerOS = [bool]$IsDomainControllerOS
    }
    Process = @{
        Is64BitProcess = [bool]$Is64BitProcess
        Architecture = $psArchitecture
    }
    PowerShell = @{
        PSVersion = if ($envPSVersion) { $envPSVersion.ToString() } else { '' }
        PSVersionMajor = if ($envPSVersionMajor) { $envPSVersionMajor } else { 0 }
        PSVersionMinor = if ($envPSVersionMinor) { $envPSVersionMinor } else { 0 }
        PSVersionBuild = if ($envPSVersionBuild) { $envPSVersionBuild } else { 0 }
        PSVersionRevision = if ($envPSVersionRevision) { $envPSVersionRevision } else { 0 }
        CLRVersion = if ($envCLRVersion) { $envCLRVersion.ToString() } else { '' }
        CLRVersionMajor = if ($envCLRVersionMajor) { $envCLRVersionMajor } else { 0 }
        CLRVersionMinor = if ($envCLRVersionMinor) { $envCLRVersionMinor } else { 0 }
    }
    Permissions = @{
        IsAdmin = [bool]$IsAdmin
        IsLocalSystemAccount = [bool]$IsLocalSystemAccount
        IsLocalServiceAccount = [bool]$IsLocalServiceAccount
        IsNetworkServiceAccount = [bool]$IsNetworkServiceAccount
        IsServiceAccount = [bool]$IsServiceAccount
        IsProcessUserInteractive = [bool]$IsProcessUserInteractive
        SessionZero = [bool]$SessionZero
        ProcessNTAccount = $ProcessNTAccount
        ProcessNTAccountSID = $ProcessNTAccountSID
        CurrentProcessSID = $CurrentProcessSID
        LocalSystemNTAccount = $LocalSystemNTAccount
        LocalAdministratorsGroup = $LocalAdministratorsGroup
        LocalUsersGroup = $LocalUsersGroup
    }
    Users = @{
        LoggedOnUserSessions = @(if ($LoggedOnUserSessions) { $LoggedOnUserSessions } else { @() })
        CurrentConsoleUserSession = $CurrentConsoleUserSession
        CurrentLoggedOnUserSession = $CurrentLoggedOnUserSession
        RunAsActiveUser = $RunAsActiveUser
        UsersLoggedOn = @(if ($UsersLoggedOn) { $UsersLoggedOn } else { @() })
    }
    Office = @{
        Bitness = $envOfficeBitness
        Channel = $envOfficeChannel
        Version = $envOfficeVersion
    }
    Misc = @{
        RunningTaskSequence = [bool]$RunningTaskSequence
    }
} | ConvertTo-Json -Depth 5
`

	c.logger.Debug("collecting PSADT environment variables")

	data, err := c.runner.Execute(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to collect environment info: %w", err)
	}

	var env types.EnvironmentInfo
	if err := parser.ParseResponse(data, &env); err != nil {
		return nil, err
	}

	// Cache the result
	c.envMu.Lock()
	c.envCache = &env
	c.envCached = true
	c.envMu.Unlock()

	return &env, nil
}

// ExecuteRawScript executes an arbitrary PowerShell script block within the
// PSADT session. The script runs in the same persistent runner, inheriting
// the module context (session, variables, etc.). Returns raw JSON bytes.
//
// This is the escape hatch for RMM agents that need to run custom PSADT
// PowerShell logic not yet wrapped by the Go API.
func (c *Client) ExecuteRawScript(ctx context.Context, script string) ([]byte, error) {
	c.logger.Debug("executing raw script", "length", len(script))
	return c.runner.ExecuteRaw(ctx, script)
}

// ExecuteRawVoidScript is like ExecuteRawScript but for scripts with no return value.
func (c *Client) ExecuteRawVoidScript(ctx context.Context, script string) error {
	c.logger.Debug("executing raw void script", "length", len(script))
	return c.runner.ExecuteRawVoid(ctx, script)
}
