package cmd

import "fmt"

type VersionCmd struct{}

func (c *VersionCmd) Run(g *Globals) error {
	fmt.Println(g.Version)
	return nil
}
