// Copyright 2026 Mataki Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultAPIURL      = "https://api.microwave.dev"
	defaultConsoleURL  = "https://app.microwave.sh"
	defaultAPIVersion  = "2026-05-25"
	projectConfigFile  = "microwave.yaml"
	defaultHTTPTimeout = 30 * time.Second
)

type GlobalConfig struct {
	Auth      GlobalAuth
	Workspace GlobalWorkspace
	APIURL    string
}

type GlobalAuth struct {
	APIKey string
}

type GlobalWorkspace struct {
	Active string
}

func GlobalConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "microwave")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "microwave")
}

func globalConfigPath() string {
	return filepath.Join(GlobalConfigDir(), "config.toml")
}

func LoadGlobalConfig() GlobalConfig {
	data, err := os.ReadFile(globalConfigPath())
	if err != nil {
		return GlobalConfig{}
	}

	var cfg GlobalConfig
	section := ""
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := stripInlineComment(strings.TrimSpace(scanner.Text()))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]"))
			continue
		}
		key, value, ok := splitConfigKV(line)
		if !ok {
			continue
		}
		switch section {
		case "":
			if key == "api_url" {
				cfg.APIURL = value
			}
		case "auth":
			if key == "api_key" || key == "token" {
				cfg.Auth.APIKey = value
			}
		case "workspace":
			if key == "active" {
				cfg.Workspace.Active = value
			}
		}
	}
	return cfg
}

func SaveGlobalConfig(cfg GlobalConfig) error {
	if err := os.MkdirAll(GlobalConfigDir(), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	var b strings.Builder
	if cfg.APIURL != "" {
		fmt.Fprintf(&b, "api_url = %q\n\n", cfg.APIURL)
	}
	if cfg.Auth.APIKey != "" {
		fmt.Fprintf(&b, "[auth]\napi_key = %q\n\n", cfg.Auth.APIKey)
	}
	if cfg.Workspace.Active != "" {
		fmt.Fprintf(&b, "[workspace]\nactive = %q\n", cfg.Workspace.Active)
	}

	path := globalConfigPath()
	if err := os.WriteFile(path, []byte(b.String()), 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func clearGlobalAuth() error {
	cfg := LoadGlobalConfig()
	cfg.Auth.APIKey = ""
	return SaveGlobalConfig(cfg)
}

type ProjectConfig struct {
	Path        string
	Version     string
	Workspace   string
	SpecPath    string
	SDKTargets  map[string]SDKTarget
	Docs        DocsProjectConfig
	RawContents string
}

type SDKTarget struct {
	Repo   string
	Module string
}

type DocsProjectConfig struct {
	Domain    string
	GuidesDir string
}

func projectConfigPath(dir string) string {
	return filepath.Join(dir, projectConfigFile)
}

func LoadProjectConfig(dir string) (ProjectConfig, error) {
	path := projectConfigPath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("no %s found. Run `microwave config init` first", projectConfigFile)
	}
	cfg := parseProjectConfig(string(data))
	cfg.Path = path
	cfg.RawContents = string(data)
	if cfg.SpecPath == "" {
		return ProjectConfig{}, fmt.Errorf("%s is missing spec.path", projectConfigFile)
	}
	return cfg, nil
}

func parseProjectConfig(contents string) ProjectConfig {
	cfg := ProjectConfig{SDKTargets: map[string]SDKTarget{}}
	section := ""
	sdkLang := ""

	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		raw := scanner.Text()
		trimmed := stripInlineComment(strings.TrimSpace(raw))
		if trimmed == "" {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if indent == 0 && strings.HasSuffix(trimmed, ":") {
			section = strings.TrimSuffix(trimmed, ":")
			sdkLang = ""
			continue
		}
		if section == "sdks" && indent == 2 && strings.HasSuffix(trimmed, ":") {
			sdkLang = strings.TrimSuffix(trimmed, ":")
			if _, ok := cfg.SDKTargets[sdkLang]; !ok {
				cfg.SDKTargets[sdkLang] = SDKTarget{}
			}
			continue
		}

		key, value, ok := splitConfigKV(trimmed)
		if !ok {
			continue
		}
		switch {
		case indent == 0 && key == "version":
			cfg.Version = value
		case indent == 0 && key == "workspace":
			cfg.Workspace = value
		case section == "spec" && key == "path":
			cfg.SpecPath = value
		case section == "sdks" && sdkLang != "":
			target := cfg.SDKTargets[sdkLang]
			if key == "repo" {
				target.Repo = value
			}
			if key == "module" {
				target.Module = value
			}
			cfg.SDKTargets[sdkLang] = target
		case section == "docs" && key == "domain":
			cfg.Docs.Domain = value
		case section == "docs" && key == "guides_dir":
			cfg.Docs.GuidesDir = value
		}
	}
	return cfg
}

func (c ProjectConfig) SDKLanguages() []string {
	languages := make([]string, 0, len(c.SDKTargets))
	for language := range c.SDKTargets {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	return languages
}

func (c ProjectConfig) ResolveSpecPath() string {
	if filepath.IsAbs(c.SpecPath) {
		return c.SpecPath
	}
	return filepath.Join(filepath.Dir(c.Path), c.SpecPath)
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func splitConfigKV(line string) (string, string, bool) {
	eq := strings.Index(line, "=")
	colon := strings.Index(line, ":")
	if eq == -1 && colon == -1 {
		return "", "", false
	}
	cut := eq
	if cut == -1 || (colon != -1 && colon < eq) {
		cut = colon
	}
	key, value := line[:cut], line[cut+1:]
	return strings.TrimSpace(key), unquote(strings.TrimSpace(value)), true
}

func stripInlineComment(line string) string {
	inQuote := false
	for i, r := range line {
		switch r {
		case '"':
			inQuote = !inQuote
		case '#':
			if !inQuote {
				return strings.TrimSpace(line[:i])
			}
		}
	}
	return line
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"'")
	return value
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}
