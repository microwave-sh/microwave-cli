package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

func TestTBCreate_ToInput(t *testing.T) {
	c := tbCreateCmd{
		BindingType:  "custom_ci",
		Identity:     `{"repository":"octocat/hello-world","workflow":"deploy.yml"}`,
		OutputClaims: `{"tier":"prod"}`,
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	if in.BindingType != "custom_ci" {
		t.Fatalf("BindingType = %q, want custom_ci", in.BindingType)
	}
	if in.Identity["repository"] != "octocat/hello-world" {
		t.Fatalf("Identity[repository] = %v", in.Identity["repository"])
	}
	if in.OutputClaims["tier"] != "prod" {
		t.Fatalf("OutputClaims[tier] = %v", in.OutputClaims["tier"])
	}
}

func TestTBCreate_ToInput_InvalidIdentityJSON(t *testing.T) {
	c := tbCreateCmd{
		BindingType: "custom_ci",
		Identity:    `{bad json`,
	}
	if _, err := c.toInput(); err == nil {
		t.Fatal("expected error for invalid identity JSON")
	}
}

// ── enable (positional, catalog-backed) ─────────────────────────────────────

func TestCmdConnectorsEnable_PositionalKey(t *testing.T) {
	// Stub: GET /api/trust-binding-types → returns a SYSTEM terraform_cloud type.
	// POST /api/trust-bindings → returns a created binding.

	defList := []client.TrustBindingTypeDef{
		{
			ID:             "tbt_tfc",
			WorkspaceID:    "SYSTEM",
			Key:            "terraform_cloud",
			Label:          "Terraform Cloud",
			IdentityFields: []string{"terraform_organization_name", "terraform_workspace_name"},
		},
	}

	var gotBindingBody client.TrustBindingInput
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/trust-binding-types":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(defList)

		case r.Method == http.MethodPost && r.URL.Path == "/api/trust-bindings":
			if err := json.NewDecoder(r.Body).Decode(&gotBindingBody); err != nil {
				t.Fatalf("decode binding body: %v", err)
			}
			result := client.TrustBinding{
				ID:          "tb_new",
				WorkspaceID: "ws_1",
				BindingType: gotBindingBody.BindingType,
				Identity:    gotBindingBody.Identity,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(result)

		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cmd := &tbEnableCmd{
		Key:      "terraform_cloud",
		Identity: "terraform_organization_name=acme,terraform_workspace_name=prod",
	}
	g := newTestGlobals(srv.URL)

	if err := cmd.Run(context.Background(), g); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if gotBindingBody.BindingType != "terraform_cloud" {
		t.Errorf("BindingType = %q, want terraform_cloud", gotBindingBody.BindingType)
	}
	if gotBindingBody.Identity["terraform_organization_name"] != "acme" {
		t.Errorf("Identity[terraform_organization_name] = %v, want acme", gotBindingBody.Identity["terraform_organization_name"])
	}
	if gotBindingBody.Identity["terraform_workspace_name"] != "prod" {
		t.Errorf("Identity[terraform_workspace_name] = %v, want prod", gotBindingBody.Identity["terraform_workspace_name"])
	}
}

func TestCmdConnectorsEnable_MissingIdentityField(t *testing.T) {
	// Stub returns a binding type requiring fields A and B.
	// CLI only provides A — expect client-side error naming "B".
	defList := []client.TrustBindingTypeDef{
		{
			ID:             "tbt_x",
			WorkspaceID:    "SYSTEM",
			Key:            "needs_two_fields",
			Label:          "Needs Two Fields",
			IdentityFields: []string{"field_a", "field_b"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/trust-binding-types" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(defList)
			return
		}
		// POST should NOT be reached.
		t.Errorf("unexpected request %s %s — POST should not have been sent", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cmd := &tbEnableCmd{
		Key:      "needs_two_fields",
		Identity: "field_a=x", // missing field_b
	}
	g := newTestGlobals(srv.URL)

	err := cmd.Run(context.Background(), g)
	if err == nil {
		t.Fatal("expected error for missing identity field, got nil")
	}
	if !strings.Contains(err.Error(), "field_b") {
		t.Errorf("error should name the missing field 'field_b', got: %s", err.Error())
	}
}

func TestCmdConnectorsEnable_UnknownKey(t *testing.T) {
	// Stub returns empty catalog.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/trust-binding-types" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]client.TrustBindingTypeDef{})
			return
		}
		t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cmd := &tbEnableCmd{
		Key:      "ghost",
		Identity: "x=y",
	}
	g := newTestGlobals(srv.URL)

	err := cmd.Run(context.Background(), g)
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error should name the unknown key 'ghost', got: %s", err.Error())
	}
}
