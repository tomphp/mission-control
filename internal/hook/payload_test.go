package hook_test

import (
	"strings"
	"testing"

	"github.com/tomoram/mission-control/internal/hook"
)

func TestParse_ValidPayload(t *testing.T) {
	body := `{
		"session_id": "sess-123",
		"prompt_id": "prompt-abc",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd": "/Users/tom/project",
		"permission_mode": "default",
		"hook_event_name": "PreToolUse",
		"agent_id": "",
		"agent_type": ""
	}`

	p, err := hook.Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.SessionID != "sess-123" {
		t.Errorf("SessionID = %q, want %q", p.SessionID, "sess-123")
	}
	if p.Cwd != "/Users/tom/project" {
		t.Errorf("Cwd = %q, want %q", p.Cwd, "/Users/tom/project")
	}
	if p.HookEventName != hook.PreToolUse {
		t.Errorf("HookEventName = %q, want %q", p.HookEventName, hook.PreToolUse)
	}
	if p.TranscriptPath != "/tmp/transcript.jsonl" {
		t.Errorf("TranscriptPath = %q, want %q", p.TranscriptPath, "/tmp/transcript.jsonl")
	}
	if p.PermissionMode != "default" {
		t.Errorf("PermissionMode = %q, want %q", p.PermissionMode, "default")
	}
}

func TestParse_UnknownExtraFields_DoNotError(t *testing.T) {
	body := `{
		"session_id": "sess-123",
		"cwd": "/tmp",
		"hook_event_name": "PostToolUse",
		"tool_name": "Bash",
		"tool_input": {"command": "ls"},
		"tool_use_id": "abc",
		"some_future_field": {"nested": true}
	}`

	p, err := hook.Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error for payload with unknown fields: %v", err)
	}
	if p.SessionID != "sess-123" {
		t.Errorf("SessionID = %q, want %q", p.SessionID, "sess-123")
	}
}

func TestParse_MalformedJSON(t *testing.T) {
	_, err := hook.Parse(strings.NewReader(`{not json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestParse_MissingSessionID(t *testing.T) {
	body := `{"cwd": "/tmp", "hook_event_name": "Stop"}`
	_, err := hook.Parse(strings.NewReader(body))
	if err == nil {
		t.Fatal("expected error for missing session_id, got nil")
	}
}

func TestParse_MissingHookEventName(t *testing.T) {
	body := `{"session_id": "sess-1", "cwd": "/tmp"}`
	_, err := hook.Parse(strings.NewReader(body))
	if err == nil {
		t.Fatal("expected error for missing hook_event_name, got nil")
	}
}

func TestParse_MissingCwd(t *testing.T) {
	body := `{"session_id": "sess-1", "hook_event_name": "Stop"}`
	_, err := hook.Parse(strings.NewReader(body))
	if err == nil {
		t.Fatal("expected error for missing cwd, got nil")
	}
}
