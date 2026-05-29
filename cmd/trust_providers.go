package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// TrustProvidersCmd is the parent command for trust provider management.
type TrustProvidersCmd struct {
	List      tpListCmd      `cmd:"" help:"Search trust providers."`
	Get       tpGetCmd       `cmd:"" help:"Get a trust provider."`
	Create    tpCreateCmd    `cmd:"" help:"Create a trust provider."`
	Update    tpUpdateCmd    `cmd:"" help:"Update a trust provider."`
	Delete    tpDeleteCmd    `cmd:"" help:"Delete a trust provider."`
	Mint      tpMintCmd      `cmd:"" help:"Mint a token from a trust provider."`
	Discovery tpDiscoveryCmd `cmd:"" help:"Print federation (discovery/JWKS/token) URLs."`
}

// ── list ─────────────────────────────────────────────────────────────────

type tpListCmd struct{ listFlags }

func (c *tpListCmd) Run(g *Globals) error {
	page, err := g.Client().SearchTrustProviders(context.Background(), c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, p := range page.Data {
		rows[i] = []string{p.ID, p.Name, p.SigningKeySetID, fmt.Sprintf("%v", p.Active)}
	}
	output.PrintTable([]string{"ID", "Name", "Signing Key Set", "Active"}, rows, false)
	return nil
}

// ── get ──────────────────────────────────────────────────────────────────

type tpGetCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *tpGetCmd) Run(g *Globals) error {
	p, err := g.Client().GetTrustProvider(context.Background(), c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(p)
}

// ── create ───────────────────────────────────────────────────────────────

type tpCreateCmd struct {
	Name             string `help:"Name." required:""`
	Description      string `help:"Description."`
	SigningKeySetID  string `name:"signing-key-set-id" help:"Asymmetric signing key set ID." required:""`
	IssuerHost       string `name:"issuer-host" help:"Custom issuer host."`
	AllowedAudiences string `name:"allowed-audiences" help:"Comma-separated allowed audiences." required:""`
	DefaultAudience  string `name:"default-audience" help:"Default audience."`
	AllowedClaims    string `name:"allowed-claims" help:"Comma-separated allowed claim keys."`
	RequiredClaims   string `name:"required-claims" help:"Comma-separated required claim keys."`
	ConstantClaims   string `name:"constant-claims" help:"Constant claims as JSON object."`
	SubjectRequired  bool   `name:"subject-required" help:"Require a subject." default:"true"`
	TTLDefault       int64  `name:"ttl-default-seconds" help:"Default token TTL." default:"3600"`
	TTLMax           int64  `name:"ttl-max-seconds" help:"Max token TTL." default:"3600"`
}

func (c *tpCreateCmd) Run(g *Globals) error {
	constant, err := parseJSONMap(c.ConstantClaims)
	if err != nil {
		return err
	}
	in := client.TrustProviderInput{
		Name:             c.Name,
		Description:      c.Description,
		Type:             "oidc",
		SigningKeySetID:  c.SigningKeySetID,
		IssuerHost:       c.IssuerHost,
		AllowedAudiences: parseCSV(c.AllowedAudiences),
		DefaultAudience:  c.DefaultAudience,
		ClaimPolicy: client.TrustProviderClaimPolicy{
			Allowed:  parseCSV(c.AllowedClaims),
			Required: parseCSV(c.RequiredClaims),
			Constant: constant,
		},
		SubjectRequired:   c.SubjectRequired,
		TTLDefaultSeconds: c.TTLDefault,
		TTLMaxSeconds:     c.TTLMax,
		Active:            true,
	}
	p, err := g.Client().CreateTrustProvider(context.Background(), in)
	if err != nil {
		return err
	}
	return output.PrintJSON(p)
}

// ── update ───────────────────────────────────────────────────────────────

type tpUpdateCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
	tpCreateCmd
}

func (c *tpUpdateCmd) Run(g *Globals) error {
	constant, err := parseJSONMap(c.ConstantClaims)
	if err != nil {
		return err
	}
	in := client.TrustProviderInput{
		Name:             c.Name,
		Description:      c.Description,
		Type:             "oidc",
		SigningKeySetID:  c.SigningKeySetID,
		IssuerHost:       c.IssuerHost,
		AllowedAudiences: parseCSV(c.AllowedAudiences),
		DefaultAudience:  c.DefaultAudience,
		ClaimPolicy: client.TrustProviderClaimPolicy{
			Allowed:  parseCSV(c.AllowedClaims),
			Required: parseCSV(c.RequiredClaims),
			Constant: constant,
		},
		SubjectRequired:   c.SubjectRequired,
		TTLDefaultSeconds: c.TTLDefault,
		TTLMaxSeconds:     c.TTLMax,
		Active:            true,
	}
	p, err := g.Client().UpdateTrustProvider(context.Background(), c.ID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(p)
}

// ── delete ───────────────────────────────────────────────────────────────

type tpDeleteCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *tpDeleteCmd) Run(g *Globals) error {
	if err := g.Client().DeleteTrustProvider(context.Background(), c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}

// ── mint ─────────────────────────────────────────────────────────────────

type tpMintCmd struct {
	ID       string `arg:"" help:"Trust provider ID."`
	APIKey   string `name:"api-key" help:"Management API key to authenticate the mint (defaults to stored token)."`
	Subject  string `help:"Token subject."`
	Audience string `help:"Token audience."`
	Claims   string `help:"Claims as JSON object."`
	TTL      int64  `name:"ttl" help:"Token TTL seconds."`
}

func (c *tpMintCmd) Run(g *Globals) error {
	claims, err := parseJSONMap(c.Claims)
	if err != nil {
		return err
	}
	key := c.APIKey
	if key == "" {
		key, err = g.resolveToken()
		if err != nil {
			return err
		}
	}
	res, err := g.AuthClient().MintTrustProviderToken(context.Background(), c.ID, key, client.MintTokenInput{
		Subject:    c.Subject,
		Audience:   c.Audience,
		Claims:     claims,
		TTLSeconds: c.TTL,
	})
	if err != nil {
		return err
	}
	return output.PrintJSON(res)
}

// ── discovery ────────────────────────────────────────────────────────────

type tpDiscoveryCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *tpDiscoveryCmd) Run(g *Globals) error {
	base := g.authURL() + "/trust-providers/" + c.ID
	rows := [][]string{
		{"Issuer", base},
		{"Discovery", base + "/.well-known/openid-configuration"},
		{"JWKS", base + "/.well-known/jwks.json"},
		{"Token", base + "/token"},
	}
	output.PrintTable([]string{"Endpoint", "URL"}, rows, g.IsJSON())
	return nil
}
