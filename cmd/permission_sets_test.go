package cmd

import (
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

func TestParsePermissions_valid(t *testing.T) {
	specs := []string{"deploy:Deploy:true"}
	got, err := parsePermissions(specs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(got))
	}
	want := client.PermissionInput{Name: "deploy", Label: "Deploy", Dangerous: true}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}

func TestParsePermissions_missingLabel(t *testing.T) {
	specs := []string{"deploy"}
	_, err := parsePermissions(specs)
	if err == nil {
		t.Fatal("expected error for missing label, got nil")
	}
}

func TestParsePermissions_emptyLabel(t *testing.T) {
	specs := []string{"deploy:"}
	_, err := parsePermissions(specs)
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}
}

func TestParsePermissions_noDangerous(t *testing.T) {
	specs := []string{"read:Read"}
	got, err := parsePermissions(specs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Dangerous {
		t.Error("expected Dangerous=false when not specified")
	}
}

func TestParsePermissions_nil(t *testing.T) {
	got, err := parsePermissions(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}
