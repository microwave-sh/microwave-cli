package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// FederationsCmd is the parent command for Trust Federation catalog management.
type FederationsCmd struct {
	Create fedCreateCmd `cmd:"" help:"Create a trust federation."`
	List   fedListCmd   `cmd:"" help:"List trust federations."`
	Get    fedGetCmd    `cmd:"" help:"Get a trust federation."`
	Delete fedDeleteCmd `cmd:"" help:"Delete a trust federation."`
}

// ── create ──────────────────────────────────────────────────────────────────

type fedCreateCmd struct {
	Key            string `name:"key" required:"" help:"Unique key for this federation (e.g. acme_tfc)."`
	Label          string `name:"label" required:"" help:"Human-readable label."`
	IdentityFields string `name:"identity-fields" required:"" help:"Comma-separated list of required identity field names."`
	OutputKeySpec  string `name:"output-key-spec" help:"ID of the output key spec to issue tokens from."`
	Description    string `name:"description" help:"Optional description."`
	LogoURL        string `name:"logo-url" help:"Optional logo URL."`
	DocsURL        string `name:"docs-url" help:"Optional documentation URL."`
	Policy         string `name:"policy" help:"Optional CEL policy expression."`
}

func (c *fedCreateCmd) toInput() client.TrustFederationInput {
	return client.TrustFederationInput{
		Key:             c.Key,
		Label:           c.Label,
		IdentityFields:  parseCSV(c.IdentityFields),
		OutputKeySpecID: c.OutputKeySpec,
		Description:     c.Description,
		LogoURL:         c.LogoURL,
		DocsURL:         c.DocsURL,
		Policy:          c.Policy,
	}
}

func (c *fedCreateCmd) Run(ctx context.Context, g *Globals) error {
	def, err := g.Client().CreateTrustFederation(ctx, c.toInput())
	if err != nil {
		return err
	}
	fmt.Printf("%s Created federation %s (%s)\n", output.Green.Render("✓"), def.Key, def.ID)
	return output.PrintJSON(def)
}

// ── list ────────────────────────────────────────────────────────────────────

type fedListCmd struct {
	listFlags
}

func (c *fedListCmd) Run(ctx context.Context, g *Globals) error {
	page, err := g.Client().SearchTrustFederations(ctx, c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, d := range page.Data {
		rows[i] = []string{d.ID, d.Key, d.WorkspaceID, d.Label}
	}
	output.PrintTable([]string{"ID", "KEY", "WORKSPACE_ID", "LABEL"}, rows, false)
	return nil
}

// ── get ─────────────────────────────────────────────────────────────────────

type fedGetCmd struct {
	ID string `arg:"" help:"Federation ID."`
}

func (c *fedGetCmd) Run(ctx context.Context, g *Globals) error {
	def, err := g.Client().GetTrustFederation(ctx, c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(def)
}

// ── delete ───────────────────────────────────────────────────────────────────

type fedDeleteCmd struct {
	ID string `arg:"" help:"Federation ID."`
}

func (c *fedDeleteCmd) Run(ctx context.Context, g *Globals) error {
	if err := g.Client().DeleteTrustFederation(ctx, c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}
