package cmd

import (
	"context"
	"fmt"

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

type trustBindingWorkspaceFlag struct {
	WorkspaceID string `name:"workspace-id" env:"MICROWAVE_WORKSPACE" help:"Workspace ID. Defaults to the authenticated workspace."`
}

func (f trustBindingWorkspaceFlag) resolveWorkspaceID(ctx context.Context, c *client.Client) (string, error) {
	if f.WorkspaceID != "" {
		return f.WorkspaceID, nil
	}
	me, err := c.Me(ctx)
	if err != nil {
		return "", err
	}
	if me.WorkspaceID == "" {
		return "", fmt.Errorf("workspace id is required")
	}
	return me.WorkspaceID, nil
}

// ── list ────────────────────────────────────────────────────────────────

type tbListCmd struct {
	trustBindingWorkspaceFlag
}

func (c *tbListCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()
	workspaceID, err := c.resolveWorkspaceID(ctx, api)
	if err != nil {
		return err
	}
	bindings, err := api.ListTrustBindings(ctx, workspaceID)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(bindings)
	}
	rows := make([][]string, len(bindings))
	for i, b := range bindings {
		rows[i] = []string{b.ID, b.BindingType, output.FormatTimeAgo(b.CreatedAt)}
	}
	output.PrintTable([]string{"ID", "Type", "Created"}, rows, false)
	return nil
}

// ── get ─────────────────────────────────────────────────────────────────

type tbGetCmd struct {
	ID string `arg:"" help:"Trust binding ID."`
	trustBindingWorkspaceFlag
}

func (c *tbGetCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()
	workspaceID, err := c.resolveWorkspaceID(ctx, api)
	if err != nil {
		return err
	}
	binding, err := api.GetTrustBinding(ctx, workspaceID, c.ID)
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
	trustBindingWorkspaceFlag
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
	api := g.Client()
	workspaceID, err := c.resolveWorkspaceID(ctx, api)
	if err != nil {
		return err
	}
	binding, err := api.CreateTrustBinding(ctx, workspaceID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}

// ── delete ──────────────────────────────────────────────────────────────

type tbDeleteCmd struct {
	ID string `arg:"" help:"Trust binding ID."`
	trustBindingWorkspaceFlag
}

func (c *tbDeleteCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()
	workspaceID, err := c.resolveWorkspaceID(ctx, api)
	if err != nil {
		return err
	}
	if err := api.DeleteTrustBinding(ctx, workspaceID, c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}

// ── types ───────────────────────────────────────────────────────────────

type tbTypesCmd struct{}

func (c *tbTypesCmd) Run(ctx context.Context, g *Globals) error {
	types, err := g.Client().ListTrustBindingTypes(ctx)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(types)
	}
	rows := make([][]string, len(types))
	for i, t := range types {
		rows[i] = []string{t.Key, t.DisplayName, t.DocsURL}
	}
	output.PrintTable([]string{"Key", "Name", "Docs"}, rows, false)
	return nil
}

// ── enable helpers ─────────────────────────────────────────────────────

type tbEnableCmd struct {
	TerraformCloud tbEnableTerraformCloudCmd `cmd:"" name:"terraform-cloud" help:"Create a Terraform Cloud trust binding."`
	GitHubActions  tbEnableGitHubActionsCmd  `cmd:"" name:"github-actions" help:"Create a GitHub Actions trust binding."`
}

type tbEnableTerraformCloudCmd struct {
	TFCOrganization string `name:"tfc-org" required:"" help:"Terraform Cloud organization name."`
	TFCWorkspace    string `name:"tfc-workspace" required:"" help:"Terraform Cloud workspace name."`
	trustBindingWorkspaceFlag
}

func (c *tbEnableTerraformCloudCmd) toInput() client.TrustBindingInput {
	return client.TrustBindingInput{
		BindingType: "terraform_cloud",
		Identity: map[string]any{
			"terraform_organization_name": c.TFCOrganization,
			"terraform_workspace_name":    c.TFCWorkspace,
		},
	}
}

func (c *tbEnableTerraformCloudCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()
	workspaceID, err := c.resolveWorkspaceID(ctx, api)
	if err != nil {
		return err
	}
	binding, err := api.CreateTrustBinding(ctx, workspaceID, c.toInput())
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}

type tbEnableGitHubActionsCmd struct {
	Repository string `name:"repository" required:"" help:"GitHub repository in owner/repo form."`
	Workflow   string `name:"workflow" required:"" help:"Workflow filename, e.g. deploy.yml."`
	trustBindingWorkspaceFlag
}

func (c *tbEnableGitHubActionsCmd) toInput() client.TrustBindingInput {
	return client.TrustBindingInput{
		BindingType: "github_actions",
		Identity: map[string]any{
			"repository": c.Repository,
			"workflow":   c.Workflow,
		},
	}
}

func (c *tbEnableGitHubActionsCmd) Run(ctx context.Context, g *Globals) error {
	api := g.Client()
	workspaceID, err := c.resolveWorkspaceID(ctx, api)
	if err != nil {
		return err
	}
	binding, err := api.CreateTrustBinding(ctx, workspaceID, c.toInput())
	if err != nil {
		return err
	}
	return output.PrintJSON(binding)
}
