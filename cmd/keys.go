package cmd

// KeysCmd is the parent command for issued key management.
// Subcommand structs are skeletons — bodies filled in Task 10.
type KeysCmd struct {
	List   keysListCmd   `cmd:"" help:"Search issued keys."`
	Get    keysGetCmd    `cmd:"" help:"Get an issued key."`
	Update keysUpdateCmd `cmd:"" help:"Update a key."`
	Revoke keysRevokeCmd `cmd:"" help:"Revoke a key."`
	Rotate keysRotateCmd `cmd:"" help:"Rotate a key."`
	Events keysEventsCmd `cmd:"" help:"List key events."`
	Verify keysVerifyCmd `cmd:"" help:"Verify a key."`
}

type keysListCmd struct{}

func (c *keysListCmd) Run(g *Globals) error {
	return notImplementedError{"keys list"}
}

type keysGetCmd struct {
	ID string `arg:"" help:"Key ID."`
}

func (c *keysGetCmd) Run(g *Globals) error {
	return notImplementedError{"keys get"}
}

type keysUpdateCmd struct {
	ID string `arg:"" help:"Key ID."`
}

func (c *keysUpdateCmd) Run(g *Globals) error {
	return notImplementedError{"keys update"}
}

type keysRevokeCmd struct {
	ID string `arg:"" help:"Key ID."`
}

func (c *keysRevokeCmd) Run(g *Globals) error {
	return notImplementedError{"keys revoke"}
}

type keysRotateCmd struct {
	ID string `arg:"" help:"Key ID."`
}

func (c *keysRotateCmd) Run(g *Globals) error {
	return notImplementedError{"keys rotate"}
}

type keysEventsCmd struct {
	ID string `arg:"" help:"Key ID."`
}

func (c *keysEventsCmd) Run(g *Globals) error {
	return notImplementedError{"keys events"}
}

type keysVerifyCmd struct {
	Key string `arg:"" help:"Key to verify."`
}

func (c *keysVerifyCmd) Run(g *Globals) error {
	return notImplementedError{"keys verify"}
}
