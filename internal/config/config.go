package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	APIURL   string `json:"api_url"`
	APIToken string `json:"api_token"`
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".canvas-cli")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("not configured. Run: canvas-cli configure")
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("corrupt config file: %w", err)
	}
	if cfg.APIURL == "" || cfg.APIToken == "" {
		return nil, fmt.Errorf("incomplete config. Run: canvas-cli configure")
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(ConfigDir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0600)
}

func RunSetup() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Canvas CLI Configuration ===")
	fmt.Println()
	fmt.Println("You'll need your Canvas URL and an API access token.")
	fmt.Println("To generate a token: Canvas → Account → Settings → New Access Token")
	fmt.Println()
	fmt.Println("Config will be stored in:", ConfigPath())
	fmt.Println()

	fmt.Print("Canvas URL (e.g. https://myschool.instructure.com): ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	url = strings.TrimRight(url, "/")
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		url = "https://" + url
	}

	fmt.Print("API Token: ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	if url == "" || token == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	cfg := &Config{
		APIURL:   url,
		APIToken: token,
	}

	if err := Save(cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("Configuration saved to", ConfigPath())
	return cfg, nil
}
