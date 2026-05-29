package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// PermissionSetsCmd is the parent command for permission set management.
type PermissionSetsCmd struct {
	List   psListCmd   `cmd:"" help:"Search permission sets."`
	Create psCreateCmd `cmd:"" help:"Create a permission set."`
	Update psUpdateCmd `cmd:"" help:"Update a permission set."`
	Delete psDeleteCmd `cmd:"" help:"Delete a permission set."`
}

// ── list ──────────────────────────────────────────────────────────────────

type psListCmd struct{ listFlags }

func (c *psListCmd) Run(g *Globals) error {
	page, err := g.Client().SearchPermissionSets(context.Background(), c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, p := range page.Data {
		rows[i] = []string{p.ID, p.Name, fmt.Sprintf("%d", len(p.Permissions))}
	}
	output.PrintTable([]string{"ID", "Name", "#Permissions"}, rows, false)
	return nil
}

// ── create ────────────────────────────────────────────────────────────────

type psCreateCmd struct {
	Name        string   `help:"Name." required:""`
	Description string   `help:"Description."`
	Permission  []string `name:"permission" help:"Permission spec: name:label[:dangerous]. Repeatable."`
}

func (c *psCreateCmd) Run(g *Globals) error {
	perms, err := parsePermissions(c.Permission)
	if err != nil {
		return err
	}
	in := client.PermissionSetInput{
		Name:        c.Name,
		Description: c.Description,
		Permissions: perms,
	}
	ps, err := g.Client().CreatePermissionSet(context.Background(), in)
	if err != nil {
		return err
	}
	return output.PrintJSON(ps)
}

// ── update ────────────────────────────────────────────────────────────────

type psUpdateCmd struct {
	ID string `arg:"" help:"Permission set ID."`
	psCreateCmd
}

func (c *psUpdateCmd) Run(g *Globals) error {
	perms, err := parsePermissions(c.Permission)
	if err != nil {
		return err
	}
	in := client.PermissionSetInput{
		Name:        c.Name,
		Description: c.Description,
		Permissions: perms,
	}
	ps, err := g.Client().UpdatePermissionSet(context.Background(), c.ID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(ps)
}

// ── delete ────────────────────────────────────────────────────────────────

type psDeleteCmd struct {
	ID string `arg:"" help:"Permission set ID."`
}

func (c *psDeleteCmd) Run(g *Globals) error {
	if err := g.Client().DeletePermissionSet(context.Background(), c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────

// parsePermissions converts a slice of "name:label[:dangerous]" specs into
// []client.PermissionInput. Returns an error if any spec is missing name or label.
func parsePermissions(specs []string) ([]client.PermissionInput, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	out := make([]client.PermissionInput, 0, len(specs))
	for _, s := range specs {
		parts := strings.SplitN(s, ":", 3)
		if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("invalid permission spec %q: expected name:label[:dangerous]", s)
		}
		p := client.PermissionInput{
			Name:  strings.TrimSpace(parts[0]),
			Label: strings.TrimSpace(parts[1]),
		}
		if len(parts) == 3 {
			p.Dangerous = strings.TrimSpace(parts[2]) == "true"
		}
		out = append(out, p)
	}
	return out, nil
}
