package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// TrustExchangesCmd is the parent command for trust exchange management.
type TrustExchangesCmd struct {
	List   trustExchangesListCmd   `cmd:"" help:"Search trust exchanges."`
	Get    trustExchangesGetCmd    `cmd:"" help:"Get a trust exchange."`
	Create trustExchangesCreateCmd `cmd:"" help:"Create a trust exchange."`
	Update trustExchangesUpdateCmd `cmd:"" help:"Update a trust exchange."`
	Delete trustExchangesDeleteCmd `cmd:"" help:"Delete a trust exchange."`
}

// ── list ─────────────────────────────────────────────────────────────────

type trustExchangesListCmd struct {
	listFlags
	Provider   string `help:"Filter by provider (github, google, auth0, custom_oidc)."`
	OutputMode string `name:"output-mode" help:"Filter by output mode (claims, jwt)."`
	Active     *bool  `help:"Filter by active status."`
}

func (c *trustExchangesListCmd) Run(ctx context.Context, g *Globals) error {
	filter := map[string]map[string]any{}
	if c.Provider != "" {
		filter["provider"] = map[string]any{"eq": c.Provider}
	}
	if c.OutputMode != "" {
		filter["output_mode"] = map[string]any{"eq": c.OutputMode}
	}
	if c.Active != nil {
		filter["active"] = map[string]any{"eq": *c.Active}
	}
	var filterArg map[string]map[string]any
	if len(filter) > 0 {
		filterArg = filter
	}
	page, err := g.Client().SearchTrustExchanges(ctx, c.searchRequest(filterArg))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, te := range page.Data {
		rows[i] = []string{te.ID, te.Name, te.Provider, te.OutputMode, fmt.Sprintf("%v", te.Active)}
	}
	output.PrintTable([]string{"ID", "Name", "Provider", "Output", "Active"}, rows, false)
	return nil
}

// ── get ──────────────────────────────────────────────────────────────────

type trustExchangesGetCmd struct {
	ID string `arg:"" help:"Trust exchange ID."`
}

func (c *trustExchangesGetCmd) Run(ctx context.Context, g *Globals) error {
	te, err := g.Client().GetTrustExchange(ctx, c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(te)
}

// ── create / update shared fields ────────────────────────────────────────

type trustExchangesCreateCmd struct {
	Name             string `help:"Name." required:""`
	Description      string `help:"Description."`
	Provider         string `help:"Provider (github, google, auth0, custom_oidc)." required:"" enum:"github,google,auth0,custom_oidc"`
	Issuer           string `help:"OIDC issuer URL." required:""`
	DiscoveryURL     string `name:"discovery-url" help:"Discovery document URL."`
	JWKSURL          string `name:"jwks-url" help:"JWKS endpoint URL."`
	AllowedAudiences string `name:"allowed-audiences" help:"Comma-separated allowed audiences." required:""`
	SubjectExact     string `name:"subject-exact" help:"Exact subject value rule."`
	SubjectPrefix    string `name:"subject-prefix" help:"Subject prefix rule."`
	OutputMode       string `name:"output-mode" help:"Output mode (claims, jwt)." required:"" enum:"claims,jwt"`
	OutputKeySpecID  string `name:"output-key-spec-id" help:"Output key spec ID (required when output-mode=jwt)."`
	VerificationURI  string `name:"verification-uri" help:"Verification URI surfaced to the caller (e.g. device-flow verification page)."`
	ClaimRules       string `name:"claim-rules" help:"Claim rules as JSON object (e.g. {\"repo\":{\"equals\":\"x\",\"required\":true}})."`
	ClaimMapping     string `name:"claim-mapping" help:"Claim mapping as JSON object."`
	Active           bool   `name:"active" default:"true" help:"Whether the trust exchange is active."`
}

// toInput builds a TrustExchangeInput from the command flags.
func (c *trustExchangesCreateCmd) toInput() (client.TrustExchangeInput, error) {
	in := client.TrustExchangeInput{
		Name:             c.Name,
		Description:      c.Description,
		Type:             "oidc",
		Provider:         c.Provider,
		Issuer:           c.Issuer,
		DiscoveryURL:     c.DiscoveryURL,
		JWKSURL:          c.JWKSURL,
		AllowedAudiences: parseCSV(c.AllowedAudiences),
		SubjectRules: client.TrustExchangeSubjectRules{
			Exact:  c.SubjectExact,
			Prefix: c.SubjectPrefix,
		},
		OutputMode:      c.OutputMode,
		OutputKeySpecID: c.OutputKeySpecID,
		VerificationURI: c.VerificationURI,
		Active:          c.Active,
	}

	if strings.TrimSpace(c.ClaimRules) != "" {
		var rules map[string]client.TrustExchangeClaimRule
		if err := json.Unmarshal([]byte(c.ClaimRules), &rules); err != nil {
			return client.TrustExchangeInput{}, fmt.Errorf("invalid --claim-rules JSON: %w", err)
		}
		in.ClaimRules = rules
	}

	if strings.TrimSpace(c.ClaimMapping) != "" {
		var mapping client.TrustExchangeClaimMapping
		if err := json.Unmarshal([]byte(c.ClaimMapping), &mapping); err != nil {
			return client.TrustExchangeInput{}, fmt.Errorf("invalid --claim-mapping JSON: %w", err)
		}
		in.ClaimMapping = mapping
	}

	return in, nil
}

func (c *trustExchangesCreateCmd) Run(ctx context.Context, g *Globals) error {
	in, err := c.toInput()
	if err != nil {
		return err
	}
	te, err := g.Client().CreateTrustExchange(ctx, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(te)
}

// ── update ───────────────────────────────────────────────────────────────

type trustExchangesUpdateCmd struct {
	ID string `arg:"" help:"Trust exchange ID."`
	trustExchangesCreateCmd
}

func (c *trustExchangesUpdateCmd) Run(ctx context.Context, g *Globals) error {
	in, err := c.toInput()
	if err != nil {
		return err
	}
	te, err := g.Client().UpdateTrustExchange(ctx, c.ID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(te)
}

// ── delete ───────────────────────────────────────────────────────────────

type trustExchangesDeleteCmd struct {
	ID string `arg:"" help:"Trust exchange ID."`
}

func (c *trustExchangesDeleteCmd) Run(ctx context.Context, g *Globals) error {
	if err := g.Client().DeleteTrustExchange(ctx, c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}
