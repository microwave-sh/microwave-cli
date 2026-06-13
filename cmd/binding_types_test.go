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

// newTestGlobals returns a Globals wired to the given test server URL.
func newTestGlobals(serverURL string) *Globals {
	return &Globals{
		Token:  "test-token",
		APIURL: serverURL,
	}
}

func TestCmdBindingTypes_Create(t *testing.T) {
	want := client.TrustBindingTypeDef{
		ID:             "tbt_abc123",
		WorkspaceID:    "ws_1",
		Key:            "acme_tfc",
		Label:          "Acme TFC",
		IdentityFields: []string{"a", "b"},
		CreatedAt:      time.Now().UTC().Truncate(time.Second),
		UpdatedAt:      time.Now().UTC().Truncate(time.Second),
	}

	var gotBody client.TrustBindingTypeInput
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/trust-binding-types" {
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

	cmd := &btCreateCmd{
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

func TestCmdBindingTypes_List_IncludesSystemAndOwn(t *testing.T) {
	rows := []client.TrustBindingTypeDef{
		{ID: "tbt_sys1", WorkspaceID: "SYSTEM", Key: "terraform_cloud", Label: "Terraform Cloud"},
		{ID: "tbt_sys2", WorkspaceID: "SYSTEM", Key: "github_actions", Label: "GitHub Actions"},
		{ID: "tbt_sys3", WorkspaceID: "SYSTEM", Key: "google_workload_identity", Label: "Google Workload Identity"},
		{ID: "tbt_own1", WorkspaceID: "ws_1", Key: "acme_custom", Label: "Acme Custom"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/trust-binding-types" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rows)
	}))
	defer srv.Close()

	cmd := &btListCmd{}
	g := newTestGlobals(srv.URL)

	// Call the underlying client directly to test listing returns all 4.
	defs, err := g.Client().ListTrustBindingTypeDefs(context.Background())
	if err != nil {
		t.Fatalf("ListTrustBindingTypeDefs: %v", err)
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
		t.Fatalf("btListCmd.Run: %v", err)
	}
}

func TestCmdBindingTypes_Delete(t *testing.T) {
	const targetID = "tbt_xxx"
	var gotMethod, gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cmd := &btDeleteCmd{ID: targetID}
	g := newTestGlobals(srv.URL)

	if err := cmd.Run(context.Background(), g); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
	wantPath := "/api/trust-binding-types/" + targetID
	if gotPath != wantPath {
		t.Errorf("path = %s, want %s", gotPath, wantPath)
	}
}

// TestBtCreateCmd_ToInput exercises flag→input mapping without an HTTP server.
func TestBtCreateCmd_ToInput(t *testing.T) {
	cmd := btCreateCmd{
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
