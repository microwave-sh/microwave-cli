package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// ConnectCmd wires a GitHub spec repo to Microwave and creates a connection.
type ConnectCmd struct {
	Repo           string `help:"GitHub repo to connect (owner/repo)." required:""`
	Branch         string `help:"Branch to track." default:"main"`
	SpecPath       string `name:"spec-path" help:"Path to the OpenAPI spec file within the repo." default:"openapi.yaml"`
	InstallationID int64  `name:"installation-id" help:"GitHub App installation ID for the spec repo." required:""`
}

func (c *ConnectCmd) Run(ctx context.Context, g *Globals) error {
	in := client.ConnectionInput{
		GHInstallationID: c.InstallationID,
		Repo:             c.Repo,
		Branch:           c.Branch,
		SpecPath:         c.SpecPath,
	}
	conn, err := g.Client().CreateConnection(ctx, in)
	if err != nil {
		return err
	}

	if g.IsJSON() {
		return output.PrintJSON(conn)
	}

	fmt.Printf("%s Connection created\n", output.Green.Render("✓"))
	fmt.Printf("  ID:   %s\n", conn.ID)
	fmt.Printf("  Repo: %s  (branch: %s)\n", conn.Repo, conn.Branch)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Ensure the Microwave GitHub App is installed on %s\n", conn.Repo)
	fmt.Printf("  2. Install the GitHub App on each SDK repo listed under your SDK targets\n")
	fmt.Printf("  3. Run: microwave sdk targets list --connection %s\n", conn.ID)
	return nil
}
