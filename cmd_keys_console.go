// Copyright 2026 Mataki Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// KeysCmd exposes AKaaS key lifecycle operations.
type KeysCmd struct {
	List   KeysListCmd   `cmd:"" help:"List AKaaS keys."`
	Issue  KeysIssueCmd  `cmd:"" help:"Issue an AKaaS key."`
	Verify KeysVerifyCmd `cmd:"" help:"Verify an opaque AKaaS key."`
	Revoke KeysRevokeCmd `cmd:"" help:"Revoke an AKaaS key."`
}

type KeysListCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
}

func (c *KeysListCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, workspacePath(workspace, "/keys"), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type KeysIssueCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	Spec      string `help:"Key Spec name." default:"default"`
	Format    string `enum:"opaque,jwt_asymmetric,jwt_hs256" default:"opaque" help:"Key format."`
	Scopes    string `help:"Comma-separated scopes to grant."`
	Claims    string `help:"JSON object of claims to embed or bind to the key."`
	ExpiresIn string `help:"Relative expiry such as 30d or 12h."`
}

func (c *KeysIssueCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	claims := map[string]any{}
	if c.Claims != "" {
		if err := json.Unmarshal([]byte(c.Claims), &claims); err != nil {
			return fmt.Errorf("claims must be a JSON object: %w", err)
		}
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"key_spec":   c.Spec,
		"format":     c.Format,
		"scopes":     splitCSV(c.Scopes),
		"claims":     claims,
		"expires_in": c.ExpiresIn,
	}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/keys/issue"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type KeysVerifyCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	Key       string `arg:"" help:"Opaque key to verify."`
	Scopes    string `help:"Comma-separated scopes required for verification."`
}

func (c *KeysVerifyCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{"key": c.Key, "scopes": splitCSV(c.Scopes)}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/keys/verify"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type KeysRevokeCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	KeyID     string `arg:"" help:"Key ID to revoke."`
	Reason    string `help:"Revocation reason."`
}

func (c *KeysRevokeCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{"reason": c.Reason}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/keys/"+url.PathEscape(c.KeyID)+"/revoke"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

// JWKSCmd manages JWKS sets for asymmetric JWT verification.
type JWKSCmd struct {
	Get    JWKSGetCmd    `cmd:"" help:"Fetch a public JWKS set."`
	URL    JWKSURLCmd    `cmd:"" help:"Print the public JWKS URL."`
	Rotate JWKSRotateCmd `cmd:"" help:"Rotate a JWKS signing key."`
	List   JWKSListCmd   `cmd:"" help:"List JWKS sets."`
}

type JWKSGetCmd struct {
	Workspace string `help:"Workspace slug. Defaults to active workspace."`
	Set       string `arg:"" optional:"" default:"default" help:"JWKS set name."`
}

func (c *JWKSGetCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(false)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, publicJWKSURL(workspace, c.Set), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, true)
}

type JWKSURLCmd struct {
	Workspace string `help:"Workspace slug. Defaults to active workspace."`
	Set       string `arg:"" optional:"" default:"default" help:"JWKS set name."`
}

func (c *JWKSURLCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	jwksURL := publicJWKSURL(workspace, c.Set)
	if g.isJSON() {
		return printJSON(map[string]string{"url": jwksURL})
	}
	fmt.Println(jwksURL)
	return nil
}

type JWKSRotateCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	Set       string `arg:"" optional:"" default:"default" help:"JWKS set name."`
}

func (c *JWKSRotateCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/jwks/"+url.PathEscape(c.Set)+"/rotate"), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type JWKSListCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
}

func (c *JWKSListCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, workspacePath(workspace, "/jwks"), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

// ConsoleCmd exposes terminal-friendly Developer Console helpers.
type ConsoleCmd struct {
	URL      ConsoleURLCmd      `cmd:"" help:"Print the Developer Console URL."`
	Activity ConsoleActivityCmd `cmd:"" help:"Fetch recent workspace activity."`
	Usage    ConsoleUsageCmd    `cmd:"" help:"Fetch workspace usage metrics."`
}

type ConsoleURLCmd struct {
	Workspace string `help:"Workspace slug. Defaults to active workspace."`
	Path      string `help:"Console path." default:"/overview"`
}

func (c *ConsoleURLCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	path := "/" + strings.TrimLeft(c.Path, "/")
	consoleURL := defaultConsoleURL + "/w/" + url.PathEscape(workspace) + path
	if g.isJSON() {
		return printJSON(map[string]string{"url": consoleURL})
	}
	fmt.Println(consoleURL)
	return nil
}

type ConsoleActivityCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	Limit     int    `help:"Number of events to fetch." default:"25"`
}

func (c *ConsoleActivityCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	path := workspacePath(workspace, "/activity?limit="+url.QueryEscape(fmt.Sprint(c.Limit)))
	resp, err := client.do(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type ConsoleUsageCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	Last      string `help:"Lookback window." default:"7d"`
}

func (c *ConsoleUsageCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	path := workspacePath(workspace, "/usage?last="+url.QueryEscape(c.Last))
	resp, err := client.do(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

func publicJWKSURL(workspace string, set string) string {
	return "https://" + workspace + ".microwave.sh/.well-known/jwks/" + url.PathEscape(set) + ".json"
}
