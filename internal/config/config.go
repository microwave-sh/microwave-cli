package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// GlobalConfig represents ~/.config/microwave/config.toml
type GlobalConfig struct {
	Auth    AuthConfig `toml:"auth"`
	APIURL  string     `toml:"api_url,omitempty"`
	AuthURL string     `toml:"auth_url,omitempty"`
}

// AuthConfig holds the stored management API key.
type AuthConfig struct {
	Token string `toml:"token"`
}

// GlobalConfigDir returns the XDG-compliant config directory for microwave.
func GlobalConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "microwave")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "microwave")
}

// ResolveToken returns the auth token using priority:
// MICROWAVE_TOKEN env → MICROWAVE_API_KEY env (legacy) → config file → error.
func ResolveToken(globalDir string) (string, error) {
	if key := os.Getenv("MICROWAVE_TOKEN"); key != "" {
		return key, nil
	}
	if key := os.Getenv("MICROWAVE_API_KEY"); key != "" {
		return key, nil
	}

	path := filepath.Join(globalDir, "config.toml")
	data, err := os.ReadFile(path)
	if err == nil {
		var cfg GlobalConfig
		if err := toml.Unmarshal(data, &cfg); err == nil && cfg.Auth.Token != "" {
			return cfg.Auth.Token, nil
		}
	}

	return "", fmt.Errorf("not logged in. Run `microwave login` to authenticate")
}

// WriteGlobalAuth saves an auth token to the global config file,
// preserving any existing api_url / auth_url values.
func WriteGlobalAuth(token string) error {
	return WriteGlobalAuthTo(GlobalConfigDir(), token)
}

// WriteGlobalAuthTo is the testable variant of WriteGlobalAuth that writes
// to an explicit directory rather than GlobalConfigDir().
func WriteGlobalAuthTo(dir, token string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(dir, "config.toml")

	// Preserve existing api_url / auth_url when rewriting.
	var cfg GlobalConfig
	if data, err := os.ReadFile(path); err == nil {
		toml.Unmarshal(data, &cfg) //nolint:errcheck
	}

	cfg.Auth.Token = token

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(cfg)
}

// ResolveAPIURL returns the management API base URL using priority:
// MICROWAVE_API_URL env → config api_url → default.
func ResolveAPIURL() string {
	if v := os.Getenv("MICROWAVE_API_URL"); v != "" {
		return v
	}
	path := filepath.Join(GlobalConfigDir(), "config.toml")
	if data, err := os.ReadFile(path); err == nil {
		var cfg GlobalConfig
		if toml.Unmarshal(data, &cfg) == nil && cfg.APIURL != "" {
			return cfg.APIURL
		}
	}
	return "https://api.microwave.sh"
}

// ResolveAuthURL returns the auth-plane base URL using priority:
// MICROWAVE_AUTH_URL env → config auth_url → default.
func ResolveAuthURL() string {
	if v := os.Getenv("MICROWAVE_AUTH_URL"); v != "" {
		return v
	}
	path := filepath.Join(GlobalConfigDir(), "config.toml")
	if data, err := os.ReadFile(path); err == nil {
		var cfg GlobalConfig
		if toml.Unmarshal(data, &cfg) == nil && cfg.AuthURL != "" {
			return cfg.AuthURL
		}
	}
	return "https://auth.microwave.sh"
}

// ClearAuth removes the global config file (ignore not-found errors).
// Used by the logout command.
func ClearAuth() error {
	path := filepath.Join(GlobalConfigDir(), "config.toml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove config: %w", err)
	}
	return nil
}
