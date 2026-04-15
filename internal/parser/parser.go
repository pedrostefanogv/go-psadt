//go:build windows

package parser

import (
	"encoding/json"
	"fmt"
)

// Response represents the standard JSON wrapper returned by every PowerShell command.
type Response struct {
	Success bool            `json:"Success"`
	Data    json.RawMessage `json:"Data"`
	Error   *ErrorDetail    `json:"Error"`
}

// ErrorDetail contains error information from a PowerShell command.
type ErrorDetail struct {
	Message    string `json:"Message"`
	Type       string `json:"Type"`
	StackTrace string `json:"StackTrace"`
}

// Parse parses a raw JSON response from PowerShell into the Response wrapper.
func Parse(data []byte) (*Response, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from PowerShell")
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse PowerShell response: %w (raw: %s)", err, truncate(string(data), 200))
	}

	return &resp, nil
}

// ParseInto parses a Response's Data field into a target struct.
func ParseInto(resp *Response, target interface{}) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}

	if !resp.Success {
		return NewPSADTError(resp.Error)
	}

	if target == nil {
		return nil
	}

	// Handle null data
	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return nil
	}

	if err := json.Unmarshal(resp.Data, target); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w (raw: %s)", err, truncate(string(resp.Data), 200))
	}

	return nil
}

// ParseResponse is a convenience function that combines Parse + ParseInto.
func ParseResponse(data []byte, target interface{}) error {
	resp, err := Parse(data)
	if err != nil {
		return err
	}
	return ParseInto(resp, target)
}

// ParseBool parses a Response that returns a boolean value.
func ParseBool(data []byte) (bool, error) {
	resp, err := Parse(data)
	if err != nil {
		return false, err
	}

	if !resp.Success {
		return false, NewPSADTError(resp.Error)
	}

	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return false, nil
	}

	var result bool
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return false, fmt.Errorf("failed to parse bool response: %w", err)
	}

	return result, nil
}

// ParseString parses a Response that returns a string value.
func ParseString(data []byte) (string, error) {
	resp, err := Parse(data)
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", NewPSADTError(resp.Error)
	}

	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return "", nil
	}

	var result string
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("failed to parse string response: %w", err)
	}

	return result, nil
}

// ParseUint64 parses a Response that returns a uint64 value.
func ParseUint64(data []byte) (uint64, error) {
	resp, err := Parse(data)
	if err != nil {
		return 0, err
	}

	if !resp.Success {
		return 0, NewPSADTError(resp.Error)
	}

	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return 0, nil
	}

	var result uint64
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return 0, fmt.Errorf("failed to parse uint64 response: %w", err)
	}

	return result, nil
}

// CheckSuccess parses a response and only checks for success (no data extraction).
func CheckSuccess(data []byte) error {
	resp, err := Parse(data)
	if err != nil {
		return err
	}

	if !resp.Success {
		return NewPSADTError(resp.Error)
	}

	return nil
}

// truncate truncates a string to maxLen characters for error messages.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
