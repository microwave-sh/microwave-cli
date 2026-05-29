package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

// KeySpecsCmd is the parent command for key-spec management.
type KeySpecsCmd struct {
	List          ksListCmd          `cmd:"" help:"Search key specs."`
	Create        ksCreateCmd        `cmd:"" help:"Create a key spec."`
	Update        ksUpdateCmd        `cmd:"" help:"Update a key spec."`
	Delete        ksDeleteCmd        `cmd:"" help:"Delete a key spec."`
	Events        ksEventsCmd        `cmd:"" help:"List key-spec events."`
	WidgetSession ksWidgetSessionCmd `cmd:"" name:"widget-session" help:"Create a widget session token."`
	Keys          ksKeysCmd          `cmd:"" help:"Manage keys for a key spec."`
}

// ── list ──────────────────────────────────────────────────────────────

type ksListCmd struct {
	listFlags
	Format string `name:"format" help:"Filter by format (opaque|jwt)."`
}

func (c *ksListCmd) Run(g *Globals) error {
	var filter map[string]map[string]any
	if c.Format != "" {
		filter = map[string]map[string]any{
			"format": {"eq": c.Format},
		}
	}
	page, err := g.Client().SearchKeySpecs(context.Background(), c.searchRequest(filter))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, s := range page.Data {
		rows[i] = []string{s.ID, s.Name, s.Format, output.FormatTimeAgo(s.CreatedAt)}
	}
	output.PrintTable([]string{"ID", "Name", "Format", "Created"}, rows, false)
	return nil
}

// ── create / update shared flags ──────────────────────────────────────

type ksSpecFlags struct {
	Name                   string `help:"Name." required:""`
	Description            string `help:"Description."`
	Format                 string `help:"Key format (opaque|jwt)." required:""`
	PermissionSetID        string `name:"permission-set-id" help:"Permission set ID."`
	SigningKeySetID        string `name:"signing-key-set-id" help:"Signing key set ID."`
	JWTAlgorithm           string `name:"jwt-algorithm" help:"JWT algorithm (RS256|ES256|HS256)."`
	JWTIssuer              string `name:"jwt-issuer" help:"JWT issuer claim."`
	JWTAudience            string `name:"jwt-audience" help:"JWT audience claim."`
	DefaultTTL             string `name:"default-ttl" help:"Default key TTL (e.g. 24h)."`
	MaxTTL                 string `name:"max-ttl" help:"Maximum key TTL."`
	AllowNever             bool   `name:"allow-never" help:"Allow keys that never expire."`
	RotationReminderDays   int    `name:"rotation-reminder-days" help:"Days before expiry to send rotation reminder."`
	StandardClaims         string `name:"standard-claims" help:"Comma-separated standard JWT claim names."`
	AllowCustomExpiry      bool   `name:"allow-custom-expiry" help:"Allow issuers to override expiry."`
	AllowCustomScopes      bool   `name:"allow-custom-scopes" help:"Allow issuers to override scopes."`
	AllowCustomClaims      bool   `name:"allow-custom-claims" help:"Allow issuers to override claims."`
	WebhookEndpoint        string `name:"webhook-endpoint" help:"Webhook delivery URL."`
	WebhookEvents          string `name:"webhook-events" help:"Comma-separated webhook event types."`
	WebhookSigningKeySetID string `name:"webhook-signing-key-set-id" help:"Signing key set ID for webhook signatures."`
	OpaquePrefix           string `name:"opaque-prefix" help:"Opaque key prefix."`
	OpaqueLookupResponse   string `name:"opaque-lookup-response" help:"Opaque lookup response template."`
}

// toInput converts the flag values into a KeySpecInput, building all nested structs.
func (c *ksSpecFlags) toInput() (client.KeySpecInput, error) {
	return client.KeySpecInput{
		Name:            c.Name,
		Description:     c.Description,
		Format:          c.Format,
		PermissionSetID: c.PermissionSetID,
		SigningKeySetID: c.SigningKeySetID,
		Opaque: client.OpaqueConfig{
			Prefix:         c.OpaquePrefix,
			LookupResponse: c.OpaqueLookupResponse,
		},
		JWT: client.JWTConfig{
			Algorithm: c.JWTAlgorithm,
			Issuer:    c.JWTIssuer,
			Audience:  c.JWTAudience,
		},
		Expiry: client.ExpiryPolicy{
			DefaultTTL:           c.DefaultTTL,
			MaxTTL:               c.MaxTTL,
			AllowNever:           c.AllowNever,
			RotationReminderDays: c.RotationReminderDays,
		},
		Claims: client.ClaimsConfig{
			Standard: parseCSV(c.StandardClaims),
		},
		OverridePolicy: client.OverridePolicy{
			AllowCustomExpiry: c.AllowCustomExpiry,
			AllowCustomScopes: c.AllowCustomScopes,
			AllowCustomClaims: c.AllowCustomClaims,
		},
		Webhooks: client.WebhookConfig{
			Endpoint: c.WebhookEndpoint,
			Events:   parseCSV(c.WebhookEvents),
		},
		WebhookSigningKeySetID: c.WebhookSigningKeySetID,
	}, nil
}

// ── create ────────────────────────────────────────────────────────────

type ksCreateCmd struct {
	ksSpecFlags
}

func (c *ksCreateCmd) Run(g *Globals) error {
	in, err := c.toInput()
	if err != nil {
		return err
	}
	spec, err := g.Client().CreateKeySpec(context.Background(), in)
	if err != nil {
		return err
	}
	return output.PrintJSON(spec)
}

// ── update ────────────────────────────────────────────────────────────

type ksUpdateCmd struct {
	ID string `arg:"" help:"Key spec ID."`
	ksSpecFlags
}

func (c *ksUpdateCmd) Run(g *Globals) error {
	in, err := c.toInput()
	if err != nil {
		return err
	}
	spec, err := g.Client().UpdateKeySpec(context.Background(), c.ID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(spec)
}

// ── delete ────────────────────────────────────────────────────────────

type ksDeleteCmd struct {
	ID string `arg:"" help:"Key spec ID."`
}

func (c *ksDeleteCmd) Run(g *Globals) error {
	if err := g.Client().DeleteKeySpec(context.Background(), c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}

// ── events ────────────────────────────────────────────────────────────

type ksEventsCmd struct {
	ID      string `arg:"" help:"Key spec ID."`
	Subject string `name:"subject" help:"Filter by subject."`
}

func (c *ksEventsCmd) Run(g *Globals) error {
	events, err := g.Client().KeySpecEvents(context.Background(), c.ID, c.Subject)
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

// ── widget-session ────────────────────────────────────────────────────

type ksWidgetSessionCmd struct {
	ID          string `arg:"" help:"Key spec ID."`
	Subject     string `name:"subject" help:"Session subject." required:""`
	Scopes      string `name:"scopes" help:"Comma-separated scopes."`
	Claims      string `name:"claims" help:"Claims as JSON object."`
	RedirectURL string `name:"redirect-url" help:"Post-session redirect URL."`
	TTL         string `name:"ttl" help:"Session TTL (e.g. 1h)."`
}

func (c *ksWidgetSessionCmd) Run(g *Globals) error {
	claims, err := parseJSONMap(c.Claims)
	if err != nil {
		return err
	}
	in := client.WidgetSessionInput{
		Subject:     c.Subject,
		Scopes:      parseCSV(c.Scopes),
		Claims:      claims,
		RedirectURL: c.RedirectURL,
		TTL:         c.TTL,
	}
	tok, err := g.Client().CreateWidgetSession(context.Background(), c.ID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(tok)
}

// ── keys nested group ─────────────────────────────────────────────────

type ksKeysCmd struct {
	Issue           ksKeysIssueCmd           `cmd:"" help:"Issue a key from a spec."`
	List            ksKeysListCmd            `cmd:"" help:"Search keys for a spec."`
	RevokeBySubject ksKeysRevokeBySubjectCmd `cmd:"" name:"revoke-by-subject" help:"Revoke all keys for a subject."`
}

// ── keys issue ────────────────────────────────────────────────────────

type ksKeysIssueCmd struct {
	SpecID    string `arg:"" name:"spec-id" help:"Key spec ID."`
	Subject   string `name:"subject" help:"Key subject." required:""`
	Name      string `name:"name" help:"Key name." required:""`
	Scopes    string `name:"scopes" help:"Comma-separated scopes."`
	Claims    string `name:"claims" help:"Claims as JSON object."`
	Metadata  string `name:"metadata" help:"Metadata as JSON object."`
	ExpiresIn string `name:"expires-in" help:"Key TTL (e.g. 90d)."`
}

func (c *ksKeysIssueCmd) Run(g *Globals) error {
	claims, err := parseJSONMap(c.Claims)
	if err != nil {
		return err
	}
	metadata, err := parseJSONMap(c.Metadata)
	if err != nil {
		return err
	}
	in := client.IssueKeyInput{
		Subject:   c.Subject,
		Name:      c.Name,
		Scopes:    parseCSV(c.Scopes),
		Claims:    claims,
		Metadata:  metadata,
		ExpiresIn: c.ExpiresIn,
	}
	result, err := g.Client().IssueKey(context.Background(), c.SpecID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(result)
}

// ── keys list ─────────────────────────────────────────────────────────

type ksKeysListCmd struct {
	SpecID string `arg:"" name:"spec-id" help:"Key spec ID."`
	listFlags
}

func (c *ksKeysListCmd) Run(g *Globals) error {
	page, err := g.Client().SearchSpecKeys(context.Background(), c.SpecID, c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, k := range page.Data {
		rows[i] = []string{k.ID, k.Subject, k.Status, output.FormatTimeAgo(k.CreatedAt)}
	}
	output.PrintTable([]string{"ID", "Subject", "Status", "Created"}, rows, false)
	return nil
}

// ── keys revoke-by-subject ────────────────────────────────────────────

type ksKeysRevokeBySubjectCmd struct {
	SpecID  string `arg:"" name:"spec-id" help:"Key spec ID."`
	Subject string `name:"subject" help:"Subject whose keys to revoke." required:""`
}

func (c *ksKeysRevokeBySubjectCmd) Run(g *Globals) error {
	result, err := g.Client().RevokeKeysBySubject(context.Background(), c.SpecID, c.Subject)
	if err != nil {
		return err
	}
	return output.PrintJSON(result)
}
