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
	"net/http"
)

// Globals holds flags available to every command.
type Globals struct {
	APIKey  string `env:"MICROWAVE_API_KEY" help:"Microwave API key. Defaults to MICROWAVE_API_KEY env var." short:"k"`
	BaseURL string `env:"MICROWAVE_BASE_URL" help:"Override the API base URL." hidden:""`
	Output  string `name:"output" short:"o" enum:"table,json" default:"table" help:"Output format."`
}

func (g *Globals) client(requireAuth bool) (*managementClient, error) {
	cfg := LoadGlobalConfig()
	apiKey := g.APIKey
	if apiKey == "" {
		apiKey = cfg.Auth.APIKey
	}
	if requireAuth && apiKey == "" {
		return nil, fmt.Errorf("not logged in. Run `microwave login <api-key>` or set MICROWAVE_API_KEY")
	}

	baseURL := g.BaseURL
	if baseURL == "" {
		baseURL = cfg.APIURL
	}
	if baseURL == "" {
		baseURL = defaultAPIURL
	}

	return &managementClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		apiVersion: defaultAPIVersion,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
	}, nil
}

func (g *Globals) activeWorkspace() string {
	return LoadGlobalConfig().Workspace.Active
}

func (g *Globals) isJSON() bool {
	return g.Output == "json"
}
