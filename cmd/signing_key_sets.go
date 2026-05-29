package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// SigningKeySetsCmd is the parent command for signing key set management.
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

// ── list ──────────────────────────────────────────────────────────

type signingKeySetsListCmd struct{ listFlags }

func (c *signingKeySetsListCmd) Run(ctx context.Context, g *Globals) error {
	page, err := g.Client().SearchSigningKeySets(ctx, c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, s := range page.Data {
		rows[i] = []string{s.ID, s.Name, s.Kind, s.Algorithm}
	}
	output.PrintTable([]string{"ID", "Name", "Kind", "Algorithm"}, rows, false)
	return nil
}

// ── get ───────────────────────────────────────────────────────────

type signingKeySetsGetCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsGetCmd) Run(ctx context.Context, g *Globals) error {
	detail, err := g.Client().GetSigningKeySet(ctx, c.Kind, c.Name)
	if err != nil {
		return err
	}
	return output.PrintJSON(detail)
}

// ── create ────────────────────────────────────────────────────────

type signingKeySetsCreateCmd struct {
	Kind      string `help:"Key set kind (asymmetric|symmetric)." required:"" enum:"asymmetric,symmetric"`
	Name      string `help:"Key set name." required:""`
	Algorithm string `help:"Signing algorithm (e.g. RS256, ES256, HS256)." required:""`
}

func (c *signingKeySetsCreateCmd) Run(ctx context.Context, g *Globals) error {
	in := client.SigningKeySetInput{
		Name:      c.Name,
		Kind:      c.Kind,
		Algorithm: c.Algorithm,
	}
	s, err := g.Client().CreateSigningKeySet(ctx, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(s)
}

// ── update ────────────────────────────────────────────────────────

type signingKeySetsUpdateCmd struct {
	Kind    string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name    string `arg:"" help:"Key set name."`
	NewName string `name:"name" help:"New name for the key set." required:""`
}

func (c *signingKeySetsUpdateCmd) Run(ctx context.Context, g *Globals) error {
	s, err := g.Client().UpdateSigningKeySet(ctx, c.Kind, c.Name, c.NewName)
	if err != nil {
		return err
	}
	return output.PrintJSON(s)
}

// ── delete ────────────────────────────────────────────────────────

type signingKeySetsDeleteCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsDeleteCmd) Run(ctx context.Context, g *Globals) error {
	if err := g.Client().DeleteSigningKeySet(ctx, c.Kind, c.Name); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s/%s\n", output.Green.Render("✓"), c.Kind, c.Name)
	return nil
}

// ── sign ──────────────────────────────────────────────────────────

type signingKeySetsSignCmd struct {
	Kind    string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name    string `arg:"" help:"Key set name."`
	Payload string `help:"JWT payload claims as a JSON object." required:""`
	KID     string `name:"kid" help:"Key ID hint."`
	Header  string `help:"Additional JWT header fields as a JSON object."`
}

// parseSignJWTInput is a factored helper so tests can exercise the JSON parsing
// without invoking the HTTP client.
func parseSignJWTInput(payload, header, kid string) (client.SignJWTInput, error) {
	p, err := parseJSONMap(payload)
	if err != nil {
		return client.SignJWTInput{}, fmt.Errorf("--payload: %w", err)
	}
	h, err := parseJSONMap(header)
	if err != nil {
		return client.SignJWTInput{}, fmt.Errorf("--header: %w", err)
	}
	return client.SignJWTInput{Payload: p, KID: kid, Header: h}, nil
}

func (c *signingKeySetsSignCmd) Run(ctx context.Context, g *Globals) error {
	in, err := parseSignJWTInput(c.Payload, c.Header, c.KID)
	if err != nil {
		return err
	}
	res, err := g.Client().SignJWT(ctx, c.Kind, c.Name, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(res)
}

// ── secret ────────────────────────────────────────────────────────

type signingKeySetsSecretCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsSecretCmd) Run(ctx context.Context, g *Globals) error {
	state, err := g.Client().SigningKeySetSecret(ctx, c.Kind, c.Name)
	if err != nil {
		return err
	}
	return output.PrintJSON(state)
}

// ── rotate-secret ─────────────────────────────────────────────────

type signingKeySetsRotateSecretCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsRotateSecretCmd) Run(ctx context.Context, g *Globals) error {
	state, err := g.Client().RotateSigningKeySetSecret(ctx, c.Kind, c.Name)
	if err != nil {
		return err
	}
	return output.PrintJSON(state)
}

// ── keys group ────────────────────────────────────────────────────

// signingKeySetsKeysCmd is the nested group for signing key lifecycle within a set.
type signingKeySetsKeysCmd struct {
	Generate signingKeySetsKeysGenerateCmd `cmd:"" help:"Generate a new signing key."`
	Activate signingKeySetsKeysActivateCmd `cmd:"" help:"Activate a signing key (symmetric)."`
	Revoke   signingKeySetsKeysRevokeCmd   `cmd:"" help:"Revoke a signing key."`
	Secret   signingKeySetsKeysSecretCmd   `cmd:"" help:"Reveal a key secret (symmetric)."`
}

// ── keys generate ─────────────────────────────────────────────────

type signingKeySetsKeysGenerateCmd struct {
	Kind string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name string `arg:"" help:"Key set name."`
}

func (c *signingKeySetsKeysGenerateCmd) Run(ctx context.Context, g *Globals) error {
	key, err := g.Client().GenerateSigningKey(ctx, c.Kind, c.Name)
	if err != nil {
		return err
	}
	return output.PrintJSON(key)
}

// ── keys activate ─────────────────────────────────────────────────

type signingKeySetsKeysActivateCmd struct {
	Kind  string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name  string `arg:"" help:"Key set name."`
	KeyID string `arg:"" name:"key-id" help:"Key ID to activate."`
}

func (c *signingKeySetsKeysActivateCmd) Run(ctx context.Context, g *Globals) error {
	key, err := g.Client().ActivateSigningKey(ctx, c.Kind, c.Name, c.KeyID)
	if err != nil {
		return err
	}
	return output.PrintJSON(key)
}

// ── keys revoke ───────────────────────────────────────────────────

type signingKeySetsKeysRevokeCmd struct {
	Kind  string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name  string `arg:"" help:"Key set name."`
	KeyID string `arg:"" name:"key-id" help:"Key ID to revoke."`
}

func (c *signingKeySetsKeysRevokeCmd) Run(ctx context.Context, g *Globals) error {
	key, err := g.Client().RevokeSigningKey(ctx, c.Kind, c.Name, c.KeyID)
	if err != nil {
		return err
	}
	return output.PrintJSON(key)
}

// ── keys secret ───────────────────────────────────────────────────

type signingKeySetsKeysSecretCmd struct {
	Kind  string `arg:"" help:"Key set kind (asymmetric|symmetric)."`
	Name  string `arg:"" help:"Key set name."`
	KeyID string `arg:"" name:"key-id" help:"Key ID."`
}

func (c *signingKeySetsKeysSecretCmd) Run(ctx context.Context, g *Globals) error {
	m, err := g.Client().SigningKeySecret(ctx, c.Kind, c.Name, c.KeyID)
	if err != nil {
		return err
	}
	return output.PrintJSON(m)
}
