package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// FederationBindingsCmd is the parent command for Trust Federation Binding management.
type FederationBindingsCmd struct {
	List   fbListCmd   `cmd:"" help:"List federation bindings."`
	Get    fbGetCmd    `cmd:"" help:"Get a federation binding."`
	Create fbCreateCmd `cmd:"" help:"Create a federation binding."`
	Delete fbDeleteCmd `cmd:"" help:"Delete a federation binding."`
	Bind   fbBindCmd   `cmd:"" help:"Bind a well-known federation by key (resolves catalog, validates identity fields, creates the binding)."`
}

// ── list ────────────────────────────────────────────────────────────────

type fbListCmd struct {
	listFlags
}

func (c *fbListCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()
	page, err := api.SearchTrustFederationBindings(ctx, c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, b := range page.Data {
		rows[i] = []string{b.ID, b.FederationKey, output.FormatTimeAgo(b.CreatedAt)}
	}
	output.PrintTable([]string{"ID", "Federation", "Created"}, rows, false)
	return nil
}

// ── get ─────────────────────────────────────────────────────────────────

type fbGetCmd struct {
	ID string `arg:"" help:"Federation binding ID."`
}

func (c *fbGetCmd) Run(ctx context.Context, g *Globals) error {
	binding, err := g.Client().GetTrustFederationBinding(ctx, c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}

// ── create ──────────────────────────────────────────────────────────────

type fbCreateCmd struct {
	FederationKey string `arg:"" help:"Federation key."`
	Identity      string `name:"identity" required:"" help:"Identity claims as a JSON object."`
	OutputClaims  string `name:"output-claims" help:"Output claims as a JSON object."`
}

func (c *fbCreateCmd) toInput() (client.TrustFederationBindingInput, error) {
	identity, err := parseJSONMap(c.Identity)
	if err != nil {
		return client.TrustFederationBindingInput{}, err
	}
	outputClaims, err := parseJSONMap(c.OutputClaims)
	if err != nil {
		return client.TrustFederationBindingInput{}, err
	}
	return client.TrustFederationBindingInput{
		FederationKey: c.FederationKey,
		Identity:      identity,
		OutputClaims:  outputClaims,
	}, nil
}

func (c *fbCreateCmd) Run(ctx context.Context, g *Globals) error {
	in, err := c.toInput()
	if err != nil {
		return err
	}
	binding, err := g.Client().CreateTrustFederationBinding(ctx, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}

// ── delete ──────────────────────────────────────────────────────────────

type fbDeleteCmd struct {
	ID string `arg:"" help:"Federation binding ID."`
}

func (c *fbDeleteCmd) Run(ctx context.Context, g *Globals) error {
	if err := g.Client().DeleteTrustFederationBinding(ctx, c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}

// ── bind ────────────────────────────────────────────────────────────────
//
// bind <catalog-key> --identity k=v,k=v [--output-claims k=v,k=v]
//
// 1. Resolves the federation catalog row via GET /api/trust-federations.
// 2. Validates all required identity_fields are present in --identity.
// 3. POSTs to /api/trust-federation-bindings.

type fbBindCmd struct {
	Key          string `arg:"" help:"Federation catalog key (e.g. terraform_cloud, github_actions)."`
	Identity     string `name:"identity" required:"" help:"Identity fields as comma-separated key=value pairs (e.g. terraform_organization_name=acme,terraform_workspace_name=prod)."`
	OutputClaims string `name:"output-claims" help:"Additional output claims as comma-separated key=value pairs."`
}

func (c *fbBindCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()

	// Resolve the catalog row by key (list + filter client-side).
	defs, err := api.ListTrustFederations(ctx)
	if err != nil {
		return fmt.Errorf("lookup federation %q: %w", c.Key, err)
	}

	var matched *client.TrustFederation
	for i := range defs {
		if defs[i].Key == c.Key {
			matched = &defs[i]
			break
		}
	}
	if matched == nil {
		return fmt.Errorf("unknown federation %q: not found in catalog (run 'federations list' to see available federations)", c.Key)
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

	in := client.TrustFederationBindingInput{
		FederationKey: c.Key,
		Identity:      identity,
		OutputClaims:  outputClaims,
	}

	binding, err := api.CreateTrustFederationBinding(ctx, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}
