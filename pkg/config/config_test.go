package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Weeks != 53 {
		t.Errorf("expected weeks 53, got %d", cfg.Weeks)
	}
	if cfg.MaxHeight != 14 {
		t.Errorf("expected max height 14, got %d", cfg.MaxHeight)
	}
	if cfg.Port != 8088 {
		t.Errorf("expected port 8088, got %d", cfg.Port)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("GITHUB_USERNAME", "octocat")
	os.Setenv("CITY_WEEKS", "26")
	os.Setenv("PORT", "9090")
	defer func() {
		os.Unsetenv("GITHUB_USERNAME")
		os.Unsetenv("CITY_WEEKS")
		os.Unsetenv("PORT")
	}()

	cfg := LoadConfig("")
	if cfg.GitHubUsername != "octocat" {
		t.Errorf("expected octocat, got %s", cfg.GitHubUsername)
	}
	if cfg.Weeks != 26 {
		t.Errorf("expected weeks 26, got %d", cfg.Weeks)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
}

func TestLoadDotEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.test")
	content := "GITHUB_USERNAME=tester123\nCITY_THEME=cyberpunk\n# comment\nCITY_MAX_HEIGHT=20\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write tmp env: %v", err)
	}

	cfg := LoadConfig(envPath)
	if cfg.Theme != "cyberpunk" {
		t.Errorf("expected cyberpunk, got %s", cfg.Theme)
	}
}
