package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

// listFlags are embedded by every resource `list` subcommand.
type listFlags struct {
	Limit  int    `help:"Max results." default:"25"`
	Cursor string `help:"Pagination cursor."`
	Sort   string `help:"Sort field (e.g. created_at)."`
	Desc   bool   `help:"Sort descending." default:"true"`
}

func (f listFlags) searchRequest(filter map[string]map[string]any) client.SearchRequest {
	req := client.SearchRequest{Limit: f.Limit, Cursor: f.Cursor, Filter: filter}
	if f.Sort != "" {
		dir := "asc"
		if f.Desc {
			dir = "desc"
		}
		req.Sort = []client.SortDirective{{Field: f.Sort, Direction: dir}}
	}
	return req
}

// parseCSV splits a comma-separated string, trims whitespace, and drops empty segments.
func parseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// parseJSONMap unmarshals a JSON object string into map[string]any.
// Returns nil, nil for an empty or whitespace-only string.
func parseJSONMap(s string) (map[string]any, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return m, nil
}

// parseKVMap parses a comma-separated list of key=value pairs into map[string]any.
// Example: "terraform_organization_name=acme,terraform_workspace_name=prod"
// Returns nil, nil for an empty or whitespace-only string.
func parseKVMap(s string) (map[string]any, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	m := make(map[string]any, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.IndexByte(p, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid key=value pair %q: missing '='", p)
		}
		k := strings.TrimSpace(p[:idx])
		v := strings.TrimSpace(p[idx+1:])
		if k == "" {
			return nil, fmt.Errorf("invalid key=value pair %q: empty key", p)
		}
		m[k] = v
	}
	return m, nil
}
