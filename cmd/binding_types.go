package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// BindingTypesCmd is the parent command for Trust Binding Type catalog management.
type BindingTypesCmd struct {
	Create btCreateCmd `cmd:"" help:"Create a binding type."`
	List   btListCmd   `cmd:"" help:"List binding types."`
	Get    btGetCmd    `cmd:"" help:"Get a binding type."`
	Delete btDeleteCmd `cmd:"" help:"Delete a binding type."`
}

// ── create ──────────────────────────────────────────────────────────────────

type btCreateCmd struct {
	Key             string `name:"key" required:"" help:"Unique key for this binding type (e.g. acme_tfc)."`
	Label           string `name:"label" required:"" help:"Human-readable label."`
	IdentityFields  string `name:"identity-fields" required:"" help:"Comma-separated list of required identity field names."`
	OutputKeySpec   string `name:"output-key-spec" help:"ID of the output key spec to issue tokens from."`
	Description     string `name:"description" help:"Optional description."`
	LogoURL         string `name:"logo-url" help:"Optional logo URL."`
	DocsURL         string `name:"docs-url" help:"Optional documentation URL."`
	Policy          string `name:"policy" help:"Optional CEL policy expression."`
}

func (c *btCreateCmd) toInput() client.TrustBindingTypeInput {
	return client.TrustBindingTypeInput{
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

func (c *btCreateCmd) Run(ctx context.Context, g *Globals) error {
	def, err := g.Client().CreateTrustBindingTypeDef(ctx, c.toInput())
	if err != nil {
		return err
	}
	fmt.Printf("%s Created binding type %s (%s)\n", output.Green.Render("✓"), def.Key, def.ID)
	return output.PrintJSON(def)
}

// ── list ────────────────────────────────────────────────────────────────────

type btListCmd struct{}

func (c *btListCmd) Run(ctx context.Context, g *Globals) error {
	defs, err := g.Client().ListTrustBindingTypeDefs(ctx)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(defs)
	}
	rows := make([][]string, len(defs))
	for i, d := range defs {
		rows[i] = []string{d.ID, d.Key, d.WorkspaceID, d.Label}
	}
	output.PrintTable([]string{"ID", "KEY", "WORKSPACE_ID", "LABEL"}, rows, false)
	return nil
}

// ── get ─────────────────────────────────────────────────────────────────────

type btGetCmd struct {
	ID string `arg:"" help:"Binding type ID."`
}

func (c *btGetCmd) Run(ctx context.Context, g *Globals) error {
	def, err := g.Client().GetTrustBindingTypeDef(ctx, c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(def)
}

// ── delete ───────────────────────────────────────────────────────────────────

type btDeleteCmd struct {
	ID string `arg:"" help:"Binding type ID."`
}

func (c *btDeleteCmd) Run(ctx context.Context, g *Globals) error {
	if err := g.Client().DeleteTrustBindingTypeDef(ctx, c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}
