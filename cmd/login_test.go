package cmd_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-cli/cmd"
)

// TestLoginDeviceFlowStoresToken drives `microwave login --no-browser` through
// the shared SDK auth core against a mock authorization server: RFC 8414
// discovery, an RFC 8628 device-authorization request, then a `/token`
// device-code poll that returns authorization_pending once before succeeding.
func TestLoginDeviceFlowStoresToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	var base string
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/oauth-authorization-server":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issuer":                        base,
				"token_endpoint":                base + "/token",
				"device_authorization_endpoint": base + "/device_authorization",
				"grant_types_supported":         []string{"urn:ietf:params:oauth:grant-type:device_code"},
			})
		case "/device_authorization":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":               "device_abc",
				"user_code":                 "ABCD-EFGH",
				"verification_uri":          "https://example.invalid/activate",
				"verification_uri_complete": "https://example.invalid/activate?user_code=ABCD-EFGH",
				"expires_in":                900,
				"interval":                  1,
			})
		case "/token":
			polls++
			if polls < 2 {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "authorization_pending"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "jwt-xyz", "token_type": "Bearer", "expires_in": 86400})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	base = srv.URL

	c := &cmd.LoginCmd{NoBrowser: true}
	g := &cmd.Globals{AuthURL: srv.URL, Version: "test", Output: "table"}
	if err := c.Run(context.Background(), g); err != nil {
		t.Fatalf("login: %v", err)
	}
	if polls < 2 {
		t.Fatalf("expected to poll through authorization_pending, polls=%d", polls)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "microwave", "config.toml"))
	if !strings.Contains(string(data), "jwt-xyz") {
		t.Fatalf("config did not contain jwt-xyz:\n%s", data)
	}
}
