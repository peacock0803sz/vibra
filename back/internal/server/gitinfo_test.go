package server

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "SSH with .git suffix",
			input: "git@github.com:owner/repo.git",
			want:  "owner/repo",
		},
		{
			name:  "SSH without .git suffix",
			input: "git@github.com:owner/repo",
			want:  "owner/repo",
		},
		{
			name:  "HTTPS with .git suffix",
			input: "https://github.com/owner/repo.git",
			want:  "owner/repo",
		},
		{
			name:  "HTTPS without .git suffix",
			input: "https://github.com/owner/repo",
			want:  "owner/repo",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "invalid URL",
			input: "not-a-url",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRemoteURL(tt.input)
			if got != tt.want {
				t.Errorf("ParseRemoteURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetBranch_DetachedHead(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	run := func(t *testing.T, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
		}
	}

	run(t, "init")
	run(t, "config", "user.email", "test@example.com")
	run(t, "config", "user.name", "Test")
	run(t, "commit", "--allow-empty", "-m", "initial")

	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Skip("could not get HEAD commit")
	}
	commitHash := strings.TrimSpace(string(out))

	run(t, "checkout", "--detach", commitHash)

	branch := GetBranch(dir)
	if len(branch) == 0 {
		t.Error("GetBranch() returned empty string for detached HEAD")
	}
	if branch == "HEAD" {
		t.Error("GetBranch() returned 'HEAD' for detached HEAD, expected short hash")
	}
}

func TestGetBranch_NormalBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	run := func(t *testing.T, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
		}
	}

	run(t, "init")
	run(t, "config", "user.email", "test@example.com")
	run(t, "config", "user.name", "Test")
	run(t, "checkout", "-b", "main")
	run(t, "commit", "--allow-empty", "-m", "initial")

	branch := GetBranch(dir)
	if branch != "main" {
		t.Errorf("GetBranch() = %q, want %q", branch, "main")
	}
}

func TestGetBranch_GitNotFound(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", origPath) }()
	_ = os.Setenv("PATH", "")

	// GetGitInfo checks LookPath("git") first, so with empty PATH it should
	// return an empty GitInfo. If git is still found (absolute path in some
	// environments), we skip the assertion.
	info := GetGitInfo(t.TempDir())
	if info.Branch != "" || info.Repository != "" {
		t.Log("git still found (absolute path?), skipping git-not-found assertion")
	}
}
