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

import microwave "github.com/microwave-sh/microwave-go"

// Globals holds flags available to every command.
type Globals struct {
	APIKey  string `env:"MICROWAVE_API_KEY" help:"Microwave API key. Defaults to MICROWAVE_API_KEY env var." short:"k"`
	BaseURL string `env:"MICROWAVE_BASE_URL" help:"Override the API base URL." default:"https://api.microwave.dev" hidden:""`
}

func (g *Globals) client() (*microwave.Client, error) {
	return microwave.NewClient(
		microwave.WithBaseURL(g.BaseURL),
		microwave.WithAPIKey(g.APIKey),
	)
}
