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
	"fmt"
	"os"
	"strings"
)

// VersionCmd prints the current CLI version.
type VersionCmd struct{}

func (c *VersionCmd) Run(g *Globals) error {
	fmt.Println(Version)
	return nil
}

// ConfigCmd groups local configuration helpers.
type ConfigCmd struct {
	Init     ConfigInitCmd     `cmd:"" help:"Create microwave.yaml for a spec repository."`
	Validate ConfigValidateCmd `cmd:"" help:"Validate local microwave.yaml and the referenced spec file."`
	Path     ConfigPathCmd     `cmd:"" help:"Print config file locations."`
}

type ConfigInitCmd struct {
	Workspace  string `help:"Workspace slug or ID to bind this repo to."`
	Spec       string `help:"OpenAPI spec path." default:"openapi.yaml"`
	GoRepo     string `name:"go-repo" help:"GitHub repo for the Go SDK target."`
	GoModule   string `name:"go-module" help:"Go module path for the Go SDK target."`
	PythonRepo string `name:"python-repo" help:"GitHub repo for the Python SDK target."`
	NodeRepo   string `name:"node-repo" help:"GitHub repo for the TypeScript/Node SDK target."`
	DocsDomain string `name:"docs-domain" help:"Custom docs domain."`
	Force      bool   `help:"Overwrite an existing microwave.yaml."`
}

func (c *ConfigInitCmd) Run(g *Globals) error {
	path := projectConfigPath(".")
	if _, err := os.Stat(path); err == nil && !c.Force {
		return fmt.Errorf("%s already exists. Pass --force to overwrite", projectConfigFile)
	}

	goRepo := c.GoRepo
	pythonRepo := c.PythonRepo
	nodeRepo := c.NodeRepo
	if c.Workspace != "" {
		if goRepo == "" {
			goRepo = c.Workspace + "/" + c.Workspace + "-go"
		}
		if pythonRepo == "" {
			pythonRepo = c.Workspace + "/" + c.Workspace + "-python"
		}
		if nodeRepo == "" {
			nodeRepo = c.Workspace + "/" + c.Workspace + "-node"
		}
	}

	module := c.GoModule
	if module == "" && goRepo != "" {
		module = "github.com/" + goRepo
	}

	docsDomain := c.DocsDomain
	if docsDomain == "" && c.Workspace != "" {
		docsDomain = c.Workspace + ".microwave.sh"
	}

	body := fmt.Sprintf(`version: %q
workspace: %q

spec:
  path: %s

sdks:
  go:
    repo: %s
    module: %s
    style:
      error_handling: idiomatic
      context: required
      options_pattern: functional
      pagination: iterator
  python:
    repo: %s
    style:
      typing: strict
      models: dataclasses
      async: dual
  node:
    repo: %s
    style:
      language: typescript
      types: branded
      unions: discriminated
      pagination: async_iterator

docs:
  domain: %s
  features:
    api_explorer: true
    changelog: true
    search: true
    versioning: true
  guides_dir: ./guides

notifications:
  on_sdk_pr: true
  on_breaking_change: true
`, defaultAPIVersion, c.Workspace, quoteYAML(c.Spec), quoteYAML(goRepo), quoteYAML(module), quoteYAML(pythonRepo), quoteYAML(nodeRepo), quoteYAML(docsDomain))

	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		return fmt.Errorf("write %s: %w", projectConfigFile, err)
	}
	fmt.Printf("Created %s\n", path)
	return nil
}

type ConfigValidateCmd struct{}

func (c *ConfigValidateCmd) Run(g *Globals) error {
	cfg, err := LoadProjectConfig(".")
	if err != nil {
		return err
	}
	specPath := cfg.ResolveSpecPath()
	if _, err := os.Stat(specPath); err != nil {
		return fmt.Errorf("spec.path %q does not exist", cfg.SpecPath)
	}
	languages := cfg.SDKLanguages()
	if len(languages) == 0 {
		return fmt.Errorf("%s has no sdks targets", projectConfigFile)
	}

	if g.isJSON() {
		return printJSON(map[string]any{
			"config":      cfg.Path,
			"workspace":   cfg.Workspace,
			"spec_path":   specPath,
			"sdk_targets": languages,
			"docs_domain": cfg.Docs.Domain,
		})
	}
	fmt.Printf("Config: %s\n", cfg.Path)
	fmt.Printf("Workspace: %s\n", valueOrDash(cfg.Workspace))
	fmt.Printf("Spec: %s\n", specPath)
	fmt.Printf("SDK targets: %s\n", strings.Join(languages, ", "))
	fmt.Printf("Docs domain: %s\n", valueOrDash(cfg.Docs.Domain))
	return nil
}

type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Run(g *Globals) error {
	if g.isJSON() {
		return printJSON(map[string]string{
			"global":  globalConfigPath(),
			"project": projectConfigPath("."),
		})
	}
	fmt.Printf("Global: %s\n", globalConfigPath())
	fmt.Printf("Project: %s\n", projectConfigPath("."))
	return nil
}

func quoteYAML(value string) string {
	if value == "" {
		return `""`
	}
	return fmt.Sprintf("%q", value)
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
