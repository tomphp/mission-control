package config

import (
	"os"
	"path/filepath"
)

// GlobalSettingsPath returns ~/.claude/settings.json.
func GlobalSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// ProjectSettingsPath returns ./.claude/settings.json relative to the
// current working directory.
func ProjectSettingsPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".claude", "settings.json"), nil
}
