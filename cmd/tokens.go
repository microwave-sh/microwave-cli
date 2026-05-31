package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

type TokensCmd struct {
	Create TokensCreateCmd `cmd:"" help:"Create a management API key."`
	List   TokensListCmd   `cmd:"" help:"List management API keys."`
	Revoke TokensRevokeCmd `cmd:"" help:"Revoke a management API key."`
}

type TokensCreateCmd struct {
	Name   string   `name:"name" required:"" help:"A label for the key."`
	Scopes []string `name:"scope" help:"Permission scope (repeatable). Defaults to full management scopes."`
}

func (c *TokensCreateCmd) Run(ctx context.Context, g *Globals) error {
	res, err := g.Client().CreateManagementKey(ctx, client.ManagementKeyInput{Name: c.Name, Scopes: c.Scopes})
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(res)
	}
	fmt.Printf("%s Created %s\n", output.Green.Render("✓"), res.ID)
	fmt.Printf("\n  %s\n\n", output.Bold.Render(res.Key))
	fmt.Println(output.Dim.Render("  Copy it now — it will not be shown again."))
	return nil
}

type TokensListCmd struct{}

func (c *TokensListCmd) Run(ctx context.Context, g *Globals) error {
	keys, err := g.Client().ListManagementKeys(ctx)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(keys)
	}
	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []string{k.ID, k.Name, k.KeyHint, k.Status, k.CreatedAt})
	}
	output.PrintTable([]string{"ID", "Name", "Hint", "Status", "Created"}, rows, false)
	return nil
}

type TokensRevokeCmd struct {
	ID string `arg:"" help:"Management key id."`
}

func (c *TokensRevokeCmd) Run(ctx context.Context, g *Globals) error {
	if err := g.Client().RevokeManagementKey(ctx, c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Revoked %s\n", output.Green.Render("✓"), c.ID)
	return nil
}
