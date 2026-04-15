//go:build windows

package types

// RegSvr32Options options for Invoke-ADTRegSvr32.
type RegSvr32Options struct {
	FilePath        string `ps:"FilePath"`
	Action          string `ps:"Action"`
	ContinueOnError bool   `ps:"ContinueOnError,switch"`
}
