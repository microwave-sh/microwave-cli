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

// TestLoginDeviceApprovalStoresToken drives `microwave login --no-browser`
// end-to-end against a mock that advertises cli_login_flow=device_approval: the
// CLI requests a device code, polls through one pending response, and stores the
// approved token.
func TestLoginDeviceApprovalStoresToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/oauth-authorization-server":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issuer":         "x",
				"token_endpoint": "http://x/token",
				"cli_login_flow": "device_approval",
			})
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      "device_abc",
				"user_code":        "ABCD-EFGH",
				"verification_uri": "https://example.invalid/device",
				"expires_in":       900,
				"interval":         0,
			})
		case "/auth/device/token":
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

	c := &cmd.LoginCmd{NoBrowser: true}
	g := &cmd.Globals{AuthURL: srv.URL, APIURL: srv.URL, Version: "test", Output: "table"}
	if err := c.Run(context.Background(), g); err != nil {
		t.Fatalf("login: %v", err)
	}
	if polls < 2 {
		t.Fatalf("expected to poll through pending, polls=%d", polls)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "microwave", "config.toml"))
	if !strings.Contains(string(data), "jwt-xyz") {
		t.Fatalf("config did not contain jwt-xyz:\n%s", data)
	}
}
