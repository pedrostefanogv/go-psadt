//go:build windows

package parser

import (
	"encoding/json"
	"testing"
)

func TestParse_Success(t *testing.T) {
	raw := `{"Success": true, "Data": {"ExitCode": 0, "StdOut": "hello"}, "Error": null}`
	resp, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
	if resp.Error != nil {
		t.Error("expected nil Error")
	}
}

func TestParse_Error(t *testing.T) {
	raw := `{"Success": false, "Data": null, "Error": {"Message": "file not found", "Type": "System.IO.FileNotFoundException", "StackTrace": "at line 1"}}`
	resp, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if resp.Success {
		t.Error("expected Success=false")
	}
	if resp.Error == nil {
		t.Fatal("expected non-nil Error")
	}
	if resp.Error.Message != "file not found" {
		t.Errorf("unexpected error message: %s", resp.Error.Message)
	}
}

func TestParse_Empty(t *testing.T) {
	_, err := Parse([]byte{})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParse_Invalid(t *testing.T) {
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseInto_Struct(t *testing.T) {
	type result struct {
		ExitCode int    `json:"ExitCode"`
		StdOut   string `json:"StdOut"`
	}

	raw := json.RawMessage(`{"ExitCode": 0, "StdOut": "hello"}`)
	resp := &Response{
		Success: true,
		Data:    raw,
	}

	var r result
	err := ParseInto(resp, &r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got %d", r.ExitCode)
	}
	if r.StdOut != "hello" {
		t.Errorf("expected StdOut='hello', got '%s'", r.StdOut)
	}
}

func TestParseInto_ErrorResponse(t *testing.T) {
	resp := &Response{
		Success: false,
		Error: &ErrorDetail{
			Message: "something failed",
			Type:    "System.Exception",
		},
	}

	var r interface{}
	err := ParseInto(resp, &r)
	if err == nil {
		t.Fatal("expected error for failed response")
	}

	psErr, ok := IsPSADTError(err)
	if !ok {
		t.Fatal("expected PSADTError")
	}
	if psErr.Message != "something failed" {
		t.Errorf("unexpected message: %s", psErr.Message)
	}
}

func TestParseInto_NullData(t *testing.T) {
	resp := &Response{
		Success: true,
		Data:    json.RawMessage("null"),
	}

	var r map[string]interface{}
	err := ParseInto(resp, &r)
	if err != nil {
		t.Fatalf("unexpected error for null data: %v", err)
	}
}

func TestParseBool_True(t *testing.T) {
	raw := `{"Success": true, "Data": true, "Error": null}`
	result, err := ParseBool([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true")
	}
}

func TestParseBool_False(t *testing.T) {
	raw := `{"Success": true, "Data": false, "Error": null}`
	result, err := ParseBool([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false")
	}
}

func TestParseString(t *testing.T) {
	raw := `{"Success": true, "Data": "hello world", "Error": null}`
	result, err := ParseString([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}

func TestCheckSuccess_OK(t *testing.T) {
	raw := `{"Success": true, "Data": null, "Error": null}`
	err := CheckSuccess([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckSuccess_Fail(t *testing.T) {
	raw := `{"Success": false, "Data": null, "Error": {"Message": "boom", "Type": "", "StackTrace": ""}}`
	err := CheckSuccess([]byte(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPSADTError_IsRebootRequired(t *testing.T) {
	err := NewPSADTErrorWithCode("reboot needed", 3010)
	if !IsRebootRequired(err) {
		t.Error("expected IsRebootRequired to be true")
	}
}

func TestPSADTError_IsUserCancelled(t *testing.T) {
	err := NewPSADTErrorWithCode("user cancelled", 1602)
	if !IsUserCancelled(err) {
		t.Error("expected IsUserCancelled to be true")
	}
}
