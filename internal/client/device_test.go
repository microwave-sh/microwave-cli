package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

func TestRequestAndPollDeviceCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device":
			if r.Header.Get("Authorization") != "" {
				t.Errorf("device request must be tokenless")
			}
			json.NewEncoder(w).Encode(map[string]any{
				"device_code": "device_abc", "authorize_url": "https://app/authorize/device_abc",
				"expires_in": 900, "interval": 5,
			})
		case "/auth/device/token":
			json.NewEncoder(w).Encode(map[string]any{"status": "approved", "token": "jwt-xyz", "expires_in": 86400})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := client.New(srv.URL, "", "test", false)
	dc, err := c.RequestDeviceCode(context.Background())
	if err != nil || dc.DeviceCode != "device_abc" || dc.Interval != 5 {
		t.Fatalf("request = %#v, err=%v", dc, err)
	}
	poll, err := c.PollDeviceToken(context.Background(), "device_abc")
	if err != nil || poll.Status != "approved" || poll.Token != "jwt-xyz" {
		t.Fatalf("poll = %#v, err=%v", poll, err)
	}
}
