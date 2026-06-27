package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

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

	// Bridge Ctrl+C to the login: while a spinner runs, Bubbletea holds the
	// terminal in raw mode and swallows the interrupt (it never reaches the
	// signal.NotifyContext in main), so cancel this child context off the
	// spinner instead.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	pr := &loginProgress{cancel: cancel}

	cfg := auth.LoginConfig{
		MetadataURL:       strings.TrimRight(g.authURL(), "/") + "/.well-known/oauth-authorization-server",
		DeviceApprovalURL: strings.TrimRight(g.apiURL(), "/"),
		Mode:              auth.LoginAuto, // auth plane advertises cli_login_flow=device_approval
		Output:            os.Stderr,
		Progress:          pr,
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
		if pr.Cancelled() {
			return fmt.Errorf("login cancelled")
		}
		return err
	}
	return c.store(creds.AccessToken)
}

// loginProgress renders auth.ProgressReporter phase events with the CLI's
// spinner so `microwave login` matches the rest of the CLI's status output
// (and the sibling sandbar CLI): an animated step that resolves to a green ✓
// or red ✗. The SDK pairs each Begin with exactly one Succeed or Fail, so at
// most one spinner is live at a time.
type loginProgress struct {
	cancel    context.CancelFunc
	sp        *output.Spinner
	watchDone chan struct{}
	cancelled atomic.Bool
}

// Begin starts a spinner for the phase and watches it for a Ctrl+C, which
// cancels the login. The watcher exits when the phase ends (watchDone) so it
// never outlives its spinner.
func (p *loginProgress) Begin(message string) {
	p.sp = output.NewSpinner(message)
	p.watchDone = make(chan struct{})
	go func(sp *output.Spinner, done chan struct{}) {
		select {
		case <-sp.CancelledC:
			p.cancelled.Store(true)
			if p.cancel != nil {
				p.cancel()
			}
		case <-done:
		}
	}(p.sp, p.watchDone)
}

// Succeed resolves the active phase with a green ✓.
func (p *loginProgress) Succeed(message string) { p.finish(false, message) }

// Fail resolves the active phase with a red ✗.
func (p *loginProgress) Fail(message string) { p.finish(true, message) }

func (p *loginProgress) finish(failed bool, message string) {
	if p.sp == nil {
		return
	}
	sp := p.sp
	p.sp = nil
	close(p.watchDone)
	// On Ctrl+C the spinner already rendered its own "Cancelled" line and the
	// Bubbletea program exited; don't send to a finished program.
	if p.cancelled.Load() {
		return
	}
	if failed {
		sp.Fail(message)
	} else {
		sp.Stop(message)
	}
}

// Cancelled reports whether the operator interrupted the login with Ctrl+C.
func (p *loginProgress) Cancelled() bool { return p.cancelled.Load() }

func (c *LoginCmd) store(key string) error {
	if err := config.WriteGlobalAuth(key); err != nil {
		return err
	}
	fmt.Printf("%s Saved to %s\n",
		output.Green.Render("✓"),
		output.Dim.Render(filepath.Join(config.GlobalConfigDir(), "config.toml")))
	return nil
}
