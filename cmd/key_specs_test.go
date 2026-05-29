package cmd

import (
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

// TestKSCreate_ToInput verifies that ksSpecFlags.toInput() populates Format and
// all nested config structs (JWT, Expiry, OverridePolicy, Webhooks).
func TestKSCreate_ToInput(t *testing.T) {
	flags := ksSpecFlags{
		Name:                   "my-spec",
		Description:            "test spec",
		Format:                 "jwt",
		PermissionSetID:        "ps_1",
		SigningKeySetID:         "sks_1",
		JWTAlgorithm:           "RS256",
		JWTIssuer:              "https://issuer.example.com",
		JWTAudience:            "api://myapp",
		DefaultTTL:             "24h",
		MaxTTL:                 "720h",
		AllowNever:             true,
		RotationReminderDays:   30,
		StandardClaims:         "sub, email",
		AllowCustomExpiry:      true,
		AllowCustomScopes:      false,
		AllowCustomClaims:      true,
		WebhookEndpoint:        "https://hooks.example.com/akaas",
		WebhookEvents:          "key.issued, key.revoked",
		WebhookSigningKeySetID: "sks_webhook",
		OpaquePrefix:           "",
		OpaqueLookupResponse:   "",
	}

	in, err := flags.toInput()
	if err != nil {
		t.Fatalf("toInput() error: %v", err)
	}

	// Top-level fields
	if in.Format != "jwt" {
		t.Errorf("Format = %q, want jwt", in.Format)
	}
	if in.Name != "my-spec" {
		t.Errorf("Name = %q, want my-spec", in.Name)
	}
	if in.PermissionSetID != "ps_1" {
		t.Errorf("PermissionSetID = %q, want ps_1", in.PermissionSetID)
	}
	if in.SigningKeySetID != "sks_1" {
		t.Errorf("SigningKeySetID = %q, want sks_1", in.SigningKeySetID)
	}

	// JWT config
	if in.JWT.Algorithm != "RS256" {
		t.Errorf("JWT.Algorithm = %q, want RS256", in.JWT.Algorithm)
	}
	if in.JWT.Issuer != "https://issuer.example.com" {
		t.Errorf("JWT.Issuer = %q, want https://issuer.example.com", in.JWT.Issuer)
	}
	if in.JWT.Audience != "api://myapp" {
		t.Errorf("JWT.Audience = %q, want api://myapp", in.JWT.Audience)
	}

	// Expiry policy
	if in.Expiry.DefaultTTL != "24h" {
		t.Errorf("Expiry.DefaultTTL = %q, want 24h", in.Expiry.DefaultTTL)
	}
	if in.Expiry.MaxTTL != "720h" {
		t.Errorf("Expiry.MaxTTL = %q, want 720h", in.Expiry.MaxTTL)
	}
	if !in.Expiry.AllowNever {
		t.Error("Expiry.AllowNever = false, want true")
	}
	if in.Expiry.RotationReminderDays != 30 {
		t.Errorf("Expiry.RotationReminderDays = %d, want 30", in.Expiry.RotationReminderDays)
	}

	// Claims config — standard claims parsed from CSV
	if len(in.Claims.Standard) != 2 {
		t.Errorf("Claims.Standard = %v, want 2 entries", in.Claims.Standard)
	}

	// Override policy
	if !in.OverridePolicy.AllowCustomExpiry {
		t.Error("OverridePolicy.AllowCustomExpiry = false, want true")
	}
	if in.OverridePolicy.AllowCustomScopes {
		t.Error("OverridePolicy.AllowCustomScopes = true, want false")
	}
	if !in.OverridePolicy.AllowCustomClaims {
		t.Error("OverridePolicy.AllowCustomClaims = false, want true")
	}

	// Webhooks
	if in.Webhooks.Endpoint != "https://hooks.example.com/akaas" {
		t.Errorf("Webhooks.Endpoint = %q, want https://hooks.example.com/akaas", in.Webhooks.Endpoint)
	}
	if len(in.Webhooks.Events) != 2 {
		t.Errorf("Webhooks.Events = %v, want 2 entries", in.Webhooks.Events)
	}
	if in.WebhookSigningKeySetID != "sks_webhook" {
		t.Errorf("WebhookSigningKeySetID = %q, want sks_webhook", in.WebhookSigningKeySetID)
	}
}

// TestKSCreate_OpaqueFormat verifies that opaque-specific fields are mapped.
func TestKSCreate_OpaqueFormat(t *testing.T) {
	flags := ksSpecFlags{
		Name:                 "opaque-spec",
		Format:               "opaque",
		OpaquePrefix:         "tok_",
		OpaqueLookupResponse: "full",
	}
	in, err := flags.toInput()
	if err != nil {
		t.Fatalf("toInput() error: %v", err)
	}
	if in.Format != "opaque" {
		t.Errorf("Format = %q, want opaque", in.Format)
	}
	if in.Opaque.Prefix != "tok_" {
		t.Errorf("Opaque.Prefix = %q, want tok_", in.Opaque.Prefix)
	}
	if in.Opaque.LookupResponse != "full" {
		t.Errorf("Opaque.LookupResponse = %q, want full", in.Opaque.LookupResponse)
	}
}

// TestKSCreate_EmptyStandardClaims ensures nil (not empty slice) for empty CSV.
func TestKSCreate_EmptyStandardClaims(t *testing.T) {
	flags := ksSpecFlags{Name: "s", Format: "opaque"}
	in, err := flags.toInput()
	if err != nil {
		t.Fatalf("toInput() error: %v", err)
	}
	if in.Claims.Standard != nil {
		t.Errorf("Claims.Standard = %v, want nil for empty input", in.Claims.Standard)
	}
}

// TestKSIssue_BuildsIssueKeyInput verifies IssueKeyInput fields are set correctly.
func TestKSIssue_BuildsIssueKeyInput(t *testing.T) {
	cmd := ksKeysIssueCmd{
		SpecID:    "ks_1",
		Subject:   "user_abc",
		Name:      "my-key",
		Scopes:    "read:data, write:data",
		Claims:    `{"role":"admin"}`,
		Metadata:  `{"env":"prod"}`,
		ExpiresIn: "30d",
	}

	claims, err := parseJSONMap(cmd.Claims)
	if err != nil {
		t.Fatalf("parseJSONMap claims: %v", err)
	}
	metadata, err := parseJSONMap(cmd.Metadata)
	if err != nil {
		t.Fatalf("parseJSONMap metadata: %v", err)
	}

	in := client.IssueKeyInput{
		Subject:   cmd.Subject,
		Name:      cmd.Name,
		Scopes:    parseCSV(cmd.Scopes),
		Claims:    claims,
		Metadata:  metadata,
		ExpiresIn: cmd.ExpiresIn,
	}

	if in.Subject != "user_abc" {
		t.Errorf("Subject = %q, want user_abc", in.Subject)
	}
	if in.Name != "my-key" {
		t.Errorf("Name = %q, want my-key", in.Name)
	}
	if len(in.Scopes) != 2 {
		t.Errorf("Scopes = %v, want 2 entries", in.Scopes)
	}
	if in.Claims["role"] != "admin" {
		t.Errorf("Claims[role] = %v, want admin", in.Claims["role"])
	}
	if in.Metadata["env"] != "prod" {
		t.Errorf("Metadata[env] = %v, want prod", in.Metadata["env"])
	}
	if in.ExpiresIn != "30d" {
		t.Errorf("ExpiresIn = %q, want 30d", in.ExpiresIn)
	}
}
