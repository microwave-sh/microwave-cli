package cmd

type PermissionSetsCmd struct {
	List   permissionSetsListCmd   `cmd:"" help:"Search permission sets."`
	Create permissionSetsCreateCmd `cmd:"" help:"Create a permission set."`
	Update permissionSetsUpdateCmd `cmd:"" help:"Update a permission set."`
	Delete permissionSetsDeleteCmd `cmd:"" help:"Delete a permission set."`
}

type permissionSetsListCmd struct{}

func (c *permissionSetsListCmd) Run(g *Globals) error {
	return notImplementedError{"permission-sets list"}
}

type permissionSetsCreateCmd struct{}

func (c *permissionSetsCreateCmd) Run(g *Globals) error {
	return notImplementedError{"permission-sets create"}
}

type permissionSetsUpdateCmd struct {
	ID string `arg:"" help:"Permission set ID."`
}

func (c *permissionSetsUpdateCmd) Run(g *Globals) error {
	return notImplementedError{"permission-sets update"}
}

type permissionSetsDeleteCmd struct {
	ID string `arg:"" help:"Permission set ID."`
}

func (c *permissionSetsDeleteCmd) Run(g *Globals) error {
	return notImplementedError{"permission-sets delete"}
}
