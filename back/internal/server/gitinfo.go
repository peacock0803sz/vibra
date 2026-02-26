package server

import (
	"os/exec"
	"strings"
)

// GitInfo holds repository context gathered from the local git environment.
type GitInfo struct {
	Repository string // "owner/repo" form; empty if git unavailable or no remote
	Branch     string // branch name or short commit hash (detached HEAD); empty if not a git repo
}

// GetGitInfo returns git repository context for the given directory.
// Returns an empty GitInfo if git is not installed or the directory is not a git repo.
func GetGitInfo(dir string) GitInfo {
	if _, err := exec.LookPath("git"); err != nil {
		return GitInfo{}
	}
	remote := getRemoteURL(dir)
	return GitInfo{
		Repository: ParseRemoteURL(remote),
		Branch:     GetBranch(dir),
	}
}

// GetBranch returns the current branch name for the given directory.
// Returns a short commit hash if the HEAD is detached.
// Returns an empty string if git is unavailable or the directory is not a git repo.
func GetBranch(dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		// detached HEAD: substitute with short commit hash
		hash, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(hash))
	}
	return branch
}

// ParseRemoteURL converts an SSH or HTTPS git remote URL to "owner/repo" form.
// Returns an empty string if the URL cannot be parsed.
func ParseRemoteURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	// SSH形式: git@github.com:owner/repo.git
	if strings.HasPrefix(rawURL, "git@") {
		parts := strings.SplitN(rawURL, ":", 2)
		if len(parts) != 2 {
			return ""
		}
		return strings.TrimSuffix(parts[1], ".git")
	}
	// HTTPS形式: https://github.com/owner/repo.git
	if strings.Contains(rawURL, "://") {
		parts := strings.SplitN(rawURL, "://", 2)
		if len(parts) != 2 {
			return ""
		}
		// ホスト部分を除去: "github.com/owner/repo.git" -> "owner/repo"
		_, path, ok := strings.Cut(parts[1], "/")
		if !ok {
			return ""
		}
		return strings.TrimSuffix(path, ".git")
	}
	return ""
}

// getRemoteURL returns the URL of the "origin" remote, or empty string if unavailable.
func getRemoteURL(dir string) string {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
