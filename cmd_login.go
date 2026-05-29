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
	"os"
)

// LoginCmd configures API credentials.
type LoginCmd struct {
	Key     string `arg:"" optional:"" help:"API key to store. If omitted, prints the console key URL."`
	BaseURL string `help:"Persist a custom API base URL."`
}

func (c *LoginCmd) Run(g *Globals) error {
	if c.Key == "" {
		fmt.Println("Create or reveal an API key in the Microwave Console:")
		fmt.Printf("  %s/keys\n\n", defaultConsoleURL)
		fmt.Println("Then run:")
		fmt.Println("  microwave login <api-key>")
		return nil
	}
	cfg := LoadGlobalConfig()
	cfg.Auth.APIKey = c.Key
	if c.BaseURL != "" {
		cfg.APIURL = c.BaseURL
	}
	if err := SaveGlobalConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("API key saved to %s\n", globalConfigPath())
	return nil
}

// LogoutCmd clears stored API credentials.
type LogoutCmd struct{}

func (c *LogoutCmd) Run(g *Globals) error {
	if err := clearGlobalAuth(); err != nil {
		return err
	}
	fmt.Println("Logged out.")
	return nil
}

// AuthCmd groups auth-related subcommands.
type AuthCmd struct {
	Whoami WhoamiCmd `cmd:"" help:"Print the authenticated identity."`
	Logout LogoutCmd `cmd:"" help:"Clear stored credentials."`
}

// WhoamiCmd prints the authenticated identity.
type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(g *Globals) error {
	client, err := g.client(true)
	if err != nil {
		return err
	}
	for _, path := range []string{"/workspaces/me", "/auth/whoami", "/me"} {
		resp, err := client.do(context.Background(), "GET", path, nil)
		if err == nil {
			return printAPIResponse(resp, g.isJSON())
		}
	}

	cfg := LoadGlobalConfig()
	key := g.APIKey
	if key == "" {
		key = cfg.Auth.APIKey
	}
	if g.isJSON() {
		return printJSON(map[string]string{"api_key": maskSecret(key)})
	}
	fmt.Fprintf(os.Stdout, "API key: %s\n", maskSecret(key))
	return nil
}
