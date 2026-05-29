package cmd

// TrustProvidersCmd is the parent command for trust provider management.
// Subcommand structs are skeletons — bodies filled in Task 7.
type TrustProvidersCmd struct {
	List      trustProvidersListCmd      `cmd:"" help:"Search trust providers."`
	Get       trustProvidersGetCmd       `cmd:"" help:"Get a trust provider."`
	Create    trustProvidersCreateCmd    `cmd:"" help:"Create a trust provider."`
	Update    trustProvidersUpdateCmd    `cmd:"" help:"Update a trust provider."`
	Delete    trustProvidersDeleteCmd    `cmd:"" help:"Delete a trust provider."`
	Mint      trustProvidersMintCmd      `cmd:"" help:"Mint a token from a trust provider."`
	Discovery trustProvidersDiscoveryCmd `cmd:"" help:"Print federation (discovery/JWKS/token) URLs."`
}

type trustProvidersListCmd struct{}

func (c *trustProvidersListCmd) Run(g *Globals) error {
	return notImplementedError{"trust-providers list"}
}

type trustProvidersGetCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *trustProvidersGetCmd) Run(g *Globals) error {
	return notImplementedError{"trust-providers get"}
}

type trustProvidersCreateCmd struct{}

func (c *trustProvidersCreateCmd) Run(g *Globals) error {
	return notImplementedError{"trust-providers create"}
}

type trustProvidersUpdateCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *trustProvidersUpdateCmd) Run(g *Globals) error {
	return notImplementedError{"trust-providers update"}
}

type trustProvidersDeleteCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *trustProvidersDeleteCmd) Run(g *Globals) error {
	return notImplementedError{"trust-providers delete"}
}

type trustProvidersMintCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *trustProvidersMintCmd) Run(g *Globals) error {
	return notImplementedError{"trust-providers mint"}
}

type trustProvidersDiscoveryCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *trustProvidersDiscoveryCmd) Run(g *Globals) error {
	return notImplementedError{"trust-providers discovery"}
}
