package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/output"
	"github.com/microwave-sh/microwave-go/auth"
)

type LoginCmd struct {
	Key       string `name:"key" help:"Paste a management API key instead of the browser login (CI/manual)."`
	NoBrowser bool   `name:"no-browser" help:"Do not open a browser; print the approval URL instead (headless / SSH)."`
	Device    bool   `name:"device" help:"Do not open a browser; print the approval URL instead (alias of --no-browser)."`

	loginFn func(context.Context, auth.LoginConfig) (*auth.Credentials, error) // test seam; nil → auth.Login
}

// Run authenticates via the shared SDK login core. It discovers the auth server
// from the auth-plane metadata, which advertises cli_login_flow=device_approval,
// so auth.Login drives the management device-approval flow: it surfaces a console
// approval URL where the operator approves with their session (carrying their
// per-operator permissions), then polls for the minted, least-privilege token.
// No client-id is involved. A pasted --key skips the interactive flow entirely.
func (c *LoginCmd) Run(ctx context.Context, g *Globals) error {
	if key := strings.TrimSpace(c.Key); key != "" {
		return c.store(key)
	}

	cfg := auth.LoginConfig{
		MetadataURL:       strings.TrimRight(g.authURL(), "/") + "/.well-known/oauth-authorization-server",
		DeviceApprovalURL: strings.TrimRight(g.apiURL(), "/"),
		Mode:              auth.LoginAuto, // auth plane advertises cli_login_flow=device_approval
		Output:            os.Stderr,
	}
	if c.Device || c.NoBrowser {
		cfg.OpenBrowser = func(string) error { return nil } // print the URL, don't open it
	}

	login := c.loginFn
	if login == nil {
		login = auth.Login
	}
	creds, err := login(ctx, cfg)
	if err != nil {
		return err
	}
	return c.store(creds.AccessToken)
}

func (c *LoginCmd) store(key string) error {
	if err := config.WriteGlobalAuth(key); err != nil {
		return err
	}
	fmt.Printf("%s Saved to %s\n",
		output.Green.Render("✓"),
		output.Dim.Render(filepath.Join(config.GlobalConfigDir(), "config.toml")))
	return nil
}
