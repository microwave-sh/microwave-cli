package cmd

import (
	"context"
	"testing"

	"github.com/microwave-sh/microwave-go/auth"
)

// TestLoginUsesDeviceApprovalConfig pins that `microwave login` drives the
// device-approval flow: it passes the auth host as DeviceApprovalURL (where the
// device endpoints live), no client-id, and lets LoginAuto pick the flow from
// the server's metadata hint.
func TestLoginUsesDeviceApprovalConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	var got auth.LoginConfig
	c := &LoginCmd{}
	c.loginFn = func(ctx context.Context, cfg auth.LoginConfig) (*auth.Credentials, error) {
		got = cfg
		return &auth.Credentials{AccessToken: "tok"}, nil
	}
	g := &Globals{AuthURL: "https://auth.test.invalid", APIURL: "https://api.test.invalid", Version: "test"}
	if err := c.Run(context.Background(), g); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got.ClientID != "" {
		t.Fatalf("ClientID = %q, want empty (device-approval needs none)", got.ClientID)
	}
	if got.DeviceApprovalURL != "https://auth.test.invalid" {
		t.Fatalf("DeviceApprovalURL = %q, want the auth host", got.DeviceApprovalURL)
	}
	if got.Mode != auth.LoginAuto {
		t.Fatalf("Mode = %v, want LoginAuto (server-advertised device-approval)", got.Mode)
	}
}
