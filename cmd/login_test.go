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

func TestLoginDeviceFlowStoresToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code": "device_abc", "authorize_url": "https://example.invalid/authorize/device_abc",
				"expires_in": 900, "interval": 0,
			})
		case "/auth/device/token":
			polls++
			if polls < 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "pending"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "approved", "token": "jwt-xyz", "expires_in": 86400})
		}
	}))
	defer srv.Close()

	c := &cmd.LoginCmd{NoBrowser: true}
	g := &cmd.Globals{APIURL: srv.URL, Version: "test", Output: "table"}
	if err := c.Run(context.Background(), g); err != nil {
		t.Fatalf("login: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "microwave", "config.toml"))
	if !strings.Contains(string(data), "jwt-xyz") {
		t.Fatalf("config did not contain jwt-xyz:\n%s", data)
	}
}
