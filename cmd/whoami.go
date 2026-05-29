package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/output"
)

type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(g *Globals) error {
	token, err := g.resolveToken()
	if err != nil {
		return err
	}
	res, err := g.Client().VerifyKey(context.Background(), token)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(res)
	}
	if !res.Valid {
		return fmt.Errorf("stored key is not valid (%s)", res.Code)
	}
	output.PrintTable(
		[]string{"Subject", "Key ID", "Scopes"},
		[][]string{{res.Subject, res.KeyID, fmt.Sprintf("%v", res.Scopes)}},
		false,
	)
	return nil
}
