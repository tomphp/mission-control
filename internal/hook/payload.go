// Package hook parses the JSON payloads Claude Code sends on stdin to hook commands.
package hook

import (
	"encoding/json"
	"fmt"
	"io"
)

// EventName is one of the hook_event_name values Claude Code sends.
// Only the events relevant to session-state tracking are named as constants;
// any other value decodes fine but is treated as a no-op by the state machine.
type EventName string

const (
	SessionStart       EventName = "SessionStart"
	SessionEnd         EventName = "SessionEnd"
	UserPromptSubmit   EventName = "UserPromptSubmit"
	PreToolUse         EventName = "PreToolUse"
	PostToolUse        EventName = "PostToolUse"
	PostToolUseFailure EventName = "PostToolUseFailure"
	Notification       EventName = "Notification"
	Stop               EventName = "Stop"
	StopFailure        EventName = "StopFailure"
)

// Payload is the subset of the hook stdin JSON that mission-control cares about.
// Fields absent from a given event are simply left at their zero value.
type Payload struct {
	SessionID      string    `json:"session_id"`
	PromptID       string    `json:"prompt_id"`
	TranscriptPath string    `json:"transcript_path"`
	Cwd            string    `json:"cwd"`
	PermissionMode string    `json:"permission_mode"`
	HookEventName  EventName `json:"hook_event_name"`
	AgentID        string    `json:"agent_id"`
	AgentType      string    `json:"agent_type"`
}

// Parse decodes a hook stdin payload, tolerating unknown/extra fields since
// Claude Code's hook schema is larger than the fields mission-control models.
func Parse(r io.Reader) (Payload, error) {
	var p Payload
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return Payload{}, fmt.Errorf("hook: decode payload: %w", err)
	}
	if p.SessionID == "" {
		return Payload{}, fmt.Errorf("hook: missing session_id")
	}
	if p.HookEventName == "" {
		return Payload{}, fmt.Errorf("hook: missing hook_event_name")
	}
	if p.Cwd == "" {
		return Payload{}, fmt.Errorf("hook: missing cwd")
	}
	return p, nil
}
