package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDo_SetsAuthAndVersionHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing bearer: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("API-Version") != apiVersion {
			t.Errorf("API-Version = %q", r.Header.Get("API-Version"))
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok", "test", false)
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.Do(context.Background(), "GET", "/api/keys/search", nil, &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatal("expected ok")
	}
}

func TestDo_MapsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"type":"invalid_input","message":"bad","errors":[{"field":"name","message":"required"}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok", "test", false)
	err := c.Do(context.Background(), "POST", "/api/key-specs", map[string]string{}, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type = %T, want *APIError", err)
	}
	if apiErr.Status != 422 || apiErr.Type != "invalid_input" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}
