//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
	"github.com/pedrostefanogv/go-psadt/types"
)

// CopyFile copies files from source to destination.
func (s *Session) CopyFile(opts types.CopyFileOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Copy-ADTFile", opts)
	return s.executeVoid(ctx, cmd)
}

// CopyFileToUserProfiles copies files to all user profiles.
func (s *Session) CopyFileToUserProfiles(opts types.CopyFileToUserProfilesOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Copy-ADTFileToUserProfiles", opts)
	return s.executeVoid(ctx, cmd)
}

// RemoveFile removes files.
func (s *Session) RemoveFile(opts types.RemoveFileOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Remove-ADTFile", opts)
	return s.executeVoid(ctx, cmd)
}

// RemoveFileFromUserProfiles removes files from all user profiles.
func (s *Session) RemoveFileFromUserProfiles(opts types.RemoveFileFromUserProfilesOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Remove-ADTFileFromUserProfiles", opts)
	return s.executeVoid(ctx, cmd)
}

// NewFolder creates a new folder.
func (s *Session) NewFolder(path string) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := fmt.Sprintf("New-ADTFolder -Path %s", cmdbuilder.EscapeString(path))
	return s.executeVoid(ctx, cmd)
}

// RemoveFolder removes a folder.
func (s *Session) RemoveFolder(opts types.RemoveFolderOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Remove-ADTFolder", opts)
	return s.executeVoid(ctx, cmd)
}

// CopyContentToCache copies content to the PSADT cache directory.
func (s *Session) CopyContentToCache(opts types.CopyContentToCacheOptions) (string, error) {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Copy-ADTContentToCache", opts)
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// RemoveContentFromCache removes content from the PSADT cache directory.
func (s *Session) RemoveContentFromCache(opts types.RemoveContentFromCacheOptions) error {
	ctx, cancel := s.client.defaultContext()
	defer cancel()
	cmd := cmdbuilder.Build("Remove-ADTContentFromCache", opts)
	return s.executeVoid(ctx, cmd)
}
