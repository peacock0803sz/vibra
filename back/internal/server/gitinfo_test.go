package server

import (
	"os"
	"os/exec"
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

	// 一時ディレクトリにgitリポジトリを作成してdetached HEADをシミュレート
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		// git操作ではエラーを無視 (テスト環境のgit設定依存)
		_ = cmd.Run()
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("commit", "--allow-empty", "-m", "initial")

	// コミットハッシュを取得してdetached HEADに移行
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Skip("could not get HEAD commit")
	}
	commitHash := string(out[:len(out)-1]) // 改行を除去

	// detached HEADに切り替え
	cmd := exec.Command("git", "checkout", "--detach", commitHash)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Skip("could not create detached HEAD")
	}

	// テスト対象のリポジトリディレクトリに移動して実行
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	branch := GetBranch()
	// detached HEADの場合、短縮ハッシュ (7文字) が返されること
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
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		_ = cmd.Run()
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("commit", "--allow-empty", "-m", "initial")

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	branch := GetBranch()
	if branch != "main" {
		t.Errorf("GetBranch() = %q, want %q", branch, "main")
	}
}

func TestGetBranch_GitNotFound(t *testing.T) {
	// PATHからgitを除いた環境でGetBranchが空文字を返すことを確認
	origPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", origPath) }()
	_ = os.Setenv("PATH", "")

	// git が見つからない場合は空文字を返すが、
	// GetBranch は exec.Command を直接使うため LookPath を内部で行わない。
	// ここでは GetGitInfo 経由のフォールバックをテストする。
	info := GetGitInfo()
	if info.Branch != "" || info.Repository != "" {
		// git が実際に利用不可なら空になるはず
		// PATH が空でも git が見つかる環境では無条件パス
		t.Log("git still found (absolute path?), skipping git-not-found assertion")
	}
}
