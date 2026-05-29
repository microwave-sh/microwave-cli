package cmd

import (
	"testing"
	"time"
)

// ── buildKeysFilter ───────────────────────────────────────────────────────

func TestBuildKeysFilter_AllEmpty(t *testing.T) {
	f := buildKeysFilter("", "", "")
	if f != nil {
		t.Fatalf("expected nil filter when all flags empty, got %v", f)
	}
}

func TestBuildKeysFilter_StatusOnly(t *testing.T) {
	f := buildKeysFilter("", "", "active")
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	v, ok := f["status"]
	if !ok {
		t.Fatal("expected status key in filter")
	}
	if v["eq"] != "active" {
		t.Fatalf("expected eq=active, got %v", v["eq"])
	}
	if _, hasSpec := f["spec_id"]; hasSpec {
		t.Fatal("spec_id should not be present when flag is empty")
	}
}

func TestBuildKeysFilter_AllSet(t *testing.T) {
	f := buildKeysFilter("spec-1", "user@example.com", "revoked")
	if len(f) != 3 {
		t.Fatalf("expected 3 filter entries, got %d", len(f))
	}
	for field, want := range map[string]string{
		"spec_id": "spec-1",
		"subject": "user@example.com",
		"status":  "revoked",
	} {
		entry, ok := f[field]
		if !ok {
			t.Fatalf("missing filter field %q", field)
		}
		if got := entry["eq"]; got != want {
			t.Fatalf("filter[%q][eq] = %v, want %q", field, got, want)
		}
	}
}

// ── parseExpiresAt ────────────────────────────────────────────────────────

func TestParseExpiresAt_Empty(t *testing.T) {
	tp, err := parseExpiresAt("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp != nil {
		t.Fatalf("expected nil for empty string, got %v", tp)
	}
}

func TestParseExpiresAt_ValidRFC3339(t *testing.T) {
	tp, err := parseExpiresAt("2026-01-15T12:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp == nil {
		t.Fatal("expected non-nil *time.Time")
	}
	want := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	if !tp.Equal(want) {
		t.Fatalf("got %v, want %v", *tp, want)
	}
}

func TestParseExpiresAt_InvalidFormat(t *testing.T) {
	_, err := parseExpiresAt("2026-01-15")
	if err == nil {
		t.Fatal("expected error for non-RFC3339 input, got nil")
	}
}
