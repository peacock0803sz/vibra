package sandbox

import (
	"sort"
	"strings"
	"testing"
)

func TestValidateDir_Allowed(t *testing.T) {
	cfg := &Config{
		AllowedDirs: []string{"/home/user/projects", "/tmp/work"},
	}

	tests := []struct {
		dir     string
		wantErr bool
	}{
		{"/home/user/projects", false},
		{"/home/user/projects/myapp", false},
		{"/home/user/projects/myapp/src", false},
		{"/tmp/work", false},
		{"/tmp/work/test", false},
		{"/home/user/secret", true},
		{"/etc/passwd", true},
		{"/", true},
		{"/home/user/projectsevil", true}, // prevent path traversal
	}

	for _, tt := range tests {
		err := cfg.ValidateDir(tt.dir)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateDir(%q): got err=%v, wantErr=%v", tt.dir, err, tt.wantErr)
		}
	}
}

func TestValidateDir_NoAllowedDirs(t *testing.T) {
	cfg := &Config{}

	err := cfg.ValidateDir("/any/path")
	if err == nil {
		t.Fatal("expected error when no allowed dirs configured")
	}
}

func TestFilterEnv(t *testing.T) {
	cfg := &Config{
		AllowedEnvs: []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY"},
	}

	env := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-xxx",
		"OPENAI_API_KEY":    "sk-xxx",
		"HOME":              "/root",
		"SECRET":            "password",
	}

	result := cfg.FilterEnv(env)
	sort.Strings(result)

	if len(result) != 2 {
		t.Fatalf("expected 2 env vars, got %d: %v", len(result), result)
	}

	keys := make(map[string]bool)
	for _, e := range result {
		parts := strings.SplitN(e, "=", 2)
		keys[parts[0]] = true
	}

	if !keys["ANTHROPIC_API_KEY"] {
		t.Error("ANTHROPIC_API_KEY should be allowed")
	}
	if !keys["OPENAI_API_KEY"] {
		t.Error("OPENAI_API_KEY should be allowed")
	}
	if keys["HOME"] {
		t.Error("HOME should be filtered out")
	}
	if keys["SECRET"] {
		t.Error("SECRET should be filtered out")
	}
}

func TestFilterEnv_Empty(t *testing.T) {
	cfg := &Config{
		AllowedEnvs: []string{"ALLOWED_KEY"},
	}

	result := cfg.FilterEnv(map[string]string{})
	if len(result) != 0 {
		t.Errorf("expected 0 env vars, got %d", len(result))
	}
}

func TestFilterEnv_NoMatch(t *testing.T) {
	cfg := &Config{
		AllowedEnvs: []string{"ALLOWED_KEY"},
	}

	result := cfg.FilterEnv(map[string]string{
		"OTHER_KEY": "value",
	})
	if len(result) != 0 {
		t.Errorf("expected 0 env vars, got %d", len(result))
	}
}
