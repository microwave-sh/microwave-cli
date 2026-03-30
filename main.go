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
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// CLI is the top-level kong command grammar.
type CLI struct {
	Globals

	Login     LoginCmd     `cmd:"" help:"Configure API credentials."`
	Auth      AuthCmd      `cmd:"" help:"Authentication commands."`
	Address   AddressCmd   `cmd:"" help:"Address and postal code operations."`
	Encoding  EncodingCmd  `cmd:"" help:"Encoding, hashing, and format conversion."`
	Financial FinancialCmd `cmd:"" help:"Currency, FX, tax, and financial calculations."`
	Geo       GeoCmd       `cmd:"" help:"Geospatial calculations and lookups."`
	Math      MathCmd      `cmd:"" help:"Mathematical calculations and statistics."`
	Text      TextCmd      `cmd:"" help:"Text parsing, transformation, and analysis."`
	Time      TimeCmd      `cmd:"" help:"Timezone, date, and scheduling operations."`
}

func main() {
	if os.Getenv("NO_COLOR") == "" && termenv.ColorProfile() == termenv.Ascii {
		lipgloss.SetColorProfile(termenv.TrueColor)
	}

	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("microwave"),
		kong.Description("The Microwave CLI — stateless utility APIs at your fingertips."),
		kong.UsageOnError(),
	)
	if err := ctx.Run(&cli.Globals); err != nil {
		ctx.FatalIfErrorf(err)
	}
}
