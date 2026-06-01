package client

import (
	"context"
	"net/url"
)

// ── SDK generation types ─────────────────────────────────────────────────────

// Connection represents a spec-sync connection (a GitHub repo wired to Microwave).
type Connection struct {
	ID           string `json:"id"`
	Repo         string `json:"repo"`
	Branch       string `json:"branch"`
	Status       string `json:"status"`
	StatusDetail string `json:"status_detail,omitempty"`
}

// ConnectionInput is the request body for creating a connection.
type ConnectionInput struct {
	GHInstallationID int64  `json:"gh_installation_id"`
	Repo             string `json:"repo"`
	Branch           string `json:"branch"`
	SpecPath         string `json:"spec_path,omitempty"`
	ConfigPath       string `json:"config_path,omitempty"`
}

// SDKTarget is a single language SDK target attached to a connection.
type SDKTarget struct {
	ID           string `json:"id"`
	Language     string `json:"language"`
	Repo         string `json:"repo"`
	Instructions string `json:"instructions,omitempty"`
}

// SDKRun is a single generation run record for an SDK target.
type SDKRun struct {
	ID      string `json:"id"`
	RunType string `json:"run_type"`
	Status  string `json:"status"`
	PRUrl   string `json:"pr_url,omitempty"`
	Breaking bool  `json:"breaking,omitempty"`
	Error   string `json:"error,omitempty"`
}

// listEnvelope is used to unmarshal { "data": [...] } list responses.
type listEnvelope[T any] struct {
	Data []T `json:"data"`
}

// ── Connection methods ───────────────────────────────────────────────────────

// ListConnections returns all connections for the authenticated workspace.
func (c *Client) ListConnections(ctx context.Context) ([]Connection, error) {
	var env listEnvelope[Connection]
	if err := c.Do(ctx, "GET", "/api/connections", nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// CreateConnection creates a new spec-sync connection.
func (c *Client) CreateConnection(ctx context.Context, in ConnectionInput) (*Connection, error) {
	var out Connection
	return &out, c.Do(ctx, "POST", "/api/connections", in, &out)
}

// ── SDK target methods ───────────────────────────────────────────────────────

// ListSDKTargets returns all SDK targets for a connection.
func (c *Client) ListSDKTargets(ctx context.Context, connID string) ([]SDKTarget, error) {
	path := "/api/connections/" + url.PathEscape(connID) + "/sdk/targets"
	var env listEnvelope[SDKTarget]
	if err := c.Do(ctx, "GET", path, nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// ── SDK run methods ──────────────────────────────────────────────────────────

// ListSDKRuns returns the run history for an SDK target.
func (c *Client) ListSDKRuns(ctx context.Context, targetID string) ([]SDKRun, error) {
	path := "/api/sdk/targets/" + url.PathEscape(targetID) + "/runs"
	var env listEnvelope[SDKRun]
	if err := c.Do(ctx, "GET", path, nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// TriggerRun enqueues a new generation run for an SDK target.
func (c *Client) TriggerRun(ctx context.Context, targetID string) (*SDKRun, error) {
	path := "/api/sdk/targets/" + url.PathEscape(targetID) + "/runs"
	var out SDKRun
	return &out, c.Do(ctx, "POST", path, map[string]any{}, &out)
}
