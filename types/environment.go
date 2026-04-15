//go:build windows

package types

// EnvironmentInfo contains all environment variables exposed by PSADT.
type EnvironmentInfo struct {
	Toolkit     ToolkitInfo     `json:"Toolkit"`
	Culture     CultureInfo     `json:"Culture"`
	Paths       SystemPaths     `json:"Paths"`
	Domain      DomainInfo      `json:"Domain"`
	OS          OSEnvironment   `json:"OS"`
	Process     ProcessInfo     `json:"Process"`
	PowerShell  PSVersionInfo   `json:"PowerShell"`
	Permissions PermissionsInfo `json:"Permissions"`
	Users       UsersInfo       `json:"Users"`
	Office      OfficeInfo      `json:"Office"`
	Misc        MiscInfo        `json:"Misc"`
}

// ToolkitInfo toolkit metadata.
type ToolkitInfo struct {
	FriendlyName string `json:"FriendlyName"`
	ShortName    string `json:"ShortName"`
	Version      string `json:"Version"`
}

// CultureInfo system language and culture.
type CultureInfo struct {
	Language   string `json:"Language"`
	UILanguage string `json:"UILanguage"`
}

// SystemPaths system and user paths (~40 variables).
type SystemPaths struct {
	ProgramFiles            string   `json:"ProgramFiles"`
	ProgramFilesX86         string   `json:"ProgramFilesX86"`
	ProgramData             string   `json:"ProgramData"`
	SystemRoot              string   `json:"SystemRoot"`
	SystemDrive             string   `json:"SystemDrive"`
	System32Directory       string   `json:"System32Directory"`
	WinDir                  string   `json:"WinDir"`
	Temp                    string   `json:"Temp"`
	CommonProgramFiles      string   `json:"CommonProgramFiles"`
	CommonProgramFilesX86   string   `json:"CommonProgramFilesX86"`
	Public                  string   `json:"Public"`
	UserProfile             string   `json:"UserProfile"`
	AppData                 string   `json:"AppData"`
	LocalAppData            string   `json:"LocalAppData"`
	UserDesktop             string   `json:"UserDesktop"`
	UserDocuments           string   `json:"UserDocuments"`
	UserStartMenu           string   `json:"UserStartMenu"`
	UserStartMenuPrograms   string   `json:"UserStartMenuPrograms"`
	UserStartUp             string   `json:"UserStartUp"`
	AllUsersProfile         string   `json:"AllUsersProfile"`
	CommonDesktop           string   `json:"CommonDesktop"`
	CommonDocuments         string   `json:"CommonDocuments"`
	CommonStartMenu         string   `json:"CommonStartMenu"`
	CommonStartMenuPrograms string   `json:"CommonStartMenuPrograms"`
	CommonStartUp           string   `json:"CommonStartUp"`
	CommonTemplates         string   `json:"CommonTemplates"`
	HomeDrive               string   `json:"HomeDrive"`
	HomePath                string   `json:"HomePath"`
	HomeShare               string   `json:"HomeShare"`
	ComputerName            string   `json:"ComputerName"`
	ComputerNameFQDN        string   `json:"ComputerNameFQDN"`
	UserName                string   `json:"UserName"`
	LogicalDrives           []string `json:"LogicalDrives"`
	SystemRAM               int      `json:"SystemRAM"`
}

// DomainInfo AD domain information.
type DomainInfo struct {
	IsMachinePartOfDomain   bool   `json:"IsMachinePartOfDomain"`
	MachineADDomain         string `json:"MachineADDomain"`
	MachineDNSDomain        string `json:"MachineDNSDomain"`
	MachineWorkgroup        string `json:"MachineWorkgroup"`
	MachineDomainController string `json:"MachineDomainController"`
	UserDNSDomain           string `json:"UserDNSDomain"`
	UserDomain              string `json:"UserDomain"`
	LogonServer             string `json:"LogonServer"`
}

// OSEnvironment expanded operating system information.
type OSEnvironment struct {
	Name                 string `json:"Name"`
	Version              string `json:"Version"`
	VersionMajor         int    `json:"VersionMajor"`
	VersionMinor         int    `json:"VersionMinor"`
	VersionBuild         int    `json:"VersionBuild"`
	VersionRevision      int    `json:"VersionRevision"`
	Architecture         string `json:"Architecture"`
	ServicePack          string `json:"ServicePack"`
	ProductType          int    `json:"ProductType"`
	ProductTypeName      string `json:"ProductTypeName"`
	Is64Bit              bool   `json:"Is64Bit"`
	IsServerOS           bool   `json:"IsServerOS"`
	IsWorkStationOS      bool   `json:"IsWorkStationOS"`
	IsDomainControllerOS bool   `json:"IsDomainControllerOS"`
}

// ProcessInfo current process architecture.
type ProcessInfo struct {
	Is64BitProcess bool   `json:"Is64BitProcess"`
	Architecture   string `json:"Architecture"`
}

// PSVersionInfo PowerShell and CLR/.NET versions.
type PSVersionInfo struct {
	PSVersion         string `json:"PSVersion"`
	PSVersionMajor    int    `json:"PSVersionMajor"`
	PSVersionMinor    int    `json:"PSVersionMinor"`
	PSVersionBuild    int    `json:"PSVersionBuild"`
	PSVersionRevision int    `json:"PSVersionRevision"`
	CLRVersion        string `json:"CLRVersion"`
	CLRVersionMajor   int    `json:"CLRVersionMajor"`
	CLRVersionMinor   int    `json:"CLRVersionMinor"`
}

// PermissionsInfo permissions and accounts of the current process.
type PermissionsInfo struct {
	IsAdmin                  bool   `json:"IsAdmin"`
	IsLocalSystemAccount     bool   `json:"IsLocalSystemAccount"`
	IsLocalServiceAccount    bool   `json:"IsLocalServiceAccount"`
	IsNetworkServiceAccount  bool   `json:"IsNetworkServiceAccount"`
	IsServiceAccount         bool   `json:"IsServiceAccount"`
	IsProcessUserInteractive bool   `json:"IsProcessUserInteractive"`
	SessionZero              bool   `json:"SessionZero"`
	ProcessNTAccount         string `json:"ProcessNTAccount"`
	ProcessNTAccountSID      string `json:"ProcessNTAccountSID"`
	CurrentProcessSID        string `json:"CurrentProcessSID"`
	LocalSystemNTAccount     string `json:"LocalSystemNTAccount"`
	LocalAdministratorsGroup string `json:"LocalAdministratorsGroup"`
	LocalUsersGroup          string `json:"LocalUsersGroup"`
}

// UsersInfo information about logged-on users.
type UsersInfo struct {
	LoggedOnUserSessions       []LoggedOnUserSession `json:"LoggedOnUserSessions"`
	CurrentConsoleUserSession  *LoggedOnUserSession  `json:"CurrentConsoleUserSession"`
	CurrentLoggedOnUserSession *LoggedOnUserSession  `json:"CurrentLoggedOnUserSession"`
	RunAsActiveUser            *LoggedOnUserSession  `json:"RunAsActiveUser"`
	UsersLoggedOn              []string              `json:"UsersLoggedOn"`
}

// LoggedOnUserSession details of a user session.
type LoggedOnUserSession struct {
	NTAccount        string `json:"NTAccount"`
	SID              string `json:"SID"`
	SessionID        int    `json:"SessionId"`
	IsConsoleSession bool   `json:"IsConsoleSession"`
	IsCurrentSession bool   `json:"IsCurrentSession"`
	IsAdmin          bool   `json:"IsAdmin"`
}

// OfficeInfo Microsoft Office information.
type OfficeInfo struct {
	Bitness string `json:"Bitness"`
	Channel string `json:"Channel"`
	Version string `json:"Version"`
}

// MiscInfo miscellaneous variables.
type MiscInfo struct {
	RunningTaskSequence bool `json:"RunningTaskSequence"`
}
