package config_test

import (
	"path/filepath"
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/config"
)

func TestResolveToken_EnvOverridesConfig(t *testing.T) {
	t.Setenv("MICROWAVE_TOKEN", "env-token")
	got, err := config.ResolveToken(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "env-token" {
		t.Fatalf("ResolveToken = %q, want env-token", got)
	}
}

func TestResolveToken_LegacyEnvFallback(t *testing.T) {
	t.Setenv("MICROWAVE_API_KEY", "legacy")
	got, err := config.ResolveToken(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "legacy" {
		t.Fatalf("ResolveToken = %q, want legacy", got)
	}
}

func TestWriteAndResolveToken_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := config.WriteGlobalAuthTo(dir, "stored-token"); err != nil {
		t.Fatal(err)
	}
	got, err := config.ResolveToken(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "stored-token" {
		t.Fatalf("ResolveToken = %q, want stored-token", got)
	}
}

func TestResolveToken_NotLoggedIn(t *testing.T) {
	if _, err := config.ResolveToken(filepath.Join(t.TempDir(), "empty")); err == nil {
		t.Fatal("expected error when no token present")
	}
}
