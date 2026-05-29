package cmd

// KeySpecsCmd is the parent command for key-spec management.
// Subcommand structs are skeletons — bodies filled in Task 9.
type KeySpecsCmd struct {
	List          keySpecsListCmd          `cmd:"" help:"Search key specs."`
	Create        keySpecsCreateCmd        `cmd:"" help:"Create a key spec."`
	Update        keySpecsUpdateCmd        `cmd:"" help:"Update a key spec."`
	Delete        keySpecsDeleteCmd        `cmd:"" help:"Delete a key spec."`
	Events        keySpecsEventsCmd        `cmd:"" help:"List key-spec events."`
	WidgetSession keySpecsWidgetSessionCmd `cmd:"" name:"widget-session" help:"Create a widget session token."`
	Keys          keySpecsKeysCmd          `cmd:"" help:"Manage keys for a key spec."`
}

type keySpecsListCmd struct{}

func (c *keySpecsListCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs list"}
}

type keySpecsCreateCmd struct{}

func (c *keySpecsCreateCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs create"}
}

type keySpecsUpdateCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *keySpecsUpdateCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs update"}
}

type keySpecsDeleteCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *keySpecsDeleteCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs delete"}
}

type keySpecsEventsCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *keySpecsEventsCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs events"}
}

type keySpecsWidgetSessionCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *keySpecsWidgetSessionCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs widget-session"}
}

// keySpecsKeysCmd is the nested group for key issuance under a spec.
type keySpecsKeysCmd struct {
	Issue           keySpecsKeysIssueCmd           `cmd:"" help:"Issue a key from a spec."`
	List            keySpecsKeysListCmd             `cmd:"" help:"Search keys for a spec."`
	RevokeBySubject keySpecsKeysRevokeBySubjectCmd  `cmd:"" name:"revoke-by-subject" help:"Revoke all keys for a subject."`
}

type keySpecsKeysIssueCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *keySpecsKeysIssueCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs keys issue"}
}

type keySpecsKeysListCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *keySpecsKeysListCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs keys list"}
}

type keySpecsKeysRevokeBySubjectCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *keySpecsKeysRevokeBySubjectCmd) Run(g *Globals) error {
	return notImplementedError{"key-specs keys revoke-by-subject"}
}
