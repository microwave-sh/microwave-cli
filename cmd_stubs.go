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

// AddressCmd provides address and postal code operations.
type AddressCmd struct{}

func (c *AddressCmd) Run(g *Globals) error { fmt.Println("not yet implemented"); return nil }

// EncodingCmd provides encoding, hashing, and format conversion.
type EncodingCmd struct{}

func (c *EncodingCmd) Run(g *Globals) error { fmt.Println("not yet implemented"); return nil }

// FinancialCmd provides currency, FX, tax, and financial calculations.
type FinancialCmd struct{}

func (c *FinancialCmd) Run(g *Globals) error { fmt.Println("not yet implemented"); return nil }

// GeoCmd provides geospatial calculations and lookups.
type GeoCmd struct{}

func (c *GeoCmd) Run(g *Globals) error { fmt.Println("not yet implemented"); return nil }

// MathCmd provides mathematical calculations and statistics.
type MathCmd struct{}

func (c *MathCmd) Run(g *Globals) error { fmt.Println("not yet implemented"); return nil }

// TextCmd provides text parsing, transformation, and analysis.
type TextCmd struct{}

func (c *TextCmd) Run(g *Globals) error { fmt.Println("not yet implemented"); return nil }

// TimeCmd provides timezone, date, and scheduling operations.
type TimeCmd struct{}

func (c *TimeCmd) Run(g *Globals) error { fmt.Println("not yet implemented"); return nil }
