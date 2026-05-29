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
	"fmt"
	"net/http"
	"os"
)

// WorkspaceCmd manages Microwave workspaces.
type WorkspaceCmd struct {
	List   WorkspaceListCmd   `cmd:"" help:"List workspaces visible to the current credential."`
	Get    WorkspaceGetCmd    `cmd:"" help:"Fetch a workspace."`
	Create WorkspaceCreateCmd `cmd:"" help:"Create a workspace."`
	Use    WorkspaceUseCmd    `cmd:"" help:"Set the active workspace for this machine."`
}

type WorkspaceListCmd struct{}

func (c *WorkspaceListCmd) Run(g *Globals) error {
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, "/workspaces", nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type WorkspaceGetCmd struct {
	Workspace string `arg:"" optional:"" help:"Workspace slug or ID. Defaults to the active workspace."`
}

func (c *WorkspaceGetCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, workspacePath(workspace, ""), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type WorkspaceCreateCmd struct {
	Name     string `help:"Display name for the workspace."`
	Slug     string `help:"URL-safe workspace slug."`
	Internal bool   `help:"Mark the workspace as Mataki-internal."`
}

func (c *WorkspaceCreateCmd) Run(g *Globals) error {
	if c.Name == "" && c.Slug == "" {
		return fmt.Errorf("at least one of --name or --slug is required")
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"name":        c.Name,
		"slug":        c.Slug,
		"is_internal": c.Internal,
	}
	resp, err := client.do(context.Background(), http.MethodPost, "/workspaces", payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type WorkspaceUseCmd struct {
	Workspace string `arg:"" help:"Workspace slug or ID."`
}

func (c *WorkspaceUseCmd) Run(g *Globals) error {
	cfg := LoadGlobalConfig()
	cfg.Workspace.Active = c.Workspace
	if err := SaveGlobalConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Active workspace set to %s\n", c.Workspace)
	return nil
}

// SpecCmd manages OpenAPI spec connection and validation.
type SpecCmd struct {
	Validate SpecValidateCmd `cmd:"" help:"Validate the local OpenAPI spec path in microwave.yaml."`
	Connect  SpecConnectCmd  `cmd:"" help:"Connect this spec repository to Microwave."`
	Diff     SpecDiffCmd     `cmd:"" help:"Ask Microwave to semantically diff two OpenAPI specs."`
	List     SpecListCmd     `cmd:"" help:"List connected specs for a workspace."`
}

type SpecValidateCmd struct {
	Remote bool `help:"Also submit the spec to Microwave's validator."`
}

func (c *SpecValidateCmd) Run(g *Globals) error {
	cfg, err := LoadProjectConfig(".")
	if err != nil {
		return err
	}
	specPath := cfg.ResolveSpecPath()
	data, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read spec.path %q: %w", cfg.SpecPath, err)
	}
	if !c.Remote {
		if g.isJSON() {
			return printJSON(map[string]any{"valid": true, "spec_path": specPath, "bytes": len(data)})
		}
		fmt.Printf("Spec exists: %s (%d bytes)\n", specPath, len(data))
		return nil
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"workspace": cfg.Workspace,
		"spec_path": cfg.SpecPath,
		"contents":  string(data),
	}
	resp, err := client.do(context.Background(), http.MethodPost, "/specs/validate", payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type SpecConnectCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace or microwave.yaml."`
	Repo      string `help:"GitHub spec repository, for example acme/acme-specs."`
}

func (c *SpecConnectCmd) Run(g *Globals) error {
	cfg, err := LoadProjectConfig(".")
	if err != nil {
		return err
	}
	workspace, err := workspaceOrActive(firstNonEmpty(c.Workspace, cfg.Workspace), g)
	if err != nil {
		return err
	}
	specPath := cfg.ResolveSpecPath()
	spec, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read spec.path %q: %w", cfg.SpecPath, err)
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"repository":    c.Repo,
		"config_path":   projectConfigFile,
		"config":        cfg.RawContents,
		"spec_path":     cfg.SpecPath,
		"spec":          string(spec),
		"sdk_languages": cfg.SDKLanguages(),
		"docs_domain":   cfg.Docs.Domain,
		"api_version":   defaultAPIVersion,
	}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/specs/connect"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type SpecDiffCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	From      string `arg:"" help:"Old OpenAPI spec path."`
	To        string `arg:"" help:"New OpenAPI spec path."`
	Policy    string `enum:"block,warn,allow" default:"warn" help:"Breaking change policy."`
}

func (c *SpecDiffCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	from, err := os.ReadFile(c.From)
	if err != nil {
		return fmt.Errorf("read %s: %w", c.From, err)
	}
	to, err := os.ReadFile(c.To)
	if err != nil {
		return fmt.Errorf("read %s: %w", c.To, err)
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"from_path": c.From,
		"from":      string(from),
		"to_path":   c.To,
		"to":        string(to),
		"policy":    c.Policy,
	}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/specs/diff"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type SpecListCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
}

func (c *SpecListCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, workspacePath(workspace, "/specs"), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
