package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns the Outlook config directory (~/.outlook-mcp or OUTLOOK_CONFIG_DIR).
func Dir() (string, error) {
	if d := os.Getenv("OUTLOOK_CONFIG_DIR"); d != "" {
		return filepath.Clean(d), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".outlook-mcp"), nil
}

// AppConfig holds client_id and client_secret from config.json.
type AppConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func LoadAppConfig(dir string) (*AppConfig, error) {
	p := filepath.Join(dir, "config.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read config.json: %w", err)
	}
	var c AppConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse config.json: %w", err)
	}
	if c.ClientID == "" {
		return nil, errors.New("config.json: missing client_id")
	}
	if c.ClientSecret == "" {
		return nil, errors.New("config.json: missing client_secret")
	}
	return &c, nil
}
