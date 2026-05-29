package cmd

// TrustExchangesCmd is the parent command for trust exchange management.
// Subcommand structs are skeletons — bodies filled in Task 12.
type TrustExchangesCmd struct {
	List   trustExchangesListCmd   `cmd:"" help:"Search trust exchanges."`
	Get    trustExchangesGetCmd    `cmd:"" help:"Get a trust exchange."`
	Create trustExchangesCreateCmd `cmd:"" help:"Create a trust exchange."`
	Update trustExchangesUpdateCmd `cmd:"" help:"Update a trust exchange."`
	Delete trustExchangesDeleteCmd `cmd:"" help:"Delete a trust exchange."`
}

type trustExchangesListCmd struct{}

func (c *trustExchangesListCmd) Run(g *Globals) error {
	return notImplementedError{"trust-exchanges list"}
}

type trustExchangesGetCmd struct {
	ID string `arg:"" help:"Trust exchange ID."`
}

func (c *trustExchangesGetCmd) Run(g *Globals) error {
	return notImplementedError{"trust-exchanges get"}
}

type trustExchangesCreateCmd struct{}

func (c *trustExchangesCreateCmd) Run(g *Globals) error {
	return notImplementedError{"trust-exchanges create"}
}

type trustExchangesUpdateCmd struct {
	ID string `arg:"" help:"Trust exchange ID."`
}

func (c *trustExchangesUpdateCmd) Run(g *Globals) error {
	return notImplementedError{"trust-exchanges update"}
}

type trustExchangesDeleteCmd struct {
	ID string `arg:"" help:"Trust exchange ID."`
}

func (c *trustExchangesDeleteCmd) Run(g *Globals) error {
	return notImplementedError{"trust-exchanges delete"}
}
