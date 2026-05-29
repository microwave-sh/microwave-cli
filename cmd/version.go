package cmd

import (
	"context"
	"fmt"
)

type VersionCmd struct{}

func (c *VersionCmd) Run(_ context.Context, g *Globals) error {
	fmt.Println(g.Version)
	return nil
}
