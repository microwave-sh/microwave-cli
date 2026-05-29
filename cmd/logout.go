package cmd

import (
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

type LogoutCmd struct{}

func (c *LogoutCmd) Run(g *Globals) error {
	if err := config.ClearAuth(); err != nil {
		return err
	}
	fmt.Printf("%s Logged out.\n", output.Green.Render("✓"))
	return nil
}
