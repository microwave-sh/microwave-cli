package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// TrustBindingsCmd is the parent command for Trust Binding management.
type TrustBindingsCmd struct {
	List   tbListCmd   `cmd:"" help:"List trust bindings."`
	Get    tbGetCmd    `cmd:"" help:"Get a trust binding."`
	Create tbCreateCmd `cmd:"" help:"Create a trust binding."`
	Delete tbDeleteCmd `cmd:"" help:"Delete a trust binding."`
	Types  tbTypesCmd  `cmd:"" help:"List available trust binding types."`
	Enable tbEnableCmd `cmd:"" help:"Enable a well-known trust binding type."`
}

// ── list ────────────────────────────────────────────────────────────────

type tbListCmd struct {
	listFlags
}

func (c *tbListCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()
	page, err := api.SearchTrustBindings(ctx, c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, b := range page.Data {
		rows[i] = []string{b.ID, b.BindingType, output.FormatTimeAgo(b.CreatedAt)}
	}
	output.PrintTable([]string{"ID", "Type", "Created"}, rows, false)
	return nil
}

// ── get ─────────────────────────────────────────────────────────────────

type tbGetCmd struct {
	ID string `arg:"" help:"Trust binding ID."`
}

func (c *tbGetCmd) Run(ctx context.Context, g *Globals) error {
	binding, err := g.Client().GetTrustBinding(ctx, c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}

// ── create ──────────────────────────────────────────────────────────────

type tbCreateCmd struct {
	BindingType  string `arg:"" help:"Trust binding type key."`
	Identity     string `name:"identity" required:"" help:"Identity claims as a JSON object."`
	OutputClaims string `name:"output-claims" help:"Output claims as a JSON object."`
}

func (c *tbCreateCmd) toInput() (client.TrustBindingInput, error) {
	identity, err := parseJSONMap(c.Identity)
	if err != nil {
		return client.TrustBindingInput{}, err
	}
	outputClaims, err := parseJSONMap(c.OutputClaims)
	if err != nil {
		return client.TrustBindingInput{}, err
	}
	return client.TrustBindingInput{
		BindingType:  c.BindingType,
		Identity:     identity,
		OutputClaims: outputClaims,
	}, nil
}

func (c *tbCreateCmd) Run(ctx context.Context, g *Globals) error {
	in, err := c.toInput()
	if err != nil {
		return err
	}
	binding, err := g.Client().CreateTrustBinding(ctx, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}

// ── delete ──────────────────────────────────────────────────────────────

type tbDeleteCmd struct {
	ID string `arg:"" help:"Trust binding ID."`
}

func (c *tbDeleteCmd) Run(ctx context.Context, g *Globals) error {
	if err := g.Client().DeleteTrustBinding(ctx, c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}

// ── types ───────────────────────────────────────────────────────────────

type tbTypesCmd struct{ listFlags }

func (c *tbTypesCmd) Run(ctx context.Context, g *Globals) error {
	page, err := g.Client().SearchTrustBindingTypes(ctx, c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, t := range page.Data {
		rows[i] = []string{t.Key, t.DisplayName, t.DocsURL}
	}
	output.PrintTable([]string{"Key", "Name", "Docs"}, rows, false)
	return nil
}

// ── enable ──────────────────────────────────────────────────────────────
//
// enable <catalog-key> --identity k=v,k=v [--output-claims k=v,k=v]
//
// 1. Resolves the binding type catalog row via GET /api/trust-binding-types.
// 2. Validates all required identity_fields are present in --identity.
// 3. POSTs to /api/trust-bindings.

type tbEnableCmd struct {
	Key          string `arg:"" help:"Binding type catalog key (e.g. terraform_cloud, github_actions)."`
	Identity     string `name:"identity" required:"" help:"Identity fields as comma-separated key=value pairs (e.g. terraform_organization_name=acme,terraform_workspace_name=prod)."`
	OutputClaims string `name:"output-claims" help:"Additional output claims as comma-separated key=value pairs."`
}

func (c *tbEnableCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()

	// Resolve the catalog row by key (list + filter client-side).
	defs, err := api.ListTrustBindingTypeDefs(ctx)
	if err != nil {
		return fmt.Errorf("lookup binding type %q: %w", c.Key, err)
	}

	var matched *client.TrustBindingTypeDef
	for i := range defs {
		if defs[i].Key == c.Key {
			matched = &defs[i]
			break
		}
	}
	if matched == nil {
		return fmt.Errorf("unknown binding type %q: not found in catalog (run 'binding-types list' to see available types)", c.Key)
	}

	// Parse identity claims.
	identity, err := parseKVMap(c.Identity)
	if err != nil {
		return fmt.Errorf("--identity: %w", err)
	}

	// Client-side validation: check all required identity_fields are present.
	var missing []string
	for _, field := range matched.IdentityFields {
		if _, ok := identity[field]; !ok {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required identity field(s) for %q: %s\n(required: %s)",
			c.Key,
			strings.Join(missing, ", "),
			strings.Join(matched.IdentityFields, ", "),
		)
	}

	// Parse optional output claims.
	outputClaims, err := parseKVMap(c.OutputClaims)
	if err != nil {
		return fmt.Errorf("--output-claims: %w", err)
	}

	in := client.TrustBindingInput{
		BindingType:  c.Key,
		Identity:     identity,
		OutputClaims: outputClaims,
	}

	binding, err := api.CreateTrustBinding(ctx, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}
