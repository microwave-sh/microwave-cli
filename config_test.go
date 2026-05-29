package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGlobalConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := GlobalConfig{
		APIURL:    "https://api.test",
		Auth:      GlobalAuth{APIKey: "mw_test_123456"},
		Workspace: GlobalWorkspace{Active: "acme"},
	}
	if err := SaveGlobalConfig(cfg); err != nil {
		t.Fatalf("SaveGlobalConfig() error = %v", err)
	}

	got := LoadGlobalConfig()
	if got.APIURL != cfg.APIURL {
		t.Fatalf("APIURL = %q, want %q", got.APIURL, cfg.APIURL)
	}
	if got.Auth.APIKey != cfg.Auth.APIKey {
		t.Fatalf("APIKey = %q, want %q", got.Auth.APIKey, cfg.Auth.APIKey)
	}
	if got.Workspace.Active != cfg.Workspace.Active {
		t.Fatalf("Active = %q, want %q", got.Workspace.Active, cfg.Workspace.Active)
	}

	info, err := os.Stat(globalConfigPath())
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0600 {
		t.Fatalf("config mode = %v, want 0600", gotMode)
	}
}

func TestParseProjectConfig(t *testing.T) {
	cfg := parseProjectConfig(`version: "2026-05-25"
workspace: "acme"

spec:
  path: openapi.yaml

sdks:
  go:
    repo: acme/acme-go
    module: github.com/acme/acme-go
  python:
    repo: acme/acme-python
  node:
    repo: acme/acme-node

docs:
  domain: docs.acme.com
  guides_dir: ./guides
`)

	if cfg.Version != defaultAPIVersion {
		t.Fatalf("Version = %q, want %q", cfg.Version, defaultAPIVersion)
	}
	if cfg.Workspace != "acme" {
		t.Fatalf("Workspace = %q, want acme", cfg.Workspace)
	}
	if cfg.SpecPath != "openapi.yaml" {
		t.Fatalf("SpecPath = %q, want openapi.yaml", cfg.SpecPath)
	}
	if cfg.SDKTargets["go"].Module != "github.com/acme/acme-go" {
		t.Fatalf("Go module = %q", cfg.SDKTargets["go"].Module)
	}
	if got, want := cfg.SDKLanguages(), []string{"go", "node", "python"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SDKLanguages() = %#v, want %#v", got, want)
	}
	if cfg.Docs.Domain != "docs.acme.com" {
		t.Fatalf("Docs.Domain = %q, want docs.acme.com", cfg.Docs.Domain)
	}
}

func TestLoadProjectConfigRequiresSpecPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, projectConfigFile), []byte("version: \"2026-05-25\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProjectConfig(dir); err == nil {
		t.Fatal("LoadProjectConfig() error = nil, want missing spec.path error")
	}
}
