package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/output"
)

type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(ctx context.Context, g *Globals) error {
	me, err := g.Client().Me(ctx)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(me)
	}
	output.PrintTable(
		[]string{"Workspace", "Actor", "Tier", "Permissions"},
		[][]string{{me.WorkspaceID, me.Actor, me.Tier, fmt.Sprintf("%v", me.Permissions)}},
		false,
	)
	return nil
}
