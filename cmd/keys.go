package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// KeysCmd is the parent command for issued key management.
type KeysCmd struct {
	List   keysListCmd   `cmd:"" help:"Search issued keys."`
	Get    keysGetCmd    `cmd:"" help:"Get an issued key."`
	Update keysUpdateCmd `cmd:"" help:"Update a key."`
	Revoke keysRevokeCmd `cmd:"" help:"Revoke a key."`
	Rotate keysRotateCmd `cmd:"" help:"Rotate a key."`
	Events keysEventsCmd `cmd:"" help:"List key events."`
	Verify keysVerifyCmd `cmd:"" help:"Verify a key."`
}

// ── list ─────────────────────────────────────────────────────────────────

type keysListCmd struct {
	listFlags
	SpecID  string `name:"spec-id" help:"Filter by key spec ID."`
	Subject string `help:"Filter by subject."`
	Status  string `help:"Filter by status (active, revoked, expired, rotating)."`
}

// buildKeysFilter constructs the filter map from the set flag values.
// Only entries for non-empty values are included.
func buildKeysFilter(specID, subject, status string) map[string]map[string]any {
	filter := map[string]map[string]any{}
	if specID != "" {
		filter["spec_id"] = map[string]any{"eq": specID}
	}
	if subject != "" {
		filter["subject"] = map[string]any{"eq": subject}
	}
	if status != "" {
		filter["status"] = map[string]any{"eq": status}
	}
	if len(filter) == 0 {
		return nil
	}
	return filter
}

func (c *keysListCmd) Run(g *Globals) error {
	filter := buildKeysFilter(c.SpecID, c.Subject, c.Status)
	page, err := g.Client().SearchKeys(context.Background(), c.searchRequest(filter))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, k := range page.Data {
		rows[i] = []string{k.ID, k.SpecID, k.Subject, output.ColorStatus(k.Status), output.FormatTimeAgo(k.CreatedAt)}
	}
	output.PrintTable([]string{"ID", "Spec", "Subject", "Status", "Created"}, rows, false)
	return nil
}

// ── get ──────────────────────────────────────────────────────────────────

type keysGetCmd struct {
	ID string `arg:"" help:"Key ID."`
}

func (c *keysGetCmd) Run(g *Globals) error {
	k, err := g.Client().GetKey(context.Background(), c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(k)
}

// ── update ───────────────────────────────────────────────────────────────

type keysUpdateCmd struct {
	ID        string `arg:"" help:"Key ID."`
	Name      string `help:"New name for the key."`
	Scopes    string `help:"Comma-separated scopes."`
	ExpiresAt string `name:"expires-at" help:"Expiry time in RFC3339 format (e.g. 2026-01-01T00:00:00Z)."`
	Metadata  string `help:"Metadata as a JSON object."`
}

// parseExpiresAt parses an RFC3339 string into a *time.Time.
// Returns nil, nil when s is empty.
func parseExpiresAt(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, fmt.Errorf("--expires-at: invalid RFC3339 timestamp %q: %w", s, err)
	}
	return &t, nil
}

func (c *keysUpdateCmd) Run(g *Globals) error {
	exp, err := parseExpiresAt(c.ExpiresAt)
	if err != nil {
		return err
	}
	meta, err := parseJSONMap(c.Metadata)
	if err != nil {
		return err
	}
	in := client.UpdateKeyInput{
		Name:      c.Name,
		Scopes:    parseCSV(c.Scopes),
		ExpiresAt: exp,
		Metadata:  meta,
	}
	k, err := g.Client().UpdateKey(context.Background(), c.ID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(k)
}

// ── revoke ───────────────────────────────────────────────────────────────

type keysRevokeCmd struct {
	ID string `arg:"" help:"Key ID."`
}

func (c *keysRevokeCmd) Run(g *Globals) error {
	k, err := g.Client().RevokeKey(context.Background(), c.ID)
	if err != nil {
		return err
	}
	fmt.Printf("%s Revoked %s\n", output.Green.Render("✓"), c.ID)
	return output.PrintJSON(k)
}

// ── rotate ───────────────────────────────────────────────────────────────

type keysRotateCmd struct {
	ID             string `arg:"" help:"Key ID."`
	OverlapSeconds int    `name:"overlap-seconds" help:"Seconds the old key remains valid after rotation." default:"0"`
}

func (c *keysRotateCmd) Run(g *Globals) error {
	res, err := g.Client().RotateKey(context.Background(), c.ID, client.RotateKeyInput{
		OverlapSeconds: c.OverlapSeconds,
	})
	if err != nil {
		return err
	}
	return output.PrintJSON(res)
}

// ── events ───────────────────────────────────────────────────────────────

type keysEventsCmd struct {
	ID string `arg:"" help:"Key ID."`
	listFlags
}

func (c *keysEventsCmd) Run(g *Globals) error {
	events, err := g.Client().KeyEvents(context.Background(), c.ID)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(events)
	}
	rows := make([][]string, len(events))
	for i, e := range events {
		rows[i] = []string{e.Type, e.Subject, e.Actor, output.FormatTimeAgo(e.Timestamp)}
	}
	output.PrintTable([]string{"Type", "Subject", "Actor", "When"}, rows, false)
	return nil
}

// ── verify ───────────────────────────────────────────────────────────────

type keysVerifyCmd struct {
	Key string `arg:"" help:"Key string to verify."`
}

func (c *keysVerifyCmd) Run(g *Globals) error {
	result, err := g.Client().VerifyKey(context.Background(), c.Key)
	if err != nil {
		return err
	}
	// verify is informational — always print the result, even when !result.Valid.
	return output.PrintJSON(result)
}
