package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Config represents runtime configuration for GitHub Contribution City
type Config struct {
	GitHubToken    string `json:"github_token"`
	GitHubUsername string `json:"github_username"`
	Weeks          int    `json:"weeks"`
	MaxHeight      int    `json:"max_height"`
	Theme          string `json:"theme"`
	Port           int    `json:"port"`
	Host           string `json:"host"`
	OutputDir      string `json:"output_dir"`
	InputDir       string `json:"input_dir"`
}

// DefaultConfig returns default sensible parameters
func DefaultConfig() *Config {
	return &Config{
		GitHubToken:    "",
		GitHubUsername: "torvalds",
		Weeks:          53,
		MaxHeight:      14,
		Theme:          "github-dark",
		Port:           8088,
		Host:           "0.0.0.0",
		OutputDir:      "output",
		InputDir:       "input",
	}
}

// LoadConfig loads configuration from optional .env file and environment variables
func LoadConfig(envFile string) *Config {
	cfg := DefaultConfig()

	// Try reading envFile if exists
	if envFile != "" {
		_ = loadDotEnv(envFile)
	} else if _, err := os.Stat(".env"); err == nil {
		_ = loadDotEnv(".env")
	}

	if val := os.Getenv("GITHUB_TOKEN"); val != "" {
		cfg.GitHubToken = strings.TrimSpace(val)
	}
	if val := os.Getenv("GITHUB_USERNAME"); val != "" {
		cfg.GitHubUsername = strings.TrimSpace(val)
	}
	if val := os.Getenv("CITY_WEEKS"); val != "" {
		if w, err := strconv.Atoi(val); err == nil && w >= 1 && w <= 53 {
			cfg.Weeks = w
		}
	}
	if val := os.Getenv("CITY_MAX_HEIGHT"); val != "" {
		if h, err := strconv.Atoi(val); err == nil && h >= 2 && h <= 50 {
			cfg.MaxHeight = h
		}
	}
	if val := os.Getenv("CITY_THEME"); val != "" {
		cfg.Theme = strings.TrimSpace(val)
	}
	if val := os.Getenv("PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil && p > 0 {
			cfg.Port = p
		}
	}
	if val := os.Getenv("HOST"); val != "" {
		cfg.Host = strings.TrimSpace(val)
	}
	if val := os.Getenv("OUTPUT_DIR"); val != "" {
		cfg.OutputDir = strings.TrimSpace(val)
	}
	if val := os.Getenv("INPUT_DIR"); val != "" {
		cfg.InputDir = strings.TrimSpace(val)
	}

	return cfg
}

func loadDotEnv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
	return scanner.Err()
}
