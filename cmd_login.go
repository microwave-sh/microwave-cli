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

import "fmt"

// LoginCmd configures API credentials.
type LoginCmd struct {
	Key string `arg:"" optional:"" help:"API key to store. If omitted, opens the browser to retrieve one."`
}

func (c *LoginCmd) Run(g *Globals) error {
	if c.Key == "" {
		fmt.Println("Opening https://microwave.dev/keys in your browser...")
		// TODO: open browser + token exchange flow
		return nil
	}
	// TODO: persist key to config file
	fmt.Printf("API key saved.\n")
	return nil
}

// AuthCmd groups auth-related subcommands.
type AuthCmd struct {
	Whoami WhoamiCmd `cmd:"" help:"Print the authenticated identity."`
}

// WhoamiCmd prints the authenticated identity.
type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(g *Globals) error {
	// TODO: call /auth/whoami
	fmt.Printf("API key: %s\n", g.APIKey)
	return nil
}
