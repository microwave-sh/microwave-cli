package main

import (
	"testing"

	"github.com/alecthomas/kong"
)

func TestKongParsesProductCommands(t *testing.T) {
	tests := [][]string{
		{"version"},
		{"config", "path"},
		{"workspace", "use", "acme"},
		{"spec", "diff", "--workspace", "acme", "old.yaml", "new.yaml"},
		{"sdk", "generate", "--workspace", "acme", "--languages", "go,python,node", "--dry-run"},
		{"docs", "open", "--workspace", "acme"},
		{"collections", "open", "microwave", "stdlib"},
		{"keys", "issue", "--workspace", "acme", "--scopes", "read:users"},
		{"jwks", "url", "--workspace", "acme"},
		{"console", "url", "--workspace", "acme"},
	}

	for _, args := range tests {
		t.Run(args[0], func(t *testing.T) {
			var cli CLI
			parser, err := kong.New(&cli, kong.Name("microwave"))
			if err != nil {
				t.Fatalf("kong.New() error = %v", err)
			}
			if _, err := parser.Parse(args); err != nil {
				t.Fatalf("Parse(%v) error = %v", args, err)
			}
		})
	}
}
