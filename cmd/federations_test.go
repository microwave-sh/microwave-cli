package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

func TestCmdFederations_Create(t *testing.T) {
	want := client.TrustFederation{
		ID:             "fed_abc123",
		WorkspaceID:    "ws_1",
		Key:            "acme_tfc",
		Label:          "Acme TFC",
		IdentityFields: []string{"a", "b"},
		CreatedAt:      time.Now().UTC().Truncate(time.Second),
		UpdatedAt:      time.Now().UTC().Truncate(time.Second),
	}

	var gotBody client.TrustFederationInput
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/trust-federations" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	cmd := &fedCreateCmd{
		Key:            "acme_tfc",
		Label:          "Acme TFC",
		IdentityFields: "a,b",
		OutputKeySpec:  "spec_x",
	}

	g := newTestGlobals(srv.URL)
	if err := cmd.Run(context.Background(), g); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if gotBody.Key != "acme_tfc" {
		t.Errorf("body.Key = %q, want acme_tfc", gotBody.Key)
	}
	if gotBody.Label != "Acme TFC" {
		t.Errorf("body.Label = %q, want Acme TFC", gotBody.Label)
	}
	if len(gotBody.IdentityFields) != 2 || gotBody.IdentityFields[0] != "a" || gotBody.IdentityFields[1] != "b" {
		t.Errorf("body.IdentityFields = %v, want [a b]", gotBody.IdentityFields)
	}
	if gotBody.OutputKeySpecID != "spec_x" {
		t.Errorf("body.OutputKeySpecID = %q, want spec_x", gotBody.OutputKeySpecID)
	}
}

func TestCmdFederations_List_IncludesSystemAndOwn(t *testing.T) {
	rows := []client.TrustFederation{
		{ID: "fed_sys1", WorkspaceID: "SYSTEM", Key: "terraform_cloud", Label: "Terraform Cloud"},
		{ID: "fed_sys2", WorkspaceID: "SYSTEM", Key: "github_actions", Label: "GitHub Actions"},
		{ID: "fed_sys3", WorkspaceID: "SYSTEM", Key: "google_workload_identity", Label: "Google Workload Identity"},
		{ID: "fed_own1", WorkspaceID: "ws_1", Key: "acme_custom", Label: "Acme Custom"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/trust-federations" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rows)
	}))
	defer srv.Close()

	cmd := &fedListCmd{}
	g := newTestGlobals(srv.URL)

	// Call the underlying client directly to test listing returns all 4.
	defs, err := g.Client().ListTrustFederations(context.Background())
	if err != nil {
		t.Fatalf("ListTrustFederations: %v", err)
	}
	if len(defs) != 4 {
		t.Fatalf("len(defs) = %d, want 4", len(defs))
	}
	keys := make([]string, len(defs))
	for i, d := range defs {
		keys[i] = d.Key
	}
	if !strings.Contains(strings.Join(keys, ","), "terraform_cloud") {
		t.Errorf("expected terraform_cloud in list, got %v", keys)
	}
	if !strings.Contains(strings.Join(keys, ","), "acme_custom") {
		t.Errorf("expected acme_custom in list, got %v", keys)
	}

	// Also verify Run does not error.
	if err := cmd.Run(context.Background(), g); err != nil {
		t.Fatalf("fedListCmd.Run: %v", err)
	}
}

func TestCmdFederations_Delete(t *testing.T) {
	const targetID = "fed_xxx"
	var gotMethod, gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cmd := &fedDeleteCmd{ID: targetID}
	g := newTestGlobals(srv.URL)

	if err := cmd.Run(context.Background(), g); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
	wantPath := "/api/trust-federations/" + targetID
	if gotPath != wantPath {
		t.Errorf("path = %s, want %s", gotPath, wantPath)
	}
}

// TestFedCreateCmd_ToInput exercises flag→input mapping without an HTTP server.
func TestFedCreateCmd_ToInput(t *testing.T) {
	cmd := fedCreateCmd{
		Key:            "acme_tfc",
		Label:          "Acme TFC",
		IdentityFields: "org,workspace",
		OutputKeySpec:  "spec_abc",
		Description:    "test desc",
		Policy:         `has(identity.org)`,
	}
	in := cmd.toInput()

	if in.Key != "acme_tfc" {
		t.Errorf("Key = %q, want acme_tfc", in.Key)
	}
	if len(in.IdentityFields) != 2 {
		t.Fatalf("IdentityFields len = %d, want 2", len(in.IdentityFields))
	}
	if in.IdentityFields[0] != "org" || in.IdentityFields[1] != "workspace" {
		t.Errorf("IdentityFields = %v", in.IdentityFields)
	}
	if in.OutputKeySpecID != "spec_abc" {
		t.Errorf("OutputKeySpecID = %q, want spec_abc", in.OutputKeySpecID)
	}
	if in.Policy != `has(identity.org)` {
		t.Errorf("Policy = %q", in.Policy)
	}
}
