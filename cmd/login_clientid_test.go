package cmd

import (
	"context"
	"testing"

	"github.com/microwave-sh/microwave-go/auth"
)

func TestLoginUsesSystemCLIClientByDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	var gotClientID string
	c := &LoginCmd{}
	c.loginFn = func(ctx context.Context, cfg auth.LoginConfig) (*auth.Credentials, error) {
		gotClientID = cfg.ClientID
		return &auth.Credentials{AccessToken: "tok"}, nil
	}
	g := &Globals{AuthURL: "https://auth.test.invalid", Version: "test"}
	if err := c.Run(context.Background(), g); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotClientID != "microwave-cli" {
		t.Fatalf("ClientID = %q, want microwave-cli", gotClientID)
	}
}
