package gitinfo_test

import (
	"os/exec"
	"testing"

	"github.com/tomoram/mission-control/internal/gitinfo"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func TestBranch_InGitRepo(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "commit", "--allow-empty", "-q", "-m", "init")

	branch, err := gitinfo.Branch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("Branch = %q, want %q", branch, "main")
	}
}

func TestBranch_AfterCheckoutNewBranch(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "commit", "--allow-empty", "-q", "-m", "init")
	runGit(t, dir, "checkout", "-q", "-b", "feature/foo")

	branch, err := gitinfo.Branch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "feature/foo" {
		t.Errorf("Branch = %q, want %q", branch, "feature/foo")
	}
}

func TestBranch_NotAGitRepo(t *testing.T) {
	dir := t.TempDir()

	branch, err := gitinfo.Branch(dir)
	if err != nil {
		t.Fatalf("expected no error for non-git dir, got: %v", err)
	}
	if branch != "" {
		t.Errorf("Branch = %q, want empty string", branch)
	}
}

func TestBranch_NonexistentDir(t *testing.T) {
	branch, err := gitinfo.Branch("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Fatalf("expected no error for nonexistent dir, got: %v", err)
	}
	if branch != "" {
		t.Errorf("Branch = %q, want empty string", branch)
	}
}
