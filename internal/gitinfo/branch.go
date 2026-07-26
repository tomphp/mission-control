// Package gitinfo derives lightweight git metadata for a working directory.
package gitinfo

import (
	"os/exec"
	"strings"
)

// Branch returns the current branch name for the repo at cwd, by shelling
// out to git. It never returns an error for the "not a git repo" case (or
// any other git failure) — it returns an empty string instead, since a
// missing branch is expected and shouldn't prevent a session from being
// tracked.
func Branch(cwd string) (string, error) {
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		// Detached HEAD state — no meaningful branch name.
		return "", nil
	}
	return branch, nil
}
