package cmd

import (
	"fmt"
	"os"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/config"
)

type Globals struct {
	Token   string
	APIURL  string
	AuthURL string
	Output  string
	Debug   bool
	Version string
}

func (g *Globals) resolveToken() (string, error) {
	if g.Token != "" {
		return g.Token, nil
	}
	return config.ResolveToken(config.GlobalConfigDir())
}

// Client returns an authenticated management API client. Exits 1 if no token.
func (g *Globals) Client() *client.Client {
	token, err := g.resolveToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	return client.New(g.apiURL(), token, g.Version, g.Debug)
}

// AuthClient returns an unauthenticated client for the public auth plane.
func (g *Globals) AuthClient() *client.Client {
	return client.New(g.authURL(), "", g.Version, g.Debug)
}

// PublicClient returns an unauthenticated management-API client for the device-flow
// public endpoints (request code, poll token). It never exits on a missing token.
func (g *Globals) PublicClient() *client.Client {
	return client.New(g.apiURL(), "", g.Version, g.Debug)
}

func (g *Globals) apiURL() string {
	if g.APIURL != "" {
		return g.APIURL
	}
	return config.ResolveAPIURL()
}

func (g *Globals) authURL() string {
	if g.AuthURL != "" {
		return g.AuthURL
	}
	return config.ResolveAuthURL()
}

func (g *Globals) IsJSON() bool { return g.Output == "json" }
