package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

type LoginCmd struct {
	Key       string `name:"key" help:"Paste a management API key instead of the browser device flow (CI/manual)."`
	NoBrowser bool   `name:"no-browser" help:"Do not attempt to open a browser automatically."`
}

func (c *LoginCmd) Run(ctx context.Context, g *Globals) error {
	if key := strings.TrimSpace(c.Key); key != "" {
		return c.store(key)
	}

	pub := g.PublicClient()
	dc, err := pub.RequestDeviceCode(ctx)
	if err != nil {
		return fmt.Errorf("start device login: %w", err)
	}

	fmt.Printf("\n  To authorize this CLI, visit:\n  %s\n\n", output.Bold.Render(dc.AuthorizeURL))
	if !c.NoBrowser {
		openBrowser(dc.AuthorizeURL)
	}
	fmt.Println("  Waiting for approval...")

	interval := time.Duration(dc.Interval) * time.Second
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		if interval > 0 {
			time.Sleep(interval)
		}
		poll, err := pub.PollDeviceToken(ctx, dc.DeviceCode)
		if err != nil {
			return err
		}
		switch poll.Status {
		case "approved":
			return c.store(poll.Token)
		case "expired":
			return fmt.Errorf("device code expired; run `microwave login` again")
		}
	}
	return fmt.Errorf("authorization timed out")
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
