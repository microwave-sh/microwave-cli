package cmd_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microwave-sh/microwave-cli/cmd"
)

func TestTokensCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/management-keys" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "key_1", "key": "mw_live_secret", "name": "ci"})
	}))
	defer srv.Close()

	g := &cmd.Globals{APIURL: srv.URL, Token: "jwt-xyz", Version: "test", Output: "json"}
	c := &cmd.TokensCmd{Create: cmd.TokensCreateCmd{Name: "ci", Scopes: []string{"keys:read"}}}
	if err := c.Create.Run(context.Background(), g); err != nil {
		t.Fatalf("tokens create: %v", err)
	}
}
