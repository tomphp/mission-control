package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tomoram/mission-control/internal/config"
)

// realFixture mirrors an actual ~/.claude/settings.json on a developer's
// machine: existing Stop/Notification sound-effect hooks, some empty hook
// arrays, and unrelated top-level keys — all of which must survive an
// install/uninstall cycle untouched.
const realFixture = `{
  "model": "sonnet",
  "hooks": {
    "PreToolUse": [],
    "SubagentStop": [],
    "PostToolUse": [],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "afplay /System/Library/Sounds/Glass.aiff 2>/dev/null || true"
          }
        ]
      }
    ],
    "Notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "afplay /System/Library/Sounds/Ping.aiff 2>/dev/null || true"
          }
        ]
      }
    ]
  },
  "enabledPlugins": {
    "swift-lsp@claude-plugins-official": true
  },
  "tui": "fullscreen",
  "editorMode": "vim"
}`

const binaryPath = "/usr/local/bin/mission-control"

func loadFixture(t *testing.T) config.RawSettings {
	t.Helper()
	var s config.RawSettings
	if err := json.Unmarshal([]byte(realFixture), &s); err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}
	return s
}

func hookEntries(t *testing.T, s config.RawSettings, event string) []config.HookEntry {
	t.Helper()
	var hooks map[string][]config.HookEntry
	if err := json.Unmarshal(s["hooks"], &hooks); err != nil {
		t.Fatalf("failed to parse hooks: %v", err)
	}
	return hooks[event]
}

func commandOf(t *testing.T, c config.HookCommand) string {
	t.Helper()
	cmd, err := c.Command()
	if err != nil {
		t.Fatalf("HookCommand.Command(): %v", err)
	}
	return cmd
}

func TestInstall_AppendsToExistingEntry_WithoutMutatingIt(t *testing.T) {
	s := loadFixture(t)

	got, err := config.Install(s, binaryPath)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	stop := hookEntries(t, got, "Stop")
	if len(stop) != 2 {
		t.Fatalf("Stop entries = %d, want 2", len(stop))
	}

	var sawOriginal, sawOurs bool
	for _, e := range stop {
		if len(e.Hooks) != 1 {
			t.Fatalf("Stop entry has %d hooks, want 1", len(e.Hooks))
		}
		cmd := commandOf(t, e.Hooks[0])
		switch cmd {
		case "afplay /System/Library/Sounds/Glass.aiff 2>/dev/null || true":
			sawOriginal = true
		case binaryPath + " report":
			sawOurs = true
		}
	}
	if !sawOriginal {
		t.Error("original afplay Stop hook was not preserved byte-identical")
	}
	if !sawOurs {
		t.Error("mission-control Stop hook was not installed")
	}
}

func TestInstall_PreservesNotificationHook(t *testing.T) {
	s := loadFixture(t)
	got, err := config.Install(s, binaryPath)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	notif := hookEntries(t, got, "Notification")
	if len(notif) != 2 {
		t.Fatalf("Notification entries = %d, want 2", len(notif))
	}
	found := false
	for _, e := range notif {
		if commandOf(t, e.Hooks[0]) == "afplay /System/Library/Sounds/Ping.aiff 2>/dev/null || true" {
			found = true
		}
	}
	if !found {
		t.Error("original afplay Notification hook was not preserved")
	}
}

func TestInstall_AddsNewEventKeys(t *testing.T) {
	s := loadFixture(t)
	got, err := config.Install(s, binaryPath)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	for _, event := range []string{"SessionStart", "SessionEnd", "UserPromptSubmit", "PostToolUseFailure", "StopFailure"} {
		entries := hookEntries(t, got, event)
		if len(entries) != 1 {
			t.Errorf("%s entries = %d, want 1", event, len(entries))
			continue
		}
		if commandOf(t, entries[0].Hooks[0]) != binaryPath+" report" {
			t.Errorf("%s command = %q, want %q", event, commandOf(t, entries[0].Hooks[0]), binaryPath+" report")
		}
	}
}

func TestInstall_LeavesUnrelatedEventsAlone(t *testing.T) {
	s := loadFixture(t)
	got, err := config.Install(s, binaryPath)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	subagentStop := hookEntries(t, got, "SubagentStop")
	if len(subagentStop) != 0 {
		t.Errorf("SubagentStop entries = %d, want 0 (not a mission-control event)", len(subagentStop))
	}
}

func TestInstall_PreservesUnrelatedTopLevelKeys(t *testing.T) {
	s := loadFixture(t)
	got, err := config.Install(s, binaryPath)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	for _, key := range []string{"model", "enabledPlugins", "tui", "editorMode"} {
		want := s[key]
		have := got[key]
		if string(want) != string(have) {
			t.Errorf("key %q changed: want %s, got %s", key, want, have)
		}
	}
}

func TestInstall_IsIdempotent(t *testing.T) {
	s := loadFixture(t)

	once, err := config.Install(s, binaryPath)
	if err != nil {
		t.Fatalf("Install (1st): %v", err)
	}
	twice, err := config.Install(once, binaryPath)
	if err != nil {
		t.Fatalf("Install (2nd): %v", err)
	}

	for _, event := range []string{"Stop", "Notification", "PreToolUse", "PostToolUse", "SessionStart"} {
		gotOnce := hookEntries(t, once, event)
		gotTwice := hookEntries(t, twice, event)
		if len(gotOnce) != len(gotTwice) {
			t.Errorf("%s: entries grew from %d to %d on repeated Install", event, len(gotOnce), len(gotTwice))
		}
	}
}

// TestInstall_IsIdempotent_WindowsBinary guards against a regression where
// isMissionControlHook's basename check didn't strip the ".exe" suffix
// Windows builds carry (see release.yml, which names the artifact
// mission-control.exe), so a repeated install would never recognize its own
// hook and kept appending duplicates. CI only runs this suite on Linux,
// where path/filepath treats backslash as an ordinary character rather than
// a separator, so a real "C:\..." path wouldn't exercise filepath.Base the
// way it does on a Windows build; forward slashes are a separator on both
// and get the .exe-stripping logic under test either way.
func TestInstall_IsIdempotent_WindowsBinary(t *testing.T) {
	s := loadFixture(t)
	winBinary := "C:/Users/tom/mission-control.exe"

	once, err := config.Install(s, winBinary)
	if err != nil {
		t.Fatalf("Install (1st): %v", err)
	}
	twice, err := config.Install(once, winBinary)
	if err != nil {
		t.Fatalf("Install (2nd): %v", err)
	}

	for _, event := range []string{"Stop", "Notification", "PreToolUse", "PostToolUse", "SessionStart"} {
		gotOnce := hookEntries(t, once, event)
		gotTwice := hookEntries(t, twice, event)
		if len(gotOnce) != len(gotTwice) {
			t.Errorf("%s: entries grew from %d to %d on repeated Install with a .exe binary", event, len(gotOnce), len(gotTwice))
		}
	}
}

func TestUninstall_RemovesOnlyMissionControlEntries(t *testing.T) {
	s := loadFixture(t)
	installed, err := config.Install(s, binaryPath)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	uninstalled, err := config.Uninstall(installed, binaryPath)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	stop := hookEntries(t, uninstalled, "Stop")
	if len(stop) != 1 {
		t.Fatalf("Stop entries after uninstall = %d, want 1", len(stop))
	}
	if commandOf(t, stop[0].Hooks[0]) != "afplay /System/Library/Sounds/Glass.aiff 2>/dev/null || true" {
		t.Errorf("remaining Stop hook = %q, want original afplay command", commandOf(t, stop[0].Hooks[0]))
	}

	notif := hookEntries(t, uninstalled, "Notification")
	if len(notif) != 1 {
		t.Fatalf("Notification entries after uninstall = %d, want 1", len(notif))
	}

	for _, event := range []string{"SessionStart", "SessionEnd", "UserPromptSubmit", "PreToolUse", "PostToolUse", "PostToolUseFailure", "StopFailure"} {
		entries := hookEntries(t, uninstalled, event)
		if len(entries) != 0 {
			t.Errorf("%s entries after uninstall = %d, want 0", event, len(entries))
		}
	}

	for _, key := range []string{"model", "enabledPlugins", "tui", "editorMode"} {
		if string(s[key]) != string(uninstalled[key]) {
			t.Errorf("key %q changed after uninstall: want %s, got %s", key, s[key], uninstalled[key])
		}
	}
}

// TestUninstall_RemovesEntries_WindowsBinary is the Uninstall counterpart to
// TestInstall_IsIdempotent_WindowsBinary: the same .exe-suffix bug also made
// Uninstall unable to recognize its own hooks, so it silently removed
// nothing.
func TestUninstall_RemovesEntries_WindowsBinary(t *testing.T) {
	s := loadFixture(t)
	winBinary := "C:/Users/tom/mission-control.exe"

	installed, err := config.Install(s, winBinary)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	uninstalled, err := config.Uninstall(installed, winBinary)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	for _, event := range []string{"SessionStart", "SessionEnd", "UserPromptSubmit", "PostToolUseFailure", "StopFailure"} {
		entries := hookEntries(t, uninstalled, event)
		if len(entries) != 0 {
			t.Errorf("%s entries after uninstall = %d, want 0", event, len(entries))
		}
	}

	stop := hookEntries(t, uninstalled, "Stop")
	if len(stop) != 1 {
		t.Fatalf("Stop entries after uninstall = %d, want 1", len(stop))
	}
}

func TestLoad_MissingFile_ReturnsEmptySettingsNoError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	s, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if len(s) != 0 {
		t.Errorf("settings = %v, want empty", s)
	}
}

func TestLoad_Save_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	if err := os.WriteFile(path, []byte(realFixture), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	installed, err := config.Install(loaded, binaryPath)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if err := config.Save(path, installed); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("reload after save: %v", err)
	}
	stop := hookEntries(t, reloaded, "Stop")
	if len(stop) != 2 {
		t.Fatalf("Stop entries after save+reload = %d, want 2", len(stop))
	}
}

func TestSave_DoesNotHTMLEscapeCommandStrings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	s := loadFixture(t)
	if err := config.Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if !strings.Contains(string(raw), "2>/dev/null") {
		t.Errorf("saved file HTML-escaped '>' in command string; got:\n%s", raw)
	}
}
