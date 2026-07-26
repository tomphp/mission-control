package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tomoram/mission-control/internal/config"
)

func readSettings(t *testing.T, path string) config.RawSettings {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var s config.RawSettings
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return s
}

func hooksOf(t *testing.T, s config.RawSettings) map[string][]config.HookEntry {
	t.Helper()
	var hooks map[string][]config.HookEntry
	if err := json.Unmarshal(s["hooks"], &hooks); err != nil {
		t.Fatalf("parse hooks: %v", err)
	}
	return hooks
}

func TestRunInstallProjectCreatesSettingsWithHooks(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	if code := runInstall([]string{"--project"}); code != 0 {
		t.Fatalf("runInstall() = %d, want 0", code)
	}

	path := filepath.Join(tmp, ".claude", "settings.json")
	hooks := hooksOf(t, readSettings(t, path))
	if entries := hooks["SessionStart"]; len(entries) != 1 {
		t.Errorf("SessionStart entries = %d, want 1", len(entries))
	}
}

func TestRunInstallGlobalUsesHomeDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if code := runInstall(nil); code != 0 {
		t.Fatalf("runInstall() = %d, want 0", code)
	}

	path := filepath.Join(tmp, ".claude", "settings.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected settings.json at %s: %v", path, err)
	}
}

func TestRunInstallDryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	if code := runInstall([]string{"--project", "--dry-run"}); code != 0 {
		t.Fatalf("runInstall() = %d, want 0", code)
	}

	path := filepath.Join(tmp, ".claude", "settings.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected no settings.json to be written, stat err = %v", err)
	}
}

func TestRunInstallUnknownFlagReturnsUsageError(t *testing.T) {
	if code := runInstall([]string{"--nope"}); code != 2 {
		t.Fatalf("runInstall() = %d, want 2", code)
	}
}

// TestRunUninstallRemovesMissionControlHooks seeds a settings.json with a
// hook entry shaped exactly like one mission-control would have installed
// (rather than relying on runInstall + os.Executable(), since the test
// binary isn't named "mission-control" and so wouldn't round-trip through
// isMissionControlHook's basename check).
func TestRunUninstallRemovesMissionControlHooks(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	settingsDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `{
		"model": "sonnet",
		"hooks": {
			"SessionStart": [{"hooks": [{"type": "command", "command": "/opt/mission-control report"}]}],
			"Stop": [{"hooks": [{"type": "command", "command": "afplay /System/Library/Sounds/Glass.aiff"}]}]
		}
	}`
	settingsPath := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := runUninstall([]string{"--project"}); code != 0 {
		t.Fatalf("runUninstall() = %d, want 0", code)
	}

	hooks := hooksOf(t, readSettings(t, settingsPath))
	if entries := hooks["SessionStart"]; len(entries) != 0 {
		t.Errorf("SessionStart entries after uninstall = %d, want 0", len(entries))
	}
	if entries := hooks["Stop"]; len(entries) != 1 {
		t.Errorf("Stop entries after uninstall = %d, want 1 (non-mission-control hook should survive)", len(entries))
	}
}

func TestRunUninstallDryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	settingsDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `{"hooks": {"SessionStart": [{"hooks": [{"type": "command", "command": "/opt/mission-control report"}]}]}}`
	settingsPath := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := runUninstall([]string{"--project", "--dry-run"}); code != 0 {
		t.Fatalf("runUninstall() = %d, want 0", code)
	}

	hooks := hooksOf(t, readSettings(t, settingsPath))
	if entries := hooks["SessionStart"]; len(entries) != 1 {
		t.Errorf("SessionStart entries = %d, want 1 (dry-run must not write)", len(entries))
	}
}
