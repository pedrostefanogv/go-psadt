//go:build windows

package types

// MountWimOptions options for Mount-ADTWimFile.
type MountWimOptions struct {
	ImagePath string `ps:"ImagePath"`
	Path      string `ps:"Path"`
	Index     int    `ps:"Index"`
	Force     bool   `ps:"Force,switch"`
}

// DismountWimOptions options for Dismount-ADTWimFile.
type DismountWimOptions struct {
	ImagePath string `ps:"ImagePath"`
	Path      string `ps:"Path"`
	Save      bool   `ps:"Save,switch"`
}

// NewZipOptions options for New-ADTZipFile.
type NewZipOptions struct {
	SourceDirectoryPath        string `ps:"SourceDirectoryPath"`
	DestinationArchiveFileName string `ps:"DestinationArchiveFileName"`
	OverWriteArchive           bool   `ps:"OverWriteArchive,switch"`
}
