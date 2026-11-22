package llm

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"encoding/json/jsontext"
	json "encoding/json/v2"
)

// TestRequestMarshalGolden checks that a Request encodes to exactly the JSON in
// testdata/request.golden.json. The golden file is the specification: it is
// what the Claude Messages API expects, and the neutral format every adapter
// translates from.
func TestRequestMarshalGolden(t *testing.T) {
	req := Request{
		Model:     "claude-haiku-5-5",
		MaxTokens: 1024,
		System:    "You are Ops Agent.",
		Messages: []Message{
			{Role: RoleUser, Content: []ContentBlock{
				{Type: BlockText, Text: "What time is it in Dhaka?"},
			}},
			{Role: RoleAssistant, Content: []ContentBlock{
				{Type: BlockToolUse, ID: "call_1", Name: "get_current_time", Input: jsontext.Value(`{}`)},
			}},
			{Role: RoleUser, Content: []ContentBlock{
				{Type: BlockToolResult, ToolUseID: "call_1", Content: "2026-11-22T20:45:00+06:00"},
			}},
		},
		Tools: []Tool{{
			Name:        "get_current_time",
			Description: "Returns the current date and time in Asia/Dhaka.",
			InputSchema: jsontext.Value(`{"type":"object","properties":{}}`),
		}},
	}

	got, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := readFile(t, "request.golden.json")

	if !sameJSON(t, got, want) {
		t.Errorf("request JSON mismatch\n got: %s\nwant: %s", got, want)
	}
}

// TestRequestOmitsEmptyOptionalFields checks that optional fields disappear
// from the JSON when unset. APIs reject or misread "system": "" and "tools": null.
func TestRequestOmitsEmptyOptionalFields(t *testing.T) {
	req := Request{
		Model:     "qwen3:14b",
		MaxTokens: 256,
		Messages:  []Message{{Role: RoleUser, Content: []ContentBlock{{Type: BlockText, Text: "hi"}}}},
	}
	got, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"model":"qwen3:14b","max_tokens":256,"messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]}`
	if !sameJSON(t, got, []byte(want)) {
		t.Errorf("optional fields not omitted\n got: %s\nwant: %s", got, want)
	}
}

// TestResponseUnmarshal decodes a real-shaped API response. It also contains
// fields we don't model ("id", "model", ...), which must be ignored, not fail.
func TestResponseUnmarshal(t *testing.T) {
	var resp Response
	if err := json.Unmarshal(readFile(t, "response.json"), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp.StopReason != StopToolUse {
		t.Errorf("StopReason = %q, want %q", resp.StopReason, StopToolUse)
	}
	if resp.Usage.InputTokens != 137 || resp.Usage.OutputTokens != 42 {
		t.Errorf("Usage = %+v, want {InputTokens:137 OutputTokens:42}", resp.Usage)
	}
	if len(resp.Content) != 2 {
		t.Fatalf("len(Content) = %d, want 2", len(resp.Content))
	}
	call := resp.Content[1]
	if call.Type != BlockToolUse || call.ID != "toolu_01" || call.Name != "get_current_time" {
		t.Errorf("tool_use block = %+v", call)
	}
	if !sameJSON(t, call.Input, []byte(`{"timezone":"Asia/Dhaka"}`)) {
		t.Errorf("Input = %s", call.Input)
	}
}

func TestResponseText(t *testing.T) {
	tests := []struct {
		name    string
		content []ContentBlock
		want    string
	}{
		{"no blocks", nil, ""},
		{"single text", []ContentBlock{{Type: BlockText, Text: "hello"}}, "hello"},
		{"joins text blocks", []ContentBlock{
			{Type: BlockText, Text: "Today is "},
			{Type: BlockText, Text: "Sunday."},
		}, "Today is Sunday."},
		{"skips tool_use first block", []ContentBlock{
			{Type: BlockToolUse, ID: "t1", Name: "get_current_time"},
			{Type: BlockText, Text: "checking"},
		}, "checking"},
		{"only tool_use", []ContentBlock{{Type: BlockToolUse, ID: "t1"}}, ""},
		{"bangla text survives", []ContentBlock{{Type: BlockText, Text: "আজ রবিবার"}}, "আজ রবিবার"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Response{Content: tt.content}
			if got := r.Text(); got != tt.want {
				t.Errorf("Text() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- helpers ---------------------------------------------------------------

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	return b
}

// sameJSON reports whether a and b are the same JSON, ignoring whitespace and
// object key order (RFC 8785 canonical form).
func sameJSON(t *testing.T, a, b []byte) bool {
	t.Helper()
	ca, cb := jsontext.Value(bytes.Clone(a)), jsontext.Value(bytes.Clone(b))
	if err := ca.Canonicalize(); err != nil {
		t.Fatalf("canonicalize %s: %v", a, err)
	}
	if err := cb.Canonicalize(); err != nil {
		t.Fatalf("canonicalize %s: %v", b, err)
	}
	return bytes.Equal(ca, cb)
}
