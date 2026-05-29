package cmd

// SigningKeySetsCmd is the parent command for signing key set management.
// Subcommand structs are skeletons — bodies filled in Task 11.
type SigningKeySetsCmd struct {
	List         signingKeySetsListCmd         `cmd:"" help:"Search signing key sets."`
	Get          signingKeySetsGetCmd          `cmd:"" help:"Get a signing key set."`
	Create       signingKeySetsCreateCmd       `cmd:"" help:"Create a signing key set."`
	Update       signingKeySetsUpdateCmd       `cmd:"" help:"Rename a signing key set."`
	Delete       signingKeySetsDeleteCmd       `cmd:"" help:"Delete a signing key set."`
	Sign         signingKeySetsSignCmd         `cmd:"" help:"Sign a JWT payload (asymmetric)."`
	Secret       signingKeySetsSecretCmd       `cmd:"" help:"Reveal symmetric secret state."`
	RotateSecret signingKeySetsRotateSecretCmd `cmd:"" name:"rotate-secret" help:"Rotate symmetric secret."`
	Keys         signingKeySetsKeysCmd         `cmd:"" help:"Manage signing keys within a set."`
}

type signingKeySetsListCmd struct{}

func (c *signingKeySetsListCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets list"}
}

type signingKeySetsGetCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsGetCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets get"}
}

type signingKeySetsCreateCmd struct{}

func (c *signingKeySetsCreateCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets create"}
}

type signingKeySetsUpdateCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsUpdateCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets update"}
}

type signingKeySetsDeleteCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsDeleteCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets delete"}
}

type signingKeySetsSignCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsSignCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets sign"}
}

type signingKeySetsSecretCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsSecretCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets secret"}
}

type signingKeySetsRotateSecretCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsRotateSecretCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets rotate-secret"}
}

// signingKeySetsKeysCmd is the nested group for signing key lifecycle within a set.
type signingKeySetsKeysCmd struct {
	Generate signingKeySetsKeysGenerateCmd `cmd:"" help:"Generate a new signing key."`
	Activate signingKeySetsKeysActivateCmd `cmd:"" help:"Activate a signing key (symmetric)."`
	Revoke   signingKeySetsKeysRevokeCmd   `cmd:"" help:"Revoke a signing key."`
	Secret   signingKeySetsKeysSecretCmd   `cmd:"" help:"Reveal a key secret (symmetric)."`
}

type signingKeySetsKeysGenerateCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsKeysGenerateCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets keys generate"}
}

type signingKeySetsKeysActivateCmd struct {
	Kind  string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name  string `arg:"" help:"Key set name."`
	KeyID string `arg:"" name:"key-id" help:"Key ID to activate."`
}

func (c *signingKeySetsKeysActivateCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets keys activate"}
}

type signingKeySetsKeysRevokeCmd struct {
	Kind  string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name  string `arg:"" help:"Key set name."`
	KeyID string `arg:"" name:"key-id" help:"Key ID to revoke."`
}

func (c *signingKeySetsKeysRevokeCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets keys revoke"}
}

type signingKeySetsKeysSecretCmd struct {
	Kind  string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name  string `arg:"" help:"Key set name."`
	KeyID string `arg:"" name:"key-id" help:"Key ID."`
}

func (c *signingKeySetsKeysSecretCmd) Run(g *Globals) error {
	return notImplementedError{"signing-key-sets keys secret"}
}
