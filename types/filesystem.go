//go:build windows

package types

// CopyFileOptions options for Copy-ADTFile.
type CopyFileOptions struct {
	Path            string `ps:"Path"`
	Destination     string `ps:"Destination"`
	Recurse         bool   `ps:"Recurse,switch"`
	Flatten         bool   `ps:"Flatten,switch"`
	ContinueOnError bool   `ps:"ContinueOnError,switch"`
	FileCopyMode    string `ps:"FileCopyMode"`
}

// CopyFileToUserProfilesOptions options for Copy-ADTFileToUserProfiles.
type CopyFileToUserProfilesOptions struct {
	Path             string   `ps:"Path"`
	Destination      string   `ps:"Destination"`
	Recurse          bool     `ps:"Recurse,switch"`
	Flatten          bool     `ps:"Flatten,switch"`
	ContinueOnError  bool     `ps:"ContinueOnError,switch"`
	ExcludeNTAccount []string `ps:"ExcludeNTAccount"`
	IncludeNTAccount []string `ps:"IncludeNTAccount"`
}

// RemoveFileOptions options for Remove-ADTFile.
type RemoveFileOptions struct {
	Path            string `ps:"Path"`
	LiteralPath     string `ps:"LiteralPath"`
	Recurse         bool   `ps:"Recurse,switch"`
	ContinueOnError bool   `ps:"ContinueOnError,switch"`
}

// RemoveFileFromUserProfilesOptions options for Remove-ADTFileFromUserProfiles.
type RemoveFileFromUserProfilesOptions struct {
	Path             string   `ps:"Path"`
	LiteralPath      string   `ps:"LiteralPath"`
	Recurse          bool     `ps:"Recurse,switch"`
	ContinueOnError  bool     `ps:"ContinueOnError,switch"`
	ExcludeNTAccount []string `ps:"ExcludeNTAccount"`
	IncludeNTAccount []string `ps:"IncludeNTAccount"`
}

// RemoveFolderOptions options for Remove-ADTFolder.
type RemoveFolderOptions struct {
	Path             string `ps:"Path"`
	LiteralPath      string `ps:"LiteralPath"`
	DisableRecursion bool   `ps:"DisableRecursion,switch"`
	ContinueOnError  bool   `ps:"ContinueOnError,switch"`
}

// CopyContentToCacheOptions options for Copy-ADTContentToCache.
type CopyContentToCacheOptions struct {
	Path string `ps:"Path"`
	Tag  string `ps:"Tag"`
}

// RemoveContentFromCacheOptions options for Remove-ADTContentFromCache.
type RemoveContentFromCacheOptions struct {
	Tag string `ps:"Tag"`
}
