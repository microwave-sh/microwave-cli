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
	NoBrowser bool   `name:"no-browser" help:"Do not open a browser; use the device-code flow instead."`
	Device    bool   `name:"device" help:"Force the device-code flow (headless / SSH)."`
	ClientID  string `name:"client-id" help:"OAuth client id (the CLI session key spec); overrides config/env."`
}

// Run authenticates via the shared SDK login core: it discovers the
// authorization server from the auth-plane metadata document, runs the loopback
// authorization-code + PKCE flow (falling back to the device grant), and stores
// the minted session. A pasted --key skips the interactive flow entirely.
func (c *LoginCmd) Run(ctx context.Context, g *Globals) error {
	if key := strings.TrimSpace(c.Key); key != "" {
		return c.store(key)
	}

	clientID := c.ClientID
	if clientID == "" {
		clientID = config.ResolveAuthClientID()
	}
	if clientID == "" {
		return fmt.Errorf("OAuth client id is required: pass --client-id, set MICROWAVE_AUTH_CLIENT_ID, or add auth_client_id to config")
	}

	mode := auth.LoginAuto
	if c.Device || c.NoBrowser {
		mode = auth.LoginDevice
	}

	cfg := auth.LoginConfig{
		MetadataURL: strings.TrimRight(g.authURL(), "/") + "/.well-known/oauth-authorization-server",
		ClientID:    clientID,
		Mode:        mode,
		Output:      os.Stderr,
	}
	if c.NoBrowser {
		cfg.OpenBrowser = func(string) error { return nil } // print the URL, don't open it
	}

	creds, err := auth.Login(ctx, cfg)
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
