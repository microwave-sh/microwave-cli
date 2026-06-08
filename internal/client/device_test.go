package client_test

import (
	"context"
	"encoding/json"
	"io"
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
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code": "device_abc", "user_code": "ABCD-EFGH",
				"verification_uri": "https://app/authorize", "verification_uri_complete": "https://app/authorize/device_abc",
				"expires_in": 900, "interval": 5,
			})
		case "/auth/device/token":
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			}
			raw, _ := io.ReadAll(r.Body)
			if string(raw) != "device_code=device_abc&grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code" {
				t.Errorf("form body = %q", raw)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "jwt-xyz", "token_type": "Bearer", "expires_in": 86400})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := client.New(srv.URL, "", "test", false)
	dc, err := c.RequestDeviceCode(context.Background())
	if err != nil || dc.DeviceCode != "device_abc" || dc.UserCode != "ABCD-EFGH" || dc.VerificationURIComplete == "" || dc.Interval != 5 {
		t.Fatalf("request = %#v, err=%v", dc, err)
	}
	poll, err := c.PollDeviceToken(context.Background(), "device_abc")
	if err != nil || poll.AccessToken != "jwt-xyz" || poll.TokenType != "Bearer" {
		t.Fatalf("poll = %#v, err=%v", poll, err)
	}
}
