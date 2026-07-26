// Package config safely merges mission-control's hook registration into a
// Claude Code settings.json without disturbing hooks or keys it doesn't
// own.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomoram/mission-control/internal/hook"
)

// RawSettings is a settings.json document with every top-level key kept as
// raw bytes except the ones this package explicitly rewrites ("hooks") —
// so unrelated keys round-trip byte-for-byte.
type RawSettings map[string]json.RawMessage

// HookEntry mirrors one entry in a settings.json hooks.<Event> array.
type HookEntry struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []HookCommand `json:"hooks"`
}

// HookCommand is a single hook command object. It's modeled as a raw field
// map (not a fixed struct) so fields this package doesn't know about
// (timeout, async, statusMessage, ...) survive a decode/re-encode cycle
// untouched.
type HookCommand map[string]json.RawMessage

// Command returns the "command" field, if present.
func (c HookCommand) Command() (string, error) {
	raw, ok := c["command"]
	if !ok {
		return "", nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", fmt.Errorf("config: decode command field: %w", err)
	}
	return s, nil
}

func newCommandHook(command string) HookCommand {
	typeJSON, _ := json.Marshal("command")
	cmdJSON, _ := json.Marshal(command)
	return HookCommand{"type": typeJSON, "command": cmdJSON}
}

// targetEvents are the hook events mission-control registers itself against
// — exactly the events internal/session's state machine and SessionEnd
// handling care about.
var targetEvents = []hook.EventName{
	hook.SessionStart,
	hook.SessionEnd,
	hook.UserPromptSubmit,
	hook.PreToolUse,
	hook.PostToolUse,
	hook.PostToolUseFailure,
	hook.Notification,
	hook.Stop,
	hook.StopFailure,
}

// reportCommand is the exact command string mission-control registers, and
// the sentinel used to identify (and later remove) its own entries.
func reportCommand(binaryPath string) string {
	return binaryPath + " report"
}

func isMissionControlHook(c HookCommand) bool {
	cmd, err := c.Command()
	if err != nil || cmd == "" {
		return false
	}
	fields := strings.Fields(cmd)
	if len(fields) != 2 || fields[1] != "report" {
		return false
	}
	return filepath.Base(fields[0]) == "mission-control"
}

func decodeHooks(s RawSettings) (map[string][]HookEntry, error) {
	hooks := map[string][]HookEntry{}
	raw, ok := s["hooks"]
	if !ok {
		return hooks, nil
	}
	if err := json.Unmarshal(raw, &hooks); err != nil {
		return nil, fmt.Errorf("config: decode hooks: %w", err)
	}
	return hooks, nil
}

func encodeHooks(s RawSettings, hooks map[string][]HookEntry) (RawSettings, error) {
	raw, err := json.Marshal(hooks)
	if err != nil {
		return nil, fmt.Errorf("config: encode hooks: %w", err)
	}
	out := RawSettings{}
	for k, v := range s {
		out[k] = v
	}
	out["hooks"] = raw
	return out, nil
}

// Install returns a copy of s with mission-control's hook registered
// against every target event. It never mutates an existing HookEntry —
// it only appends a new one — so hooks the user already configured are
// left exactly as they were. Calling Install again is a no-op for events
// where it's already registered.
func Install(s RawSettings, binaryPath string) (RawSettings, error) {
	hooks, err := decodeHooks(s)
	if err != nil {
		return nil, err
	}

	command := reportCommand(binaryPath)
	for _, event := range targetEvents {
		key := string(event)
		entries := hooks[key]

		alreadyInstalled := false
		for _, e := range entries {
			for _, c := range e.Hooks {
				if isMissionControlHook(c) {
					alreadyInstalled = true
				}
			}
		}
		if !alreadyInstalled {
			entries = append(entries, HookEntry{Hooks: []HookCommand{newCommandHook(command)}})
		}
		hooks[key] = entries
	}

	return encodeHooks(s, hooks)
}

// Uninstall returns a copy of s with every HookEntry that is solely a
// mission-control hook removed. Entries mixing a mission-control hook with
// other hooks (which mission-control itself never creates) and all other
// entries are left untouched.
func Uninstall(s RawSettings, binaryPath string) (RawSettings, error) {
	hooks, err := decodeHooks(s)
	if err != nil {
		return nil, err
	}

	for key, entries := range hooks {
		filtered := make([]HookEntry, 0, len(entries))
		for _, e := range entries {
			if len(e.Hooks) == 1 && isMissionControlHook(e.Hooks[0]) {
				continue
			}
			filtered = append(filtered, e)
		}
		hooks[key] = filtered
	}

	return encodeHooks(s, hooks)
}

// Load reads and parses a settings.json file. A missing file is not an
// error — it returns empty settings, since installing hooks is often the
// first thing that creates ~/.claude/settings.json.
func Load(path string) (RawSettings, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RawSettings{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var s RawSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return s, nil
}

// Save writes settings.json atomically (write to a temp file in the same
// directory, then rename over the target).
func Save(path string, s RawSettings) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("config: encode settings: %w", err)
	}
	data := buf.Bytes()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("config: create %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".settings-*.json.tmp")
	if err != nil {
		return fmt.Errorf("config: create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("config: write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("config: close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("config: rename temp file: %w", err)
	}
	return nil
}
