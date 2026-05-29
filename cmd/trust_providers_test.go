package cmd

import (
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

func TestTPCreate_BuildsOIDCInput(t *testing.T) {
	c := tpCreateCmd{
		Name: "deploy", SigningKeySetID: "sks_1",
		AllowedAudiences: "api://prod, api://staging", DefaultAudience: "api://prod",
		AllowedClaims: "role", RequiredClaims: "role", ConstantClaims: `{"tenant":"acme"}`,
		SubjectRequired: true, TTLDefault: 900, TTLMax: 3600,
	}
	constant, err := parseJSONMap(c.ConstantClaims)
	if err != nil {
		t.Fatalf("parseJSONMap: %v", err)
	}
	in := client.TrustProviderInput{
		Name:             c.Name,
		Type:             "oidc",
		SigningKeySetID:  c.SigningKeySetID,
		AllowedAudiences: parseCSV(c.AllowedAudiences),
		ClaimPolicy: client.TrustProviderClaimPolicy{
			Allowed:  parseCSV(c.AllowedClaims),
			Required: parseCSV(c.RequiredClaims),
			Constant: constant,
		},
	}
	if in.Type != "oidc" {
		t.Fatalf("Type = %q, want oidc", in.Type)
	}
	if len(in.AllowedAudiences) != 2 {
		t.Fatalf("AllowedAudiences = %v, want 2 entries", in.AllowedAudiences)
	}
	if in.ClaimPolicy.Constant["tenant"] != "acme" {
		t.Fatalf("ClaimPolicy.Constant[tenant] = %v, want acme", in.ClaimPolicy.Constant["tenant"])
	}
}

func TestParseCSV_TrimsAndDropsEmpty(t *testing.T) {
	got := parseCSV(" a , ,b ")
	if strings.Join(got, ",") != "a,b" {
		t.Fatalf("parseCSV = %v, want [a b]", got)
	}
}

func TestParseCSV_Empty(t *testing.T) {
	if got := parseCSV(""); got != nil {
		t.Fatalf("parseCSV(\"\") = %v, want nil", got)
	}
}

func TestParseJSONMap_ValidObject(t *testing.T) {
	m, err := parseJSONMap(`{"key":"val","n":42}`)
	if err != nil {
		t.Fatalf("parseJSONMap: %v", err)
	}
	if m["key"] != "val" {
		t.Fatalf("m[key] = %v, want val", m["key"])
	}
}

func TestParseJSONMap_EmptyString(t *testing.T) {
	m, err := parseJSONMap("")
	if err != nil || m != nil {
		t.Fatalf("parseJSONMap(\"\") = %v, %v; want nil, nil", m, err)
	}
}

func TestParseJSONMap_InvalidJSON(t *testing.T) {
	if _, err := parseJSONMap("{not json}"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
