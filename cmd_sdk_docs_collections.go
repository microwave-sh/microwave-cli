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
	"net/url"
)

// SDKCmd runs and inspects SDK generation jobs.
type SDKCmd struct {
	Generate SDKGenerateCmd `cmd:"" help:"Start an SDK generation job for a connected spec."`
	Runs     SDKRunsCmd     `cmd:"" help:"List SDK generation runs."`
	Get      SDKRunGetCmd   `cmd:"" help:"Fetch one SDK generation run."`
}

type SDKGenerateCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace or microwave.yaml."`
	Spec      string `help:"Connected spec ID or slug."`
	Languages string `help:"Comma-separated languages. Defaults to microwave.yaml SDK targets."`
	Mode      string `enum:"initial,delta" default:"delta" help:"Generation mode."`
	DryRun    bool   `help:"Validate and plan without opening SDK PRs."`
}

func (c *SDKGenerateCmd) Run(g *Globals) error {
	project, _ := LoadProjectConfig(".")
	workspace, err := workspaceOrActive(firstNonEmpty(c.Workspace, project.Workspace), g)
	if err != nil {
		return err
	}
	languages := splitCSV(c.Languages)
	if len(languages) == 0 {
		languages = project.SDKLanguages()
	}
	if len(languages) == 0 {
		return fmt.Errorf("no SDK languages supplied and none found in %s", projectConfigFile)
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"spec_id":   c.Spec,
		"languages": languages,
		"mode":      c.Mode,
		"dry_run":   c.DryRun,
	}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/sdk/generations"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type SDKRunsCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	Spec      string `help:"Filter by connected spec ID or slug."`
}

func (c *SDKRunsCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	path := workspacePath(workspace, "/sdk/runs")
	if c.Spec != "" {
		path += "?spec_id=" + url.QueryEscape(c.Spec)
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type SDKRunGetCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	RunID     string `arg:"" help:"SDK generation run ID."`
}

func (c *SDKRunGetCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, workspacePath(workspace, "/sdk/runs/"+url.PathEscape(c.RunID)), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

// DocsCmd manages generated documentation sites.
type DocsCmd struct {
	Deploy DocsDeployCmd `cmd:"" help:"Start a docs deployment."`
	Open   DocsOpenCmd   `cmd:"" help:"Print the docs URL for a workspace."`
	Export DocsExportCmd `cmd:"" help:"Request a static docs export."`
}

type DocsDeployCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace or microwave.yaml."`
	Spec      string `help:"Connected spec ID or slug."`
	Preview   bool   `help:"Create a preview deployment instead of production."`
}

func (c *DocsDeployCmd) Run(g *Globals) error {
	project, _ := LoadProjectConfig(".")
	workspace, err := workspaceOrActive(firstNonEmpty(c.Workspace, project.Workspace), g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{"spec_id": c.Spec, "preview": c.Preview}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/docs/deploys"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type DocsOpenCmd struct {
	Workspace string `help:"Workspace slug. Defaults to active workspace or microwave.yaml."`
}

func (c *DocsOpenCmd) Run(g *Globals) error {
	project, _ := LoadProjectConfig(".")
	workspace, err := workspaceOrActive(firstNonEmpty(c.Workspace, project.Workspace), g)
	if err != nil {
		return err
	}
	domain := project.Docs.Domain
	if domain == "" {
		domain = workspace + ".microwave.sh"
	}
	docsURL := "https://" + domain + "/docs"
	if g.isJSON() {
		return printJSON(map[string]string{"url": docsURL})
	}
	fmt.Println(docsURL)
	return nil
}

type DocsExportCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
	Version   string `help:"Docs version to export."`
}

func (c *DocsExportCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	path := workspacePath(workspace, "/docs/export")
	if c.Version != "" {
		path += "?version=" + url.QueryEscape(c.Version)
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

// CollectionsCmd manages shareable API collections.
type CollectionsCmd struct {
	List   CollectionsListCmd   `cmd:"" help:"List collections."`
	Import CollectionsImportCmd `cmd:"" help:"Import a collection from the local OpenAPI spec."`
	Fork   CollectionsForkCmd   `cmd:"" help:"Fork a public collection into a workspace."`
	Open   CollectionsOpenCmd   `cmd:"" help:"Print a collection URL."`
}

type CollectionsListCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace."`
}

func (c *CollectionsListCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	resp, err := client.do(context.Background(), http.MethodGet, workspacePath(workspace, "/collections"), nil)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type CollectionsImportCmd struct {
	Workspace string `help:"Workspace slug or ID. Defaults to active workspace or microwave.yaml."`
	Name      string `help:"Collection slug." required:""`
	Version   string `help:"Collection version." default:"latest"`
}

func (c *CollectionsImportCmd) Run(g *Globals) error {
	project, err := LoadProjectConfig(".")
	if err != nil {
		return err
	}
	workspace, err := workspaceOrActive(firstNonEmpty(c.Workspace, project.Workspace), g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"name":       c.Name,
		"version":    c.Version,
		"spec_path":  project.SpecPath,
		"config":     project.RawContents,
		"guides_dir": project.Docs.GuidesDir,
	}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/collections/import"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type CollectionsForkCmd struct {
	Workspace string `help:"Destination workspace slug or ID. Defaults to active workspace."`
	Source    string `arg:"" help:"Source collection path, for example microwave/stdlib/latest."`
	Name      string `help:"Destination collection slug."`
}

func (c *CollectionsForkCmd) Run(g *Globals) error {
	workspace, err := workspaceOrActive(c.Workspace, g)
	if err != nil {
		return err
	}
	client, err := g.client(true)
	if err != nil {
		return err
	}
	payload := map[string]any{"source": c.Source, "name": c.Name}
	resp, err := client.do(context.Background(), http.MethodPost, workspacePath(workspace, "/collections/fork"), payload)
	if err != nil {
		return err
	}
	return printAPIResponse(resp, g.isJSON())
}

type CollectionsOpenCmd struct {
	Workspace  string `arg:"" help:"Collection workspace slug."`
	Collection string `arg:"" help:"Collection slug."`
	Version    string `arg:"" optional:"" default:"latest" help:"Collection version."`
}

func (c *CollectionsOpenCmd) Run(g *Globals) error {
	collectionURL := fmt.Sprintf("https://microwave.sh/c/%s/%s/%s", url.PathEscape(c.Workspace), url.PathEscape(c.Collection), url.PathEscape(c.Version))
	if g.isJSON() {
		return printJSON(map[string]string{"url": collectionURL})
	}
	fmt.Println(collectionURL)
	return nil
}
