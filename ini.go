//go:build windows

package psadt

import (
	"fmt"

	"github.com/pedrostefanogv/go-psadt/internal/cmdbuilder"
	"github.com/pedrostefanogv/go-psadt/internal/parser"
)

// GetIniValue reads a value from an INI file.
func (s *Session) GetIniValue(filePath, section, key string) (string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTIniValue -FilePath %s -Section %s -Key %s",
		cmdbuilder.EscapeString(filePath),
		cmdbuilder.EscapeString(section),
		cmdbuilder.EscapeString(key))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return "", err
	}
	return parser.ParseString(data)
}

// SetIniValue writes a value to an INI file.
func (s *Session) SetIniValue(filePath, section, key, value string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Set-ADTIniValue -FilePath %s -Section %s -Key %s -Value %s",
		cmdbuilder.EscapeString(filePath),
		cmdbuilder.EscapeString(section),
		cmdbuilder.EscapeString(key),
		cmdbuilder.EscapeString(value))
	return s.executeVoid(ctx, cmd)
}

// RemoveIniValue removes a value from an INI file.
func (s *Session) RemoveIniValue(filePath, section string, key ...string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTIniValue -FilePath %s -Section %s",
		cmdbuilder.EscapeString(filePath),
		cmdbuilder.EscapeString(section))
	if len(key) > 0 && key[0] != "" {
		cmd += fmt.Sprintf(" -Key %s", cmdbuilder.EscapeString(key[0]))
	}
	return s.executeVoid(ctx, cmd)
}

// GetIniSection reads an entire section from an INI file.
func (s *Session) GetIniSection(filePath, section string) (map[string]string, error) {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Get-ADTIniSection -FilePath %s -Section %s",
		cmdbuilder.EscapeString(filePath),
		cmdbuilder.EscapeString(section))
	data, err := s.execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := parser.ParseResponse(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetIniSection writes an entire section to an INI file.
func (s *Session) SetIniSection(filePath, section string, content map[string]string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	// Build a hashtable from the content map
	pairs := ""
	for k, v := range content {
		if pairs != "" {
			pairs += "; "
		}
		pairs += fmt.Sprintf("%s=%s", cmdbuilder.EscapeString(k), cmdbuilder.EscapeString(v))
	}
	cmd := fmt.Sprintf("Set-ADTIniSection -FilePath %s -Section %s -Content @{%s}",
		cmdbuilder.EscapeString(filePath),
		cmdbuilder.EscapeString(section),
		pairs)
	return s.executeVoid(ctx, cmd)
}

// RemoveIniSection removes an entire section from an INI file.
func (s *Session) RemoveIniSection(filePath, section string) error {
	ctx, cancel := s.getContext()
	defer cancel()
	cmd := fmt.Sprintf("Remove-ADTIniSection -FilePath %s -Section %s",
		cmdbuilder.EscapeString(filePath),
		cmdbuilder.EscapeString(section))
	return s.executeVoid(ctx, cmd)
}
