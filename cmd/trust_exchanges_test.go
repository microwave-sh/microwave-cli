package cmd

import (
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

func TestTECreate_ToInput_OIDCType(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "github-exchange",
		Provider:         "github",
		Issuer:           "https://token.actions.githubusercontent.com",
		AllowedAudiences: "api://prod",
		OutputMode:       "claims",
		Active:           true,
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	if in.Type != "oidc" {
		t.Fatalf("Type = %q, want oidc", in.Type)
	}
	if !in.Active {
		t.Fatal("Active = false, want true")
	}
}

func TestTECreate_ToInput_ParsesAllowedAudiencesCSV(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "github",
		Issuer:           "https://example.com",
		AllowedAudiences: "api://prod, api://staging , api://dev",
		OutputMode:       "claims",
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	if len(in.AllowedAudiences) != 3 {
		t.Fatalf("AllowedAudiences = %v (len %d), want 3 entries", in.AllowedAudiences, len(in.AllowedAudiences))
	}
	if in.AllowedAudiences[0] != "api://prod" {
		t.Fatalf("AllowedAudiences[0] = %q, want api://prod", in.AllowedAudiences[0])
	}
	if in.AllowedAudiences[1] != "api://staging" {
		t.Fatalf("AllowedAudiences[1] = %q, want api://staging", in.AllowedAudiences[1])
	}
}

func TestTECreate_ToInput_ParsesClaimRulesJSON(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "github",
		Issuer:           "https://example.com",
		AllowedAudiences: "aud",
		OutputMode:       "claims",
		ClaimRules:       `{"repository":{"equals":"x","required":true}}`,
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	rule, ok := in.ClaimRules["repository"]
	if !ok {
		t.Fatal("ClaimRules[repository] not found")
	}
	if rule.Equals != "x" {
		t.Fatalf("ClaimRules[repository].Equals = %q, want x", rule.Equals)
	}
	if !rule.Required {
		t.Fatal("ClaimRules[repository].Required = false, want true")
	}
}

func TestTECreate_ToInput_ClaimRulesOneOf(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "google",
		Issuer:           "https://accounts.google.com",
		AllowedAudiences: "aud",
		OutputMode:       "jwt",
		ClaimRules:       `{"hd":{"one_of":["acme.com","corp.com"],"required":true}}`,
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	rule := in.ClaimRules["hd"]
	if len(rule.OneOf) != 2 {
		t.Fatalf("ClaimRules[hd].OneOf = %v, want 2 entries", rule.OneOf)
	}
}

func TestTECreate_ToInput_ParsesClaimMappingJSON(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "github",
		Issuer:           "https://example.com",
		AllowedAudiences: "aud",
		OutputMode:       "claims",
		ClaimMapping:     `{"subject_claim":"sub","scopes":["read","write"]}`,
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	if in.ClaimMapping.SubjectClaim != "sub" {
		t.Fatalf("ClaimMapping.SubjectClaim = %q, want sub", in.ClaimMapping.SubjectClaim)
	}
	if len(in.ClaimMapping.Scopes) != 2 {
		t.Fatalf("ClaimMapping.Scopes = %v, want 2 entries", in.ClaimMapping.Scopes)
	}
}

func TestTECreate_ToInput_InvalidClaimRulesJSON(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "github",
		Issuer:           "https://example.com",
		AllowedAudiences: "aud",
		OutputMode:       "claims",
		ClaimRules:       `{not valid json`,
	}
	_, err := c.toInput()
	if err == nil {
		t.Fatal("expected error for invalid --claim-rules JSON")
	}
}

func TestTECreate_ToInput_InvalidClaimMappingJSON(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "github",
		Issuer:           "https://example.com",
		AllowedAudiences: "aud",
		OutputMode:       "claims",
		ClaimMapping:     `{bad json`,
	}
	_, err := c.toInput()
	if err == nil {
		t.Fatal("expected error for invalid --claim-mapping JSON")
	}
}

func TestTECreate_ToInput_SubjectRules(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "github",
		Issuer:           "https://example.com",
		AllowedAudiences: "aud",
		OutputMode:       "claims",
		SubjectExact:     "repo:octocat/hello-world:ref:refs/heads/main",
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	if in.SubjectRules.Exact != "repo:octocat/hello-world:ref:refs/heads/main" {
		t.Fatalf("SubjectRules.Exact = %q", in.SubjectRules.Exact)
	}
	if in.SubjectRules.Prefix != "" {
		t.Fatalf("SubjectRules.Prefix = %q, want empty", in.SubjectRules.Prefix)
	}
}

func TestTECreate_ToInput_EmptyClaimRulesIsNil(t *testing.T) {
	c := trustExchangesCreateCmd{
		Name:             "x",
		Provider:         "github",
		Issuer:           "https://example.com",
		AllowedAudiences: "aud",
		OutputMode:       "claims",
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	if in.ClaimRules != nil {
		t.Fatalf("ClaimRules = %v, want nil for empty input", in.ClaimRules)
	}
}

// compile-time check that toInput exists on trustExchangesCreateCmd (satisfies structural test goal)
var _ = (*trustExchangesCreateCmd).toInput

// Ensure client types are used correctly in the context of our cmd package.
var _ client.TrustExchangeClaimRule
