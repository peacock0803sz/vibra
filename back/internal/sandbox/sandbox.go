package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds sandbox configuration for directory allowlists and env var filtering.
type Config struct {
	AllowedDirs []string // Allowed working directories (VIBRA_ALLOWED_DIRS)
	AllowedEnvs []string // Allowed environment variable names (VIBRA_ALLOWED_ENVS)
}

// NewConfigFromEnv creates a sandbox Config from environment variables.
func NewConfigFromEnv() *Config {
	cfg := &Config{}

	if dirs := os.Getenv("VIBRA_ALLOWED_DIRS"); dirs != "" {
		for _, d := range strings.Split(dirs, ",") {
			if trimmed := strings.TrimSpace(d); trimmed != "" {
				cfg.AllowedDirs = append(cfg.AllowedDirs, trimmed)
			}
		}
	}

	if envs := os.Getenv("VIBRA_ALLOWED_ENVS"); envs != "" {
		for _, e := range strings.Split(envs, ",") {
			if trimmed := strings.TrimSpace(e); trimmed != "" {
				cfg.AllowedEnvs = append(cfg.AllowedEnvs, trimmed)
			}
		}
	} else {
		// Default: only pass through API keys.
		cfg.AllowedEnvs = []string{
			"ANTHROPIC_API_KEY",
			"OPENAI_API_KEY",
			"GEMINI_API_KEY",
		}
	}

	return cfg
}

// ValidateDir checks that the working directory is within the allowlist.
func (c *Config) ValidateDir(dir string) error {
	if len(c.AllowedDirs) == 0 {
		return fmt.Errorf("no allowed directories configured (set VIBRA_ALLOWED_DIRS)")
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory path: %w", err)
	}
	abs = filepath.Clean(abs)

	for _, allowed := range c.AllowedDirs {
		allowedAbs, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		allowedAbs = filepath.Clean(allowedAbs)

		if abs == allowedAbs || strings.HasPrefix(abs, allowedAbs+string(filepath.Separator)) {
			return nil
		}
	}

	return fmt.Errorf("directory %q not in allowlist", dir)
}

// FilterEnv filters environment variables, keeping only keys in the allowlist.
// Returns a slice of "KEY=VALUE" strings.
func (c *Config) FilterEnv(env map[string]string) []string {
	allowed := make(map[string]bool, len(c.AllowedEnvs))
	for _, key := range c.AllowedEnvs {
		allowed[key] = true
	}

	var result []string
	for k, v := range env {
		if allowed[k] {
			result = append(result, k+"="+v)
		}
	}
	return result
}
