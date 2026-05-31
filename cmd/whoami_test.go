package cmd_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microwave-sh/microwave-cli/cmd"
)

func TestWhoamiUsesMeEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/me" {
			t.Errorf("expected /api/me, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer jwt-xyz" {
			t.Errorf("missing bearer token")
		}
		json.NewEncoder(w).Encode(map[string]any{
			"workspace_id": "org_42", "actor": "user_seth", "tier": "pro",
			"permissions": []string{"keys:read", "keys:write"},
		})
	}))
	defer srv.Close()

	g := &cmd.Globals{APIURL: srv.URL, Token: "jwt-xyz", Version: "test", Output: "json"}
	if err := (&cmd.WhoamiCmd{}).Run(context.Background(), g); err != nil {
		t.Fatalf("whoami: %v", err)
	}
}
