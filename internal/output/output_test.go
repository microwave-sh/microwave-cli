package output

import (
	"testing"
	"time"
)

func TestFormatTimeAgo_Zero(t *testing.T) {
	if got := FormatTimeAgo(time.Time{}); got != "—" {
		t.Fatalf("FormatTimeAgo(zero) = %q, want —", got)
	}
}

func TestFormatTimeAgo_Hours(t *testing.T) {
	got := FormatTimeAgo(time.Now().Add(-3 * time.Hour))
	if got != "3h ago" {
		t.Fatalf("FormatTimeAgo(-3h) = %q, want 3h ago", got)
	}
}

func TestColorStatus_Unknown_PassThrough(t *testing.T) {
	if got := ColorStatus("weird"); got != "weird" {
		t.Fatalf("ColorStatus(weird) = %q, want weird", got)
	}
}
