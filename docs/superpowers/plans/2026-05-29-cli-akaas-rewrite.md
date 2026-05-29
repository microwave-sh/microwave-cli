# Microwave CLI AKaaS Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild `microwave-cli` so its command set exactly mirrors what the Microwave server (`microwave-api` AKaaS surface + `microwave-auth` plane) can actually do today, and bring it into full conformance with `mataki-dev/playbooks/CLI.md` including the charmbracelet output stack, mandatory commands, and the release/distribution pipeline.

**Architecture:** The current CLI targets a non-existent product API (`/workspaces/{ws}/specs|sdk|docs|collections`). Those commands are removed. The CLI is re-scaffolded to the CLI.md layout (`cmd/`, `internal/{client,config,output}`), authenticates with a Microwave management API key via `Authorization: Bearer`, and exposes the real AKaaS resources: `permission-sets`, `key-specs` (+ key issuance/events/widget-sessions), `keys` (+ verify/rotate/events), `signing-key-sets` (+ key/secret lifecycle + sign), `trust-exchanges`, `trust-providers` (+ token mint). Distribution via GoReleaser → Homebrew + `curl|sh` + GitHub releases.

**Tech Stack:** Go 1.26, Kong (CLI parsing), Charmbracelet (bubbletea/bubbles/lipgloss), BurntSushi/toml, GoReleaser, semantic-release.

---

## Scope

- **In:** Full re-scaffold per CLI.md; charm output stack (theme/spinner/table/banners); mandatory commands (`login`, `logout`, `whoami`, `version`, `completion`); `--output table|json` (D8); `--debug` (D9); update check (D17); signal/context (D18); the six AKaaS resource command groups + auth-plane verify/mint; full distribution (GoReleaser, `.releaserc.json`, Makefile, CI/release/build workflows, install script, Homebrew formula wiring); README per D22.
- **Out:** `workspace`, `spec`, `sdk`, `docs`, `collections`, `console` commands — the server implements none of these; they are deleted. Device-code/GitHub-OIDC login — the management API issues no tokens to the CLI (keys are created in the console), so `login` stores a pasted key. Windows targets (CLI.md D12 / Appendix A — defer).

## Decided configuration values (use these literally)

| Thing | Value |
|-------|-------|
| Go module | `github.com/microwave-sh/microwave-cli` |
| Binary / product name | `microwave` |
| GitHub org | `microwave-sh` |
| CLI repo | `microwave-sh/microwave-cli` |
| Default API base URL | `https://api.microwave.sh` *(see ASSUMPTION-1)* |
| Default auth-plane base URL | `https://auth.microwave.sh` *(see ASSUMPTION-1)* |
| Console URL (printed by `login`) | `https://app.microwave.sh` |
| `API-Version` header value | `2026-05-27` |
| Token env var | `MICROWAVE_TOKEN` (accept legacy `MICROWAVE_API_KEY` as fallback) |
| API URL env var | `MICROWAVE_API_URL` |
| No-update-check env var | `MICROWAVE_NO_UPDATE_CHECK` |
| Config path | `$XDG_CONFIG_HOME/microwave/config.toml` else `~/.config/microwave/config.toml` |
| Homebrew tap repo | `microwave-sh/homebrew-microwave` (exists at `/Users/sethyates/mataki/microwave/homebrew-microwave`) |
| `ColorPrimary` / `ColorAccent` | Read `microwave-app` Tailwind/CSS `--primary` token; convert to hex. Fallback `ColorPrimary = "#2563eb"`, `ColorAccent = "#1e293b"`. |

**ASSUMPTION-1 (flag in final report):** the existing CLI defaulted to `https://api.microwave.dev`, but every other Microwave surface uses `microwave.sh`. This plan standardizes on `api.microwave.sh` / `auth.microwave.sh`, both overridable via `MICROWAVE_API_URL` + config. If the deployed hostname is actually `.dev`, only the two default constants change.

## Server surface → CLI command map (the completeness contract)

Management API (`microwave-api`, `Authorization: Bearer <key>`, header `API-Version: 2026-05-27`). Workspace is derived from the key — **never** a URL segment or flag.

| Resource | Server routes | CLI commands |
|----------|--------------|--------------|
| permission-sets | `POST /api/permission-sets`, `POST /api/permission-sets/search`, `PATCH/DELETE /api/permission-sets/{id}` | `permission-sets list\|create\|update\|delete` (`get` = search by id) |
| key-specs | `POST /api/key-specs`, `/search`, `PATCH/DELETE /{spec_id}`, `GET /{spec_id}/events`, `POST /{spec_id}/keys`, `/{spec_id}/keys/search`, `/{spec_id}/keys/revoke-by-subject`, `/{spec_id}/widget-sessions` | `key-specs list\|create\|update\|delete\|events`; `key-specs keys issue\|list\|revoke-by-subject`; `key-specs widget-session` |
| keys | `POST /api/keys/search`, `GET/PATCH /{key_id}`, `POST /{key_id}/revoke`, `/{key_id}/rotate`, `GET /{key_id}/events`, `POST /api/keys/verify` | `keys list\|get\|update\|revoke\|rotate\|events\|verify` |
| signing-key-sets | `POST /api/signing-key-sets`, `/search`, `GET/PATCH/DELETE /{kind}/{set_name}`, `POST .../keys/generate`, `.../keys/{key_id}/activate`, `.../keys/{key_id}/revoke`, `GET .../keys/{key_id}/secret`, `POST .../sign`, `GET .../secret`, `POST .../secret/rotate` | `signing-key-sets list\|get\|create\|update\|delete\|sign\|secret\|rotate-secret`; `signing-key-sets keys generate\|activate\|revoke\|secret` |
| trust-exchanges | `POST /api/trust-exchanges`, `/search`, `GET/PATCH/DELETE /{exchange_id}` | `trust-exchanges list\|get\|create\|update\|delete` |
| trust-providers | `POST /api/trust-providers`, `/search`, `GET/PATCH/DELETE /{provider_id}` | `trust-providers list\|get\|create\|update\|delete` |

Auth-plane (`microwave-auth`, public, no management key except mint which takes a key as bearer):

| Route | CLI command |
|-------|-------------|
| `POST /v1/verify` (also `POST /api/keys/verify` on mgmt) | covered by `keys verify` (uses mgmt endpoint) |
| `POST /trust-providers/{id}/token` | `trust-providers mint <id>` |
| `GET /trust-providers/{id}/.well-known/{openid-configuration,jwks.json}` | `trust-providers discovery <id>` (prints URLs + optionally fetches) |

## Target command tree

```text
microwave
├── login [<key>]            Store a management API key (prints console URL if omitted)
├── logout                   Clear stored credentials
├── whoami                   Verify the stored key; print identity (subject/scopes/workspace)
├── version                  Print CLI version
├── completion <shell>       Print bash|zsh|fish completion script
├── permission-sets
│   ├── list                 Search permission sets
│   ├── create               Create a permission set (--name, --description, --permission)
│   ├── update <id>          Update a permission set
│   └── delete <id>          Delete a permission set
├── key-specs
│   ├── list                 Search key specs
│   ├── create               Create a key spec
│   ├── update <id>          Update a key spec
│   ├── delete <id>          Delete a key spec
│   ├── events <id>          List key-spec events (--subject)
│   ├── widget-session <id>  Create a widget session token
│   └── keys
│       ├── issue <id>       Issue a key from a spec
│       ├── list <id>        Search keys for a spec
│       └── revoke-by-subject <id>  Revoke all keys for a subject
├── keys
│   ├── list                 Search issued keys
│   ├── get <id>             Get an issued key
│   ├── update <id>          Update a key (--name, --scopes, --expires-at)
│   ├── revoke <id>          Revoke a key
│   ├── rotate <id>          Rotate a key (--overlap-seconds)
│   ├── events <id>          List key events
│   └── verify <key>         Verify a key
├── signing-key-sets
│   ├── list                 Search signing key sets
│   ├── get <kind> <name>    Get a signing key set + keys
│   ├── create               Create a signing key set (--kind, --name, --algorithm)
│   ├── update <kind> <name> Rename a signing key set
│   ├── delete <kind> <name> Delete a signing key set
│   ├── sign <kind> <name>   Sign a JWT payload (asymmetric)
│   ├── secret <kind> <name> Reveal symmetric secret state
│   ├── rotate-secret <kind> <name>  Rotate symmetric secret
│   └── keys
│       ├── generate <kind> <name>            Generate a new key
│       ├── activate <kind> <name> <key-id>   Activate a key (symmetric)
│       ├── revoke <kind> <name> <key-id>     Revoke a key
│       └── secret <kind> <name> <key-id>     Reveal a key secret (symmetric)
├── trust-exchanges
│   ├── list / get <id> / create / update <id> / delete <id>
└── trust-providers
    ├── list / get <id> / create / update <id> / delete <id>
    ├── mint <id>            Mint a token (--api-key, --subject, --audience, --claims, --ttl)
    └── discovery <id>       Print discovery/JWKS/token URLs
```

## File structure

Repo: `/Users/sethyates/mataki/microwave/microwave-cli`

**Delete (vapor commands against non-existent API):** `cmd_workspace_spec.go`, `cmd_sdk_docs_collections.go`, `cmd_keys_console.go`, `cmd_stubs.go`, `cmd_login.go`, `cmd_config.go`, and the old flat `client.go`, `config.go`, `globals.go`, `version.go`, `main.go`, `cli_test.go`, `config_test.go` (rebuilt under the new layout). Keep `go.mod`, `go.sum`, `.git`, `.gitignore`, `.pre-commit-config.yaml`, `.markdownlint.yaml`, `README.md` (rewritten).

**New layout:**
- `main.go` — Kong root struct, signal/context, `kong.Bind(ctx)`, version, update-check call.
- `internal/version/version.go` — `var Version = "dev"`.
- `cmd/globals.go` — `Globals` struct (Token, APIURL, AuthURL, Output, Debug, Version) + `Client()`, `AuthClient()`, `IsJSON()`.
- `cmd/login.go`, `cmd/logout.go`, `cmd/whoami.go`, `cmd/version.go`, `cmd/completion.go`.
- `cmd/permission_sets.go`, `cmd/key_specs.go`, `cmd/keys.go`, `cmd/signing_key_sets.go`, `cmd/trust_exchanges.go`, `cmd/trust_providers.go`.
- `internal/client/client.go` — HTTP client (Bearer + API-Version + debug), `do()`, error mapping.
- `internal/client/search.go` — search request/response envelope + `Search[T]` helper.
- `internal/client/types.go` — all AKaaS request/response types.
- `internal/client/akaas.go` — typed client methods per resource.
- `internal/config/config.go` — XDG/TOML, token + API URL resolution, write/clear.
- `internal/config/update.go` — cached update check.
- `internal/output/theme.go`, `internal/output/output.go` — colors, spinner, table/json, banners, helpers.
- Tooling: `.goreleaser.yaml`, `.releaserc.json`, `Makefile`, `.github/workflows/{ci,release,build}.yml`, `scripts/install.sh`, `README.md`.
- Tests: `internal/config/config_test.go`, `internal/output/output_test.go`, `cmd/*_test.go`.

---

### Task 1: Re-scaffold to the CLI.md layout

**Files:**
- Delete: `cmd_workspace_spec.go`, `cmd_sdk_docs_collections.go`, `cmd_keys_console.go`, `cmd_stubs.go`, `cmd_login.go`, `cmd_config.go`, `client.go`, `config.go`, `globals.go`, `version.go`, `main.go`, `cli_test.go`, `config_test.go`
- Create: `internal/version/version.go`, `main.go`, `cmd/globals.go`, `cmd/version.go`

- [ ] **Step 1: Remove the vapor files**

```bash
cd /Users/sethyates/mataki/microwave/microwave-cli
git rm cmd_workspace_spec.go cmd_sdk_docs_collections.go cmd_keys_console.go cmd_stubs.go cmd_login.go cmd_config.go client.go config.go globals.go version.go main.go cli_test.go config_test.go
```

- [ ] **Step 2: Create `internal/version/version.go`**

```go
package version

// Version is set via -ldflags at build time; semantic-release manages tags.
var Version = "dev"
```

- [ ] **Step 3: Create `cmd/globals.go`**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/config"
)

type Globals struct {
	Token   string
	APIURL  string
	AuthURL string
	Output  string
	Debug   bool
	Version string
}

func (g *Globals) resolveToken() (string, error) {
	if g.Token != "" {
		return g.Token, nil
	}
	return config.ResolveToken(config.GlobalConfigDir())
}

// Client returns an authenticated management API client. Exits 1 if no token.
func (g *Globals) Client() *client.Client {
	token, err := g.resolveToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	return client.New(g.apiURL(), token, g.Version, g.Debug)
}

// AuthClient returns an unauthenticated client for the public auth plane.
func (g *Globals) AuthClient() *client.Client {
	return client.New(g.authURL(), "", g.Version, g.Debug)
}

func (g *Globals) apiURL() string {
	if g.APIURL != "" {
		return g.APIURL
	}
	return config.ResolveAPIURL()
}

func (g *Globals) authURL() string {
	if g.AuthURL != "" {
		return g.AuthURL
	}
	return config.ResolveAuthURL()
}

func (g *Globals) IsJSON() bool { return g.Output == "json" }
```

- [ ] **Step 4: Create `cmd/version.go`**

```go
package cmd

import "fmt"

type VersionCmd struct{}

func (c *VersionCmd) Run(g *Globals) error {
	fmt.Println(g.Version)
	return nil
}
```

- [ ] **Step 5: Create `main.go`**

Follow CLI.md's [Signal Handling and Context Cancellation](#) example exactly, adapted to these values: product `microwave`, module `github.com/microwave-sh/microwave-cli`, version from `internal/version`. The Kong root struct registers ONLY the commands in the target tree (no workspace/spec/sdk/docs/collections):

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/microwave-sh/microwave-cli/cmd"
	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/version"
)

type CLI struct {
	Token   string `env:"MICROWAVE_TOKEN" help:"Management API key." hidden:""`
	APIURL  string `name:"api-url" env:"MICROWAVE_API_URL" help:"API base URL override." hidden:""`
	AuthURL string `name:"auth-url" env:"MICROWAVE_AUTH_URL" help:"Auth-plane base URL override." hidden:""`
	Output  string `name:"output" short:"o" enum:"table,json" default:"table" help:"Output format (table, json)."`
	Debug   bool   `name:"debug" help:"Enable debug logging."`

	Login          cmd.LoginCmd          `cmd:"" help:"Store a management API key."`
	Logout         cmd.LogoutCmd         `cmd:"" help:"Clear stored credentials."`
	Whoami         cmd.WhoamiCmd         `cmd:"" help:"Print the authenticated identity."`
	Version        cmd.VersionCmd        `cmd:"" help:"Print version."`
	Completion     cmd.CompletionCmd     `cmd:"" help:"Print shell completion script."`
	PermissionSets cmd.PermissionSetsCmd `cmd:"" name:"permission-sets" help:"Manage permission sets."`
	KeySpecs       cmd.KeySpecsCmd       `cmd:"" name:"key-specs" help:"Manage key specs."`
	Keys           cmd.KeysCmd           `cmd:"" help:"Manage issued keys."`
	SigningKeySets cmd.SigningKeySetsCmd `cmd:"" name:"signing-key-sets" help:"Manage signing key sets."`
	TrustExchanges cmd.TrustExchangesCmd `cmd:"" name:"trust-exchanges" help:"Manage trust exchanges."`
	TrustProviders cmd.TrustProvidersCmd `cmd:"" name:"trust-providers" help:"Manage trust providers."`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cli := CLI{}
	kctx := kong.Parse(&cli,
		kong.Name("microwave"),
		kong.Description("Microwave — API key, JWKS, and OIDC federation management."),
		kong.UsageOnError(),
		kong.Bind(ctx),
	)

	config.CheckForUpdate(version.Version)

	err := kctx.Run(&cmd.Globals{
		Token:   cli.Token,
		APIURL:  cli.APIURL,
		AuthURL: cli.AuthURL,
		Output:  cli.Output,
		Debug:   cli.Debug,
		Version: version.Version,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, cmd.RenderError(err))
		os.Exit(1)
	}
}
```

NOTE: `cmd.RenderError`, `cmd.LoginCmd`, the resource command types, and the `internal/client`/`internal/config`/`internal/output` packages are created in later tasks. This task will NOT compile until Tasks 2–5 land. That is expected; do not stub them here — commit this task's files and let the build go green at the end of Task 5. (If you prefer a green commit now, temporarily comment out the resource command fields and `RenderError`; the subagent review for Task 1 only checks the scaffold shape.)

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor: remove vapor commands and scaffold CLI.md layout"
```

---

### Task 2: Output stack (theme, spinner, table, banners)

**Files:**
- Create: `internal/output/theme.go`, `internal/output/output.go`, `internal/output/output_test.go`

- [ ] **Step 1: Create `internal/output/theme.go`**

Follow CLI.md's [Theme File](#) example verbatim. Substitute the Microwave brand colors: read `microwave-app`'s Tailwind/CSS `--primary` token and convert to hex for `ColorPrimary`; use `--accent`/sidebar accent for `ColorAccent`. If you cannot resolve them, use fallback `ColorPrimary = lipgloss.Color("#2563eb")` and `ColorAccent = lipgloss.Color("#1e293b")`. Keep the shared semantic colors (success/error/warning/pending/cancelled), `statusColors` map, `ColorStatus`, `SuccessBanner`, `ErrorBanner`, and the reusable styles (Green/Red/Yellow/Bold/Dim/HeaderStyle/KeyStyle) exactly as in the playbook.

- [ ] **Step 2: Create `internal/output/output.go`**

Include, following the CLI.md examples verbatim where present:
1. `Spinner` (the full bubbletea Tier-2 [Spinner](#) example: `NewSpinner`, `Stop`, `Fail`, `Cancelled`, `CancelledC`).
2. `PrintTable(headers []string, rows [][]string, jsonOutput bool)` + `printJSON` (the [Table Output with Format Switching](#) example).
3. `PrintJSON(v any)` — marshal any value with 2-space indent to stdout (used by resource `get`/`create` outputs):

```go
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
```

4. `FormatTimeAgo(t time.Time) string` — relative time ("3m ago", "2d ago", "—" for zero):

```go
func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
```

- [ ] **Step 3: Write a unit test `internal/output/output_test.go`**

```go
package output

import (
	"testing"
	"time"
)

func TestFormatTimeAgo_Zero(t *testing.T) {
	if got := FormatTimeAgo(time.Time{}); got != "—" {
		t.Fatalf("FormatTimeAgo(zero) = %q, want —", got)
	}
}

func TestFormatTimeAgo_Hours(t *testing.T) {
	got := FormatTimeAgo(time.Now().Add(-3 * time.Hour))
	if got != "3h ago" {
		t.Fatalf("FormatTimeAgo(-3h) = %q, want 3h ago", got)
	}
}

func TestColorStatus_Unknown_PassThrough(t *testing.T) {
	if got := ColorStatus("weird"); got != "weird" {
		t.Fatalf("ColorStatus(weird) = %q, want weird", got)
	}
}
```

- [ ] **Step 4: Verify**

```bash
go test ./internal/output/... -count=1
```
Expected: PASS (3 tests). The package compiles independently of the rest.

- [ ] **Step 5: Commit**

```bash
git add internal/output
git commit -m "feat: add charmbracelet output stack (theme, spinner, table)"
```

---

### Task 3: Config and update check

**Files:**
- Create: `internal/config/config.go`, `internal/config/update.go`, `internal/config/config_test.go`

- [ ] **Step 1: Write failing config tests `internal/config/config_test.go`**

```go
package config_test

import (
	"path/filepath"
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/config"
)

func TestResolveToken_EnvOverridesConfig(t *testing.T) {
	t.Setenv("MICROWAVE_TOKEN", "env-token")
	got, err := config.ResolveToken(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "env-token" {
		t.Fatalf("ResolveToken = %q, want env-token", got)
	}
}

func TestResolveToken_LegacyEnvFallback(t *testing.T) {
	t.Setenv("MICROWAVE_API_KEY", "legacy")
	got, err := config.ResolveToken(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "legacy" {
		t.Fatalf("ResolveToken = %q, want legacy", got)
	}
}

func TestWriteAndResolveToken_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := config.WriteGlobalAuthTo(dir, "stored-token"); err != nil {
		t.Fatal(err)
	}
	got, err := config.ResolveToken(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "stored-token" {
		t.Fatalf("ResolveToken = %q, want stored-token", got)
	}
}

func TestResolveToken_NotLoggedIn(t *testing.T) {
	if _, err := config.ResolveToken(filepath.Join(t.TempDir(), "empty")); err == nil {
		t.Fatal("expected error when no token present")
	}
}
```

- [ ] **Step 2: Run, verify red**

```bash
go test ./internal/config/... -count=1
```
Expected: compile failure (functions undefined).

- [ ] **Step 3: Create `internal/config/config.go`**

Follow CLI.md's [Config File Handling](#) example, with these concrete adaptations:
- Product = `microwave`; `GlobalConfigDir()` returns `$XDG_CONFIG_HOME/microwave` or `~/.config/microwave`.
- `GlobalConfig` struct adds `AuthURL string` toml `auth_url,omitempty` alongside `APIURL` (`api_url,omitempty`) and `Auth.Token` (`auth.token`).
- `ResolveToken(globalDir string) (string, error)`: priority `MICROWAVE_TOKEN` env → `MICROWAVE_API_KEY` env (legacy) → `config.toml` `auth.token` → error `"not logged in. Run \`microwave login\` to authenticate"`.
- Add `WriteGlobalAuth(token string) error` (writes to `GlobalConfigDir()`, perms `0600`) AND a testable `WriteGlobalAuthTo(dir, token string) error` it delegates to.
- `ResolveAPIURL() string`: `MICROWAVE_API_URL` env → config `api_url` → default `"https://api.microwave.sh"`.
- `ResolveAuthURL() string`: `MICROWAVE_AUTH_URL` env → config `auth_url` → default `"https://auth.microwave.sh"`.
- `ClearAuth() error`: remove the config file (ignore `os.IsNotExist`). Used by `logout`.

Use `github.com/BurntSushi/toml` (CLI.md D4). Add it to `go.mod`.

- [ ] **Step 4: Create `internal/config/update.go`**

Follow CLI.md's [Update Check](#) example verbatim, substituting: product `microwave`, env `MICROWAVE_NO_UPDATE_CHECK`, repo `microwave-sh/microwave-cli`, upgrade hint `brew upgrade microwave  or  curl -sSL https://microwave.sh/install.sh | sh`.

- [ ] **Step 5: Run, verify green**

```bash
go test ./internal/config/... -count=1
```
Expected: PASS (4 tests).

- [ ] **Step 6: Commit**

```bash
git add internal/config go.mod go.sum
git commit -m "feat: add XDG/TOML config, token resolution, update check"
```

---

### Task 4: HTTP client + search envelope

**Files:**
- Create: `internal/client/client.go`, `internal/client/search.go`, `internal/client/client_test.go`

- [ ] **Step 1: Create `internal/client/client.go`**

```go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const apiVersion = "2026-05-27"

type Client struct {
	baseURL string
	token   string
	version string
	debug   bool
	http    *http.Client
}

func New(baseURL, token, version string, debug bool) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		version: version,
		debug:   debug,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// APIError is a structured error from the platform/errors envelope.
type APIError struct {
	Status  int
	Type    string `json:"type"`
	Message string `json:"message"`
	Errors  []struct {
		Field   string `json:"field"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		parts := make([]string, len(e.Errors))
		for i, fe := range e.Errors {
			parts[i] = fmt.Sprintf("%s: %s", fe.Field, fe.Message)
		}
		return fmt.Sprintf("%s (%s)", e.Message, strings.Join(parts, "; "))
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("request failed with status %d", e.Status)
}

// Do sends a request and decodes a JSON response into out (out may be nil).
func (c *Client) Do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(raw)
		if c.debug {
			log.Printf("→ %s %s\n%s", method, path, raw)
		}
	} else if c.debug {
		log.Printf("→ %s %s", method, path)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Version", apiVersion)
	req.Header.Set("User-Agent", "microwave-cli/"+c.version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if c.debug {
		log.Printf("← %d\n%s", resp.StatusCode, raw)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{Status: resp.StatusCode}
		_ = json.Unmarshal(raw, apiErr)
		return apiErr
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

var _ = os.Stdout // ensure os import retained if trimmed
```

(Remove the trailing `var _ = os.Stdout` and the `os` import if unused after implementation — it's only there to remind you not to leave an unused import; `gofmt`/`go vet` will catch it.)

- [ ] **Step 2: Create `internal/client/search.go`**

The platform/search envelope: request `{filter, sort, limit, cursor}`, response `{data, next_cursor, has_more, limit}`. Confirm exact request field names against `mataki-dev/platform/search` (the server uses `filter` as a nested object `{field:{op:value}}` and `sort` as `[{field,direction}]` per CLI.md's Request shape). Implement:

```go
package client

import "context"

type SortDirective struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" | "desc"
}

type SearchRequest struct {
	Filter map[string]map[string]any `json:"filter,omitempty"`
	Sort   []SortDirective           `json:"sort,omitempty"`
	Limit  int                       `json:"limit,omitempty"`
	Cursor string                    `json:"cursor,omitempty"`
}

type SearchResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

// Search POSTs to {path}/search and decodes a typed page.
func Search[T any](ctx context.Context, c *Client, path string, req SearchRequest) (*SearchResponse[T], error) {
	var page SearchResponse[T]
	if err := c.Do(ctx, "POST", path+"/search", req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
```

VERIFY the request shape against the actual `platform/search` request decoder in the server before finalizing — if the server expects `filters: [{field,operator,value}]` (array form) instead of the nested `filter` object, switch `SearchRequest` to the array form the server's `search.Validate` accepts. Read `internal/api/adapter/inbound/http`'s search registration + the platform package to confirm. This is the one shape the exploration was less certain about; resolve it by reading the server, not guessing.

- [ ] **Step 3: Write `internal/client/client_test.go`**

```go
package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDo_SetsAuthAndVersionHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing bearer: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("API-Version") != apiVersion {
			t.Errorf("API-Version = %q", r.Header.Get("API-Version"))
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok", "test", false)
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.Do(context.Background(), "GET", "/api/keys/search", nil, &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatal("expected ok")
	}
}

func TestDo_MapsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"type":"invalid_input","message":"bad","errors":[{"field":"name","message":"required"}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok", "test", false)
	err := c.Do(context.Background(), "POST", "/api/key-specs", map[string]string{}, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type = %T, want *APIError", err)
	}
	if apiErr.Status != 422 || apiErr.Type != "invalid_input" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}
```

- [ ] **Step 4: Verify**

```bash
go test ./internal/client/... -count=1
```
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/client
git commit -m "feat: add HTTP client and search envelope"
```

---

### Task 5: Typed AKaaS client (types + methods)

**Files:**
- Create: `internal/client/types.go`, `internal/client/akaas.go`

- [ ] **Step 1: Create `internal/client/types.go`**

These mirror the server DTOs exactly (field names + json tags are authoritative — copied from the server's `internal/api/adapter/inbound/http/dto`). Verify against the server while implementing; if a field differs, the server wins.

```go
package client

import "time"

// ── Permission sets ──────────────────────────────────────────────
type Permission struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Dangerous   bool   `json:"dangerous"`
}
type PermissionSetInput struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
}
type PermissionSet struct {
	ID          string       `json:"id"`
	WorkspaceID string       `json:"workspace_id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ── Key specs ────────────────────────────────────────────────────
type OpaqueConfig struct {
	Prefix         string `json:"prefix,omitempty"`
	LookupResponse string `json:"lookup_response,omitempty"`
}
type JWTConfig struct {
	Algorithm string `json:"algorithm,omitempty"`
	Issuer    string `json:"issuer,omitempty"`
	Audience  string `json:"audience,omitempty"`
}
type ExpiryPolicy struct {
	DefaultTTL           string `json:"default_ttl,omitempty"`
	MaxTTL               string `json:"max_ttl,omitempty"`
	AllowNever           bool   `json:"allow_never,omitempty"`
	RotationReminderDays int    `json:"rotation_reminder_days,omitempty"`
}
type ClaimField struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
}
type ClaimsConfig struct {
	Standard []string              `json:"standard,omitempty"`
	Custom   map[string]ClaimField `json:"custom,omitempty"`
}
type OverridePolicy struct {
	AllowCustomExpiry bool `json:"allow_custom_expiry,omitempty"`
	AllowCustomScopes bool `json:"allow_custom_scopes,omitempty"`
	AllowCustomClaims bool `json:"allow_custom_claims,omitempty"`
}
type WebhookConfig struct {
	Endpoint string   `json:"endpoint,omitempty"`
	Events   []string `json:"events,omitempty"`
}
type KeySpecInput struct {
	Name                   string         `json:"name"`
	Description            string         `json:"description,omitempty"`
	Format                 string         `json:"format"` // "opaque" | "jwt"
	PermissionSetID        string         `json:"permission_set_id,omitempty"`
	SigningKeySetID        string         `json:"signing_key_set_id,omitempty"`
	Opaque                 OpaqueConfig   `json:"opaque,omitempty"`
	JWT                    JWTConfig      `json:"jwt,omitempty"`
	Expiry                 ExpiryPolicy   `json:"expiry"`
	Claims                 ClaimsConfig   `json:"claims"`
	OverridePolicy         OverridePolicy `json:"override_policy"`
	Webhooks               WebhookConfig  `json:"webhooks"`
	WebhookSigningKeySetID string         `json:"webhook_signing_key_set_id,omitempty"`
}
type KeySpec struct {
	ID                     string         `json:"id"`
	WorkspaceID            string         `json:"workspace_id"`
	Name                   string         `json:"name"`
	Description            string         `json:"description,omitempty"`
	Format                 string         `json:"format"`
	PermissionSetID        string         `json:"permission_set_id,omitempty"`
	SigningKeySetID        string         `json:"signing_key_set_id,omitempty"`
	WidgetURL              string         `json:"widget_url,omitempty"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

// ── Issued keys ──────────────────────────────────────────────────
type IssueKeyInput struct {
	Subject   string         `json:"subject"`
	Name      string         `json:"name"`
	Scopes    []string       `json:"scopes,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	ExpiresIn string         `json:"expires_in,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}
type IssueKeyResult struct {
	ID        string     `json:"id"`
	Key       string     `json:"key"`
	KeyHint   string     `json:"key_hint"`
	Subject   string     `json:"subject"`
	SpecID    string     `json:"spec_id"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
type IssuedKey struct {
	ID             string         `json:"id"`
	SpecID         string         `json:"spec_id"`
	WorkspaceID    string         `json:"workspace_id"`
	Subject        string         `json:"subject"`
	Name           string         `json:"name"`
	KeyHint        string         `json:"key_hint"`
	Scopes         []string       `json:"scopes"`
	Claims         map[string]any `json:"claims,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	Status         string         `json:"status"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	LastVerifiedAt *time.Time     `json:"last_verified_at,omitempty"`
	RevokedAt      *time.Time     `json:"revoked_at,omitempty"`
}
type UpdateKeyInput struct {
	Name      *string        `json:"name,omitempty"`
	Scopes    []string       `json:"scopes,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
}
type RotateKeyInput struct {
	OverlapSeconds int `json:"overlap_seconds,omitempty"`
}
type VerifyKeyInput struct {
	Key string `json:"key"`
}
type VerifyKeyResult struct {
	Valid     bool           `json:"valid"`
	Code      string         `json:"code,omitempty"`
	KeyID     string         `json:"key_id,omitempty"`
	Subject   string         `json:"subject,omitempty"`
	Scopes    []string       `json:"scopes,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	JWT       string         `json:"jwt,omitempty"`
}
type KeyEvent struct {
	ID        string    `json:"id"`
	KeyID     string    `json:"key_id"`
	SpecID    string    `json:"spec_id"`
	Subject   string    `json:"subject"`
	Type      string    `json:"type"`
	IP        string    `json:"ip,omitempty"`
	Actor     string    `json:"actor"`
	Timestamp time.Time `json:"timestamp"`
}
type RevokeBySubjectInput struct {
	Subject string `json:"subject"`
}
type WidgetSessionInput struct {
	Subject     string         `json:"subject"`
	Claims      map[string]any `json:"claims,omitempty"`
	Scopes      []string       `json:"scopes,omitempty"`
	RedirectURL string         `json:"redirect_url,omitempty"`
	TTL         string         `json:"ttl,omitempty"`
}
type WidgetSessionToken struct {
	SessionToken string `json:"session_token"`
}

// ── Signing key sets ─────────────────────────────────────────────
type SigningKeySetInput struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`      // "asymmetric" | "symmetric"
	Algorithm string `json:"algorithm"` // RS256/ES256/HS256/...
}
type SigningKeySet struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Kind      string     `json:"kind"`
	Algorithm string     `json:"algorithm"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
type AsymmetricPublicMaterial struct {
	Kty, Alg, Use, N, E, X, Y, Crv string
}
type SigningKey struct {
	ID                       string                    `json:"id"`
	SetID                    string                    `json:"set_id"`
	Status                   string                    `json:"status"`
	AsymmetricPublicMaterial *AsymmetricPublicMaterial `json:"asymmetric_public_material,omitempty"`
	SecretRef                string                    `json:"secret_ref,omitempty"`
	CreatedAt                time.Time                 `json:"created_at"`
	RevokedAt                *time.Time                `json:"revoked_at,omitempty"`
}
type SigningKeySetDetail struct {
	Set  SigningKeySet `json:"set"`
	Keys []SigningKey  `json:"keys"`
}
type SigningKeySetUpdateInput struct {
	Name string `json:"name"`
}
type SignJWTInput struct {
	Payload map[string]any `json:"payload"`
	KID     string         `json:"kid,omitempty"`
	Header  map[string]any `json:"header,omitempty"`
}
type SignJWTResult struct {
	JWT string `json:"jwt"`
}

// ── Trust exchanges ──────────────────────────────────────────────
type TrustExchangeSubjectRules struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
}
type TrustExchangeClaimRule struct {
	Equals   string   `json:"equals,omitempty"`
	OneOf    []string `json:"one_of,omitempty"`
	Prefix   string   `json:"prefix,omitempty"`
	Required bool     `json:"required,omitempty"`
}
type TrustExchangeClaimMapping struct {
	SubjectClaim string            `json:"subject_claim,omitempty"`
	Scopes       []string          `json:"scopes,omitempty"`
	Claims       map[string]string `json:"claims,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}
type TrustExchangeInput struct {
	Name             string                            `json:"name"`
	Description      string                            `json:"description,omitempty"`
	Type             string                            `json:"type"`     // "oidc"
	Provider         string                            `json:"provider"` // github|google|auth0|custom_oidc
	Issuer           string                            `json:"issuer"`
	DiscoveryURL     string                            `json:"discovery_url,omitempty"`
	JWKSURL          string                            `json:"jwks_url,omitempty"`
	AllowedAudiences []string                          `json:"allowed_audiences"`
	SubjectRules     TrustExchangeSubjectRules         `json:"subject_rules"`
	ClaimRules       map[string]TrustExchangeClaimRule `json:"claim_rules"`
	ClaimMapping     TrustExchangeClaimMapping         `json:"claim_mapping"`
	OutputMode       string                            `json:"output_mode"` // claims|jwt
	OutputKeySpecID  string                            `json:"output_key_spec_id,omitempty"`
	Active           bool                              `json:"active"`
}
type TrustExchange struct {
	ID string `json:"id"`
	TrustExchangeInput
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── Trust providers ──────────────────────────────────────────────
type TrustProviderClaimPolicy struct {
	Constant map[string]any `json:"constant,omitempty"`
	Allowed  []string       `json:"allowed,omitempty"`
	Required []string       `json:"required,omitempty"`
}
type TrustProviderInput struct {
	Name              string                   `json:"name"`
	Description       string                   `json:"description,omitempty"`
	Type              string                   `json:"type"` // "oidc"
	SigningKeySetID   string                   `json:"signing_key_set_id"`
	IssuerHost        string                   `json:"issuer_host,omitempty"`
	AllowedAudiences  []string                 `json:"allowed_audiences"`
	DefaultAudience   string                   `json:"default_audience,omitempty"`
	ClaimPolicy       TrustProviderClaimPolicy `json:"claim_policy"`
	SubjectRequired   bool                     `json:"subject_required"`
	TTLDefaultSeconds int64                    `json:"ttl_default_seconds,omitempty"`
	TTLMaxSeconds     int64                    `json:"ttl_max_seconds,omitempty"`
	Active            bool                     `json:"active"`
}
type TrustProvider struct {
	ID string `json:"id"`
	TrustProviderInput
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type MintTokenInput struct {
	Subject    string         `json:"subject,omitempty"`
	Audience   string         `json:"audience,omitempty"`
	Claims     map[string]any `json:"claims,omitempty"`
	TTLSeconds int64          `json:"ttl_seconds,omitempty"`
}
type MintTokenResult struct {
	Valid     bool           `json:"valid"`
	Code      string         `json:"code,omitempty"`
	Token     string         `json:"token,omitempty"`
	Issuer    string         `json:"issuer,omitempty"`
	Subject   string         `json:"subject,omitempty"`
	Audience  string         `json:"audience,omitempty"`
	ExpiresIn int64          `json:"expires_in,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
}
```

NOTE: `AsymmetricPublicMaterial` needs json tags (`kty`,`alg`,`use`,`n`,`e`,`x`,`y`,`crv`) — copy them from the server dto. The embedded `TrustExchangeInput`/`TrustProviderInput` in the response structs is fine for decode since field json tags match; if the server response omits/renames any field vs the Input, define a separate flat response struct instead (verify against server dto and prefer an explicit struct if they diverge).

- [ ] **Step 2: Commit (after Step 1 of next sub-task compiles)**

This file is committed together with `akaas.go` in the next step — it won't compile alone if any type is referenced only there. Proceed to add `akaas.go`, then commit both.


- [ ] **Step 3: Create `internal/client/akaas.go`**

Every method is a thin `Do`/`Search` call. Paths must `url.PathEscape` user-supplied path segments. Search filter values use the nested `filter` shape from `search.go` (adjust if Task 4 Step 2 found the array shape).

```go
package client

import (
	"context"
	"net/url"
)

// permission-sets
func (c *Client) CreatePermissionSet(ctx context.Context, in PermissionSetInput) (*PermissionSet, error) {
	var out PermissionSet
	return &out, c.Do(ctx, "POST", "/api/permission-sets", in, &out)
}
func (c *Client) UpdatePermissionSet(ctx context.Context, id string, in PermissionSetInput) (*PermissionSet, error) {
	var out PermissionSet
	return &out, c.Do(ctx, "PATCH", "/api/permission-sets/"+url.PathEscape(id), in, &out)
}
func (c *Client) DeletePermissionSet(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/permission-sets/"+url.PathEscape(id), nil, nil)
}
func (c *Client) SearchPermissionSets(ctx context.Context, req SearchRequest) (*SearchResponse[PermissionSet], error) {
	return Search[PermissionSet](ctx, c, "/api/permission-sets", req)
}

// key-specs
func (c *Client) CreateKeySpec(ctx context.Context, in KeySpecInput) (*KeySpec, error) {
	var out KeySpec
	return &out, c.Do(ctx, "POST", "/api/key-specs", in, &out)
}
func (c *Client) UpdateKeySpec(ctx context.Context, id string, in KeySpecInput) (*KeySpec, error) {
	var out KeySpec
	return &out, c.Do(ctx, "PATCH", "/api/key-specs/"+url.PathEscape(id), in, &out)
}
func (c *Client) DeleteKeySpec(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/key-specs/"+url.PathEscape(id), nil, nil)
}
func (c *Client) SearchKeySpecs(ctx context.Context, req SearchRequest) (*SearchResponse[KeySpec], error) {
	return Search[KeySpec](ctx, c, "/api/key-specs", req)
}
func (c *Client) KeySpecEvents(ctx context.Context, id, subject string) ([]KeyEvent, error) {
	path := "/api/key-specs/" + url.PathEscape(id) + "/events"
	if subject != "" {
		path += "?subject=" + url.QueryEscape(subject)
	}
	var out []KeyEvent
	return out, c.Do(ctx, "GET", path, nil, &out)
}
func (c *Client) IssueKey(ctx context.Context, specID string, in IssueKeyInput) (*IssueKeyResult, error) {
	var out IssueKeyResult
	return &out, c.Do(ctx, "POST", "/api/key-specs/"+url.PathEscape(specID)+"/keys", in, &out)
}
func (c *Client) SearchSpecKeys(ctx context.Context, specID string, req SearchRequest) (*SearchResponse[IssuedKey], error) {
	return Search[IssuedKey](ctx, c, "/api/key-specs/"+url.PathEscape(specID)+"/keys", req)
}
func (c *Client) RevokeKeysBySubject(ctx context.Context, specID, subject string) (map[string]int, error) {
	var out map[string]int
	return out, c.Do(ctx, "POST", "/api/key-specs/"+url.PathEscape(specID)+"/keys/revoke-by-subject", RevokeBySubjectInput{Subject: subject}, &out)
}
func (c *Client) CreateWidgetSession(ctx context.Context, specID string, in WidgetSessionInput) (*WidgetSessionToken, error) {
	var out WidgetSessionToken
	return &out, c.Do(ctx, "POST", "/api/key-specs/"+url.PathEscape(specID)+"/widget-sessions", in, &out)
}

// keys
func (c *Client) SearchKeys(ctx context.Context, req SearchRequest) (*SearchResponse[IssuedKey], error) {
	return Search[IssuedKey](ctx, c, "/api/keys", req)
}
func (c *Client) GetKey(ctx context.Context, id string) (*IssuedKey, error) {
	var out IssuedKey
	return &out, c.Do(ctx, "GET", "/api/keys/"+url.PathEscape(id), nil, &out)
}
func (c *Client) UpdateKey(ctx context.Context, id string, in UpdateKeyInput) (*IssuedKey, error) {
	var out IssuedKey
	return &out, c.Do(ctx, "PATCH", "/api/keys/"+url.PathEscape(id), in, &out)
}
func (c *Client) RevokeKey(ctx context.Context, id string) (*IssuedKey, error) {
	var out IssuedKey
	return &out, c.Do(ctx, "POST", "/api/keys/"+url.PathEscape(id)+"/revoke", nil, &out)
}
func (c *Client) RotateKey(ctx context.Context, id string, in RotateKeyInput) (*IssueKeyResult, error) {
	var out IssueKeyResult
	return &out, c.Do(ctx, "POST", "/api/keys/"+url.PathEscape(id)+"/rotate", in, &out)
}
func (c *Client) KeyEvents(ctx context.Context, id string) ([]KeyEvent, error) {
	var out []KeyEvent
	return out, c.Do(ctx, "GET", "/api/keys/"+url.PathEscape(id)+"/events", nil, &out)
}
func (c *Client) VerifyKey(ctx context.Context, key string) (*VerifyKeyResult, error) {
	var out VerifyKeyResult
	return &out, c.Do(ctx, "POST", "/api/keys/verify", VerifyKeyInput{Key: key}, &out)
}

// signing-key-sets ({kind}/{set_name})
func sksPath(kind, name string, suffix string) string {
	return "/api/signing-key-sets/" + url.PathEscape(kind) + "/" + url.PathEscape(name) + suffix
}
func (c *Client) CreateSigningKeySet(ctx context.Context, in SigningKeySetInput) (*SigningKeySet, error) {
	var out SigningKeySet
	return &out, c.Do(ctx, "POST", "/api/signing-key-sets", in, &out)
}
func (c *Client) SearchSigningKeySets(ctx context.Context, req SearchRequest) (*SearchResponse[SigningKeySet], error) {
	return Search[SigningKeySet](ctx, c, "/api/signing-key-sets", req)
}
func (c *Client) GetSigningKeySet(ctx context.Context, kind, name string) (*SigningKeySetDetail, error) {
	var out SigningKeySetDetail
	return &out, c.Do(ctx, "GET", sksPath(kind, name, ""), nil, &out)
}
func (c *Client) UpdateSigningKeySet(ctx context.Context, kind, name, newName string) (*SigningKeySet, error) {
	var out SigningKeySet
	return &out, c.Do(ctx, "PATCH", sksPath(kind, name, ""), SigningKeySetUpdateInput{Name: newName}, &out)
}
func (c *Client) DeleteSigningKeySet(ctx context.Context, kind, name string) error {
	return c.Do(ctx, "DELETE", sksPath(kind, name, ""), nil, nil)
}
func (c *Client) GenerateSigningKey(ctx context.Context, kind, name string) (*SigningKey, error) {
	var out SigningKey
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/keys/generate"), nil, &out)
}
func (c *Client) ActivateSigningKey(ctx context.Context, kind, name, keyID string) (*SigningKey, error) {
	var out SigningKey
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/keys/"+url.PathEscape(keyID)+"/activate"), nil, &out)
}
func (c *Client) RevokeSigningKey(ctx context.Context, kind, name, keyID string) (*SigningKey, error) {
	var out SigningKey
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/keys/"+url.PathEscape(keyID)+"/revoke"), nil, &out)
}
func (c *Client) SigningKeySecret(ctx context.Context, kind, name, keyID string) (map[string]string, error) {
	var out map[string]string
	return out, c.Do(ctx, "GET", sksPath(kind, name, "/keys/"+url.PathEscape(keyID)+"/secret"), nil, &out)
}
func (c *Client) SignJWT(ctx context.Context, kind, name string, in SignJWTInput) (*SignJWTResult, error) {
	var out SignJWTResult
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/sign"), in, &out)
}
func (c *Client) SigningKeySetSecret(ctx context.Context, kind, name string) (map[string]string, error) {
	var out map[string]string
	return out, c.Do(ctx, "GET", sksPath(kind, name, "/secret"), nil, &out)
}
func (c *Client) RotateSigningKeySetSecret(ctx context.Context, kind, name string) (map[string]string, error) {
	var out map[string]string
	return out, c.Do(ctx, "POST", sksPath(kind, name, "/secret/rotate"), nil, &out)
}

// trust-exchanges
func (c *Client) CreateTrustExchange(ctx context.Context, in TrustExchangeInput) (*TrustExchange, error) {
	var out TrustExchange
	return &out, c.Do(ctx, "POST", "/api/trust-exchanges", in, &out)
}
func (c *Client) SearchTrustExchanges(ctx context.Context, req SearchRequest) (*SearchResponse[TrustExchange], error) {
	return Search[TrustExchange](ctx, c, "/api/trust-exchanges", req)
}
func (c *Client) GetTrustExchange(ctx context.Context, id string) (*TrustExchange, error) {
	var out TrustExchange
	return &out, c.Do(ctx, "GET", "/api/trust-exchanges/"+url.PathEscape(id), nil, &out)
}
func (c *Client) UpdateTrustExchange(ctx context.Context, id string, in TrustExchangeInput) (*TrustExchange, error) {
	var out TrustExchange
	return &out, c.Do(ctx, "PATCH", "/api/trust-exchanges/"+url.PathEscape(id), in, &out)
}
func (c *Client) DeleteTrustExchange(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/trust-exchanges/"+url.PathEscape(id), nil, nil)
}

// trust-providers
func (c *Client) CreateTrustProvider(ctx context.Context, in TrustProviderInput) (*TrustProvider, error) {
	var out TrustProvider
	return &out, c.Do(ctx, "POST", "/api/trust-providers", in, &out)
}
func (c *Client) SearchTrustProviders(ctx context.Context, req SearchRequest) (*SearchResponse[TrustProvider], error) {
	return Search[TrustProvider](ctx, c, "/api/trust-providers", req)
}
func (c *Client) GetTrustProvider(ctx context.Context, id string) (*TrustProvider, error) {
	var out TrustProvider
	return &out, c.Do(ctx, "GET", "/api/trust-providers/"+url.PathEscape(id), nil, &out)
}
func (c *Client) UpdateTrustProvider(ctx context.Context, id string, in TrustProviderInput) (*TrustProvider, error) {
	var out TrustProvider
	return &out, c.Do(ctx, "PATCH", "/api/trust-providers/"+url.PathEscape(id), in, &out)
}
func (c *Client) DeleteTrustProvider(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/trust-providers/"+url.PathEscape(id), nil, nil)
}

// auth-plane mint (call on AuthClient)
func (c *Client) MintTrustProviderToken(ctx context.Context, providerID, apiKey string, in MintTokenInput) (*MintTokenResult, error) {
	prev := c.token
	c.token = apiKey
	defer func() { c.token = prev }()
	var out MintTokenResult
	return &out, c.Do(ctx, "POST", "/trust-providers/"+url.PathEscape(providerID)+"/token", in, &out)
}
```

- [ ] **Step 4: Verify build + commit**

```bash
go build ./internal/...
go test ./internal/... -count=1
git add internal/client
git commit -m "feat: add typed AKaaS client methods"
```
Expected: `internal/` builds; client + config + output tests pass. (`cmd/` and `main.go` still won't build until Task 6.)

---

### Task 6: Mandatory commands + resource skeletons (build goes green)

This task makes `main.go` compile and run. It adds the mandatory commands and lightweight skeletons for the six resource groups (parent + subcommand structs with stub `Run` methods). Tasks 7–11 fill in the stub bodies and flags — they EDIT these structs, they do not redefine them.

**Files:**
- Create: `cmd/errors.go`, `cmd/login.go`, `cmd/logout.go`, `cmd/whoami.go`, `cmd/completion.go`
- Create skeletons: `cmd/permission_sets.go`, `cmd/key_specs.go`, `cmd/keys.go`, `cmd/signing_key_sets.go`, `cmd/trust_exchanges.go`, `cmd/trust_providers.go`

- [ ] **Step 1: `cmd/errors.go`** — styled error rendering (D9)

```go
package cmd

import "github.com/microwave-sh/microwave-cli/internal/output"

// RenderError wraps an error in a styled banner for TTY stderr display.
func RenderError(err error) string {
	return output.ErrorBanner("Error: " + err.Error())
}

// errNotImplemented is a placeholder used by resource skeletons until Tasks 7–11.
type notImplementedError struct{ what string }

func (e notImplementedError) Error() string { return e.what + ": not implemented yet" }
```

- [ ] **Step 2: `cmd/login.go`** — store a pasted key (no device flow; server issues no CLI tokens)

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

type LoginCmd struct {
	Key string `arg:"" optional:"" help:"Management API key. If omitted, prompts (and prints the console URL)."`
}

func (c *LoginCmd) Run(g *Globals) error {
	key := strings.TrimSpace(c.Key)
	if key == "" {
		fmt.Printf("Create a management API key at %s\n", output.Bold.Render("https://app.microwave.sh"))
		fmt.Print("Paste your key: ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		key = strings.TrimSpace(line)
	}
	if key == "" {
		return fmt.Errorf("no key provided")
	}
	if err := config.WriteGlobalAuth(key); err != nil {
		return err
	}
	fmt.Printf("%s Saved to %s\n",
		output.Green.Render("✓"),
		output.Dim.Render(filepath.Join(config.GlobalConfigDir(), "config.toml")))
	return nil
}
```

- [ ] **Step 3: `cmd/logout.go`**

```go
package cmd

import (
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/config"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

type LogoutCmd struct{}

func (c *LogoutCmd) Run(g *Globals) error {
	if err := config.ClearAuth(); err != nil {
		return err
	}
	fmt.Printf("%s Logged out.\n", output.Green.Render("✓"))
	return nil
}
```

- [ ] **Step 4: `cmd/whoami.go`** — verify the stored key for real identity (no `/me` endpoint exists)

```go
package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/output"
)

type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(g *Globals) error {
	token, err := g.resolveToken()
	if err != nil {
		return err
	}
	res, err := g.Client().VerifyKey(context.Background(), token)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(res)
	}
	if !res.Valid {
		return fmt.Errorf("stored key is not valid (%s)", res.Code)
	}
	output.PrintTable(
		[]string{"Subject", "Key ID", "Scopes"},
		[][]string{{res.Subject, res.KeyID, fmt.Sprintf("%v", res.Scopes)}},
		false,
	)
	return nil
}
```

(If `VerifyKey` returns 403 because the management key lacks `keys:verify`, fall back to printing the masked key: last 4 chars. Implement the fallback in this method.)

- [ ] **Step 5: `cmd/completion.go`** — bash/zsh/fish (D16)

```go
package cmd

import "fmt"

type CompletionCmd struct {
	Shell string `arg:"" enum:"bash,zsh,fish" help:"Shell type (bash, zsh, fish)."`
}

func (c *CompletionCmd) Run(g *Globals) error {
	switch c.Shell {
	case "bash":
		fmt.Println("# Add to ~/.bashrc:\n# source <(microwave completion bash)")
		fmt.Println(`complete -o default -C "microwave" microwave`)
	case "zsh":
		fmt.Println("# Add to ~/.zshrc:\n# source <(microwave completion zsh)")
		fmt.Println(`compdef _microwave microwave`)
	case "fish":
		fmt.Println("# microwave completion fish | source")
		fmt.Println(`complete -c microwave -f`)
	}
	return nil
}
```

(Kong supports native completion via `kong.Plugins`/`complete`; if the repo wires `github.com/posener/complete` or Kong's built-in, prefer emitting the real script. The above is the minimum acceptable per D16 — printing a script + the source line. Upgrade to Kong-generated completion if straightforward.)

- [ ] **Step 6: Resource skeletons** — create six files, each with the parent + subcommand structs from the target tree, every `Run` returning `notImplementedError{"<cmd>"}`. Example `cmd/permission_sets.go`:

```go
package cmd

type PermissionSetsCmd struct {
	List   permissionSetsListCmd   `cmd:"" help:"Search permission sets."`
	Create permissionSetsCreateCmd `cmd:"" help:"Create a permission set."`
	Update permissionSetsUpdateCmd `cmd:"" help:"Update a permission set."`
	Delete permissionSetsDeleteCmd `cmd:"" help:"Delete a permission set."`
}

type permissionSetsListCmd struct{}
func (c *permissionSetsListCmd) Run(g *Globals) error { return notImplementedError{"permission-sets list"} }
type permissionSetsCreateCmd struct{}
func (c *permissionSetsCreateCmd) Run(g *Globals) error { return notImplementedError{"permission-sets create"} }
type permissionSetsUpdateCmd struct {
	ID string `arg:"" help:"Permission set ID."`
}
func (c *permissionSetsUpdateCmd) Run(g *Globals) error { return notImplementedError{"permission-sets update"} }
type permissionSetsDeleteCmd struct {
	ID string `arg:"" help:"Permission set ID."`
}
func (c *permissionSetsDeleteCmd) Run(g *Globals) error { return notImplementedError{"permission-sets delete"} }
```

Create the analogous skeleton files for `KeySpecsCmd`, `KeysCmd`, `SigningKeySetsCmd`, `TrustExchangesCmd`, `TrustProvidersCmd` with the subcommands shown in the target command tree (matching names so Tasks 7–11 only fill bodies + add flag fields). Nested groups (`key-specs keys`, `signing-key-sets keys`) are nested structs.

- [ ] **Step 7: Build, run, verify mandatory commands work**

```bash
go build -o microwave .
./microwave --help            # shows all command groups, no workspace/spec/sdk/docs
./microwave version           # prints "dev"
./microwave completion zsh     # prints script
MICROWAVE_TOKEN=x ./microwave permission-sets list  # prints "permission-sets list: not implemented yet"
go vet ./...
```
Expected: builds; help shows the AKaaS tree; stubs error cleanly.

- [ ] **Step 8: Commit**

```bash
git add cmd main.go
git commit -m "feat: add mandatory commands and resource skeletons"
```

---

### Task 7: Command helpers + trust-providers (the exemplar)

This task adds shared command helpers and the FULL `trust-providers` implementation. Tasks 8–12 follow this exact pattern for the other resources.

**Files:**
- Create: `cmd/helpers.go`
- Replace skeleton: `cmd/trust_providers.go`
- Create: `cmd/trust_providers_test.go`

- [ ] **Step 1: `cmd/helpers.go`** — shared flag parsing + rendering

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

// listFlags are embedded by every resource `list` subcommand.
type listFlags struct {
	Limit  int    `help:"Max results." default:"25"`
	Cursor string `help:"Pagination cursor."`
	Sort   string `help:"Sort field (e.g. created_at)."`
	Desc   bool   `help:"Sort descending." default:"true"`
}

func (f listFlags) searchRequest(filter map[string]map[string]any) client.SearchRequest {
	req := client.SearchRequest{Limit: f.Limit, Cursor: f.Cursor, Filter: filter}
	if f.Sort != "" {
		dir := "asc"
		if f.Desc {
			dir = "desc"
		}
		req.Sort = []client.SortDirective{{Field: f.Sort, Direction: dir}}
	}
	return req
}

func parseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseJSONMap(s string) (map[string]any, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return m, nil
}
```

- [ ] **Step 2: Replace `cmd/trust_providers.go` with the full implementation**

```go
package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/client"
	"github.com/microwave-sh/microwave-cli/internal/output"
)

type TrustProvidersCmd struct {
	List      tpListCmd      `cmd:"" help:"Search trust providers."`
	Get       tpGetCmd       `cmd:"" help:"Get a trust provider."`
	Create    tpCreateCmd    `cmd:"" help:"Create a trust provider."`
	Update    tpUpdateCmd    `cmd:"" help:"Update a trust provider."`
	Delete    tpDeleteCmd    `cmd:"" help:"Delete a trust provider."`
	Mint      tpMintCmd      `cmd:"" help:"Mint a token from a trust provider."`
	Discovery tpDiscoveryCmd `cmd:"" help:"Print federation (discovery/JWKS/token) URLs."`
}

type tpListCmd struct{ listFlags }

func (c *tpListCmd) Run(g *Globals) error {
	page, err := g.Client().SearchTrustProviders(context.Background(), c.searchRequest(nil))
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(page)
	}
	rows := make([][]string, len(page.Data))
	for i, p := range page.Data {
		rows[i] = []string{p.ID, p.Name, p.SigningKeySetID, fmt.Sprintf("%v", p.Active)}
	}
	output.PrintTable([]string{"ID", "Name", "Signing Key Set", "Active"}, rows, false)
	return nil
}

type tpGetCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *tpGetCmd) Run(g *Globals) error {
	p, err := g.Client().GetTrustProvider(context.Background(), c.ID)
	if err != nil {
		return err
	}
	return output.PrintJSON(p)
}

type tpCreateCmd struct {
	Name             string `help:"Name." required:""`
	Description      string `help:"Description."`
	SigningKeySetID  string `name:"signing-key-set-id" help:"Asymmetric signing key set ID." required:""`
	IssuerHost       string `name:"issuer-host" help:"Custom issuer host."`
	AllowedAudiences string `name:"allowed-audiences" help:"Comma-separated allowed audiences." required:""`
	DefaultAudience  string `name:"default-audience" help:"Default audience."`
	AllowedClaims    string `name:"allowed-claims" help:"Comma-separated allowed claim keys."`
	RequiredClaims   string `name:"required-claims" help:"Comma-separated required claim keys."`
	ConstantClaims   string `name:"constant-claims" help:"Constant claims as JSON object."`
	SubjectRequired  bool   `name:"subject-required" help:"Require a subject." default:"true"`
	TTLDefault       int64  `name:"ttl-default-seconds" help:"Default token TTL." default:"3600"`
	TTLMax           int64  `name:"ttl-max-seconds" help:"Max token TTL." default:"3600"`
}

func (c *tpCreateCmd) Run(g *Globals) error {
	constant, err := parseJSONMap(c.ConstantClaims)
	if err != nil {
		return err
	}
	in := client.TrustProviderInput{
		Name:             c.Name,
		Description:      c.Description,
		Type:             "oidc",
		SigningKeySetID:  c.SigningKeySetID,
		IssuerHost:       c.IssuerHost,
		AllowedAudiences: parseCSV(c.AllowedAudiences),
		DefaultAudience:  c.DefaultAudience,
		ClaimPolicy: client.TrustProviderClaimPolicy{
			Allowed:  parseCSV(c.AllowedClaims),
			Required: parseCSV(c.RequiredClaims),
			Constant: constant,
		},
		SubjectRequired:   c.SubjectRequired,
		TTLDefaultSeconds: c.TTLDefault,
		TTLMaxSeconds:     c.TTLMax,
		Active:            true,
	}
	p, err := g.Client().CreateTrustProvider(context.Background(), in)
	if err != nil {
		return err
	}
	return output.PrintJSON(p)
}

type tpUpdateCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
	tpCreateCmd
}

func (c *tpUpdateCmd) Run(g *Globals) error {
	constant, err := parseJSONMap(c.ConstantClaims)
	if err != nil {
		return err
	}
	in := client.TrustProviderInput{
		Name: c.Name, Description: c.Description, Type: "oidc",
		SigningKeySetID: c.SigningKeySetID, IssuerHost: c.IssuerHost,
		AllowedAudiences: parseCSV(c.AllowedAudiences), DefaultAudience: c.DefaultAudience,
		ClaimPolicy: client.TrustProviderClaimPolicy{
			Allowed: parseCSV(c.AllowedClaims), Required: parseCSV(c.RequiredClaims), Constant: constant,
		},
		SubjectRequired: c.SubjectRequired, TTLDefaultSeconds: c.TTLDefault, TTLMaxSeconds: c.TTLMax, Active: true,
	}
	p, err := g.Client().UpdateTrustProvider(context.Background(), c.ID, in)
	if err != nil {
		return err
	}
	return output.PrintJSON(p)
}

type tpDeleteCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *tpDeleteCmd) Run(g *Globals) error {
	if err := g.Client().DeleteTrustProvider(context.Background(), c.ID); err != nil {
		return err
	}
	fmt.Printf("%s Deleted %s\n", output.Green.Render("✓"), c.ID)
	return nil
}

type tpMintCmd struct {
	ID       string `arg:"" help:"Trust provider ID."`
	APIKey   string `name:"api-key" help:"Management API key to authenticate the mint (defaults to stored token)."`
	Subject  string `help:"Token subject."`
	Audience string `help:"Token audience."`
	Claims   string `help:"Claims as JSON object."`
	TTL      int64  `name:"ttl" help:"Token TTL seconds."`
}

func (c *tpMintCmd) Run(g *Globals) error {
	claims, err := parseJSONMap(c.Claims)
	if err != nil {
		return err
	}
	key := c.APIKey
	if key == "" {
		key, err = g.resolveToken()
		if err != nil {
			return err
		}
	}
	res, err := g.AuthClient().MintTrustProviderToken(context.Background(), c.ID, key, client.MintTokenInput{
		Subject: c.Subject, Audience: c.Audience, Claims: claims, TTLSeconds: c.TTL,
	})
	if err != nil {
		return err
	}
	return output.PrintJSON(res)
}

type tpDiscoveryCmd struct {
	ID string `arg:"" help:"Trust provider ID."`
}

func (c *tpDiscoveryCmd) Run(g *Globals) error {
	base := g.authURL() + "/trust-providers/" + c.ID
	rows := [][]string{
		{"Issuer", base},
		{"Discovery", base + "/.well-known/openid-configuration"},
		{"JWKS", base + "/.well-known/jwks.json"},
		{"Token", base + "/token"},
	}
	output.PrintTable([]string{"Endpoint", "URL"}, rows, g.IsJSON())
	return nil
}
```

- [ ] **Step 3: `cmd/trust_providers_test.go`** — command test (no server)

```go
package cmd

import (
	"strings"
	"testing"

	"github.com/microwave-sh/microwave-cli/internal/client"
)

func TestTPCreate_BuildsOIDCInput(t *testing.T) {
	c := tpCreateCmd{
		Name: "deploy", SigningKeySetID: "sks_1",
		AllowedAudiences: "api://prod, api://staging", DefaultAudience: "api://prod",
		AllowedClaims: "role", RequiredClaims: "role", ConstantClaims: `{"tenant":"acme"}`,
		SubjectRequired: true, TTLDefault: 900, TTLMax: 3600,
	}
	constant, _ := parseJSONMap(c.ConstantClaims)
	in := client.TrustProviderInput{
		Name: c.Name, Type: "oidc", SigningKeySetID: c.SigningKeySetID,
		AllowedAudiences: parseCSV(c.AllowedAudiences),
		ClaimPolicy:      client.TrustProviderClaimPolicy{Allowed: parseCSV(c.AllowedClaims), Required: parseCSV(c.RequiredClaims), Constant: constant},
	}
	if in.Type != "oidc" || len(in.AllowedAudiences) != 2 || in.ClaimPolicy.Constant["tenant"] != "acme" {
		t.Fatalf("input = %+v", in)
	}
}

func TestParseCSV_TrimsAndDropsEmpty(t *testing.T) {
	got := parseCSV(" a , ,b ")
	if strings.Join(got, ",") != "a,b" {
		t.Fatalf("parseCSV = %v", got)
	}
}
```

- [ ] **Step 4: Build, test, commit**

```bash
go build -o microwave . && go test ./... -count=1
git add cmd
git commit -m "feat: implement trust-providers commands + shared helpers"
```
Expected: builds; tests pass; `./microwave trust-providers --help` shows list/get/create/update/delete/mint/discovery.

---

### Tasks 8–12: remaining resources (follow the Task 7 exemplar exactly)

Each task REPLACES the skeleton file with a full implementation mirroring Task 7's structure: `list` embeds `listFlags` + builds a `SearchRequest` + renders a table (or `--output json`); `get` prints JSON; `create`/`update` map flags → the `client.*Input` struct (CSV via `parseCSV`, JSON via `parseJSONMap`) and print JSON; `delete` prints a ✓ line; action subcommands call the matching client method. Each task: implement → `go build -o microwave . && go test ./... -count=1` → commit. Concrete spec per task:

- [ ] **Task 8 — `cmd/permission_sets.go`** (`SearchPermissionSets`/`CreatePermissionSet`/`UpdatePermissionSet`/`DeletePermissionSet`).
  - `list` columns: `ID, Name, #Permissions` (`len(p.Permissions)`).
  - `create`/`update <id>` flags: `--name` (required), `--description`, `--permission` (repeatable `[]string`, format `name:label[:dangerous]`, parsed into `[]client.Permission`). Write a `parsePermissions([]string) ([]client.Permission, error)` helper in this file.
  - `delete <id>`.
  - Commit: `feat: implement permission-sets commands`.

- [ ] **Task 9 — `cmd/key_specs.go`** (`SearchKeySpecs`/`CreateKeySpec`/`UpdateKeySpec`/`DeleteKeySpec`/`KeySpecEvents`/`IssueKey`/`SearchSpecKeys`/`RevokeKeysBySubject`/`CreateWidgetSession`).
  - `list` columns: `ID, Name, Format, Created` (`output.FormatTimeAgo`). Filter flag `--format opaque|jwt` → `filter{"format":{"eq":...}}`.
  - `create`/`update <id>` flags: `--name` (req), `--description`, `--format` (enum opaque|jwt, req), `--permission-set-id`, `--signing-key-set-id`, `--jwt-algorithm`, `--jwt-issuer`, `--jwt-audience`, `--default-ttl`, `--max-ttl`, `--allow-never`, `--rotation-reminder-days`, `--standard-claims` (CSV), `--allow-custom-expiry/scopes/claims`, `--webhook-endpoint`, `--webhook-events` (CSV), `--webhook-signing-key-set-id`, `--opaque-prefix`, `--opaque-lookup-response`. Map to `client.KeySpecInput`.
  - `delete <id>`, `events <id> --subject`.
  - Nested `keys` group: `key-specs keys issue <spec-id>` (flags `--subject` req, `--name` req, `--scopes` CSV, `--claims` JSON, `--metadata` JSON, `--expires-in`); `key-specs keys list <spec-id>` (search, columns `ID, Subject, Status, Created`); `key-specs keys revoke-by-subject <spec-id> --subject` (prints `{"revoked":N}`).
  - `widget-session <spec-id>` flags `--subject` req, `--scopes` CSV, `--claims` JSON, `--redirect-url`, `--ttl`.
  - Commit: `feat: implement key-specs commands`.

- [ ] **Task 10 — `cmd/keys.go`** (`SearchKeys`/`GetKey`/`UpdateKey`/`RevokeKey`/`RotateKey`/`KeyEvents`/`VerifyKey`).
  - `list` columns: `ID, Spec, Subject, Status, Created`. Filter flags `--spec-id`, `--subject`, `--status active|revoked|expired|rotating`.
  - `get <id>` prints JSON.
  - `update <id>` flags: `--name` (string pointer — only send if set; use `*string`), `--scopes` CSV, `--expires-at` (RFC3339 → `*time.Time`), `--metadata` JSON.
  - `revoke <id>` (✓ line + prints status), `rotate <id> --overlap-seconds` (prints new `IssueKeyResult` incl. `key`), `events <id>` (columns `Type, Subject, Actor, When`), `verify <key>` (prints `VerifyKeyResult`; non-zero-ish: if `!valid`, still print + return nil).
  - Commit: `feat: implement keys commands`.

- [ ] **Task 11 — `cmd/signing_key_sets.go`** (`SearchSigningKeySets`/`GetSigningKeySet`/`CreateSigningKeySet`/`UpdateSigningKeySet`/`DeleteSigningKeySet`/`SignJWT`/`SigningKeySetSecret`/`RotateSigningKeySetSecret` + keys `GenerateSigningKey`/`ActivateSigningKey`/`RevokeSigningKey`/`SigningKeySecret`).
  - `list` columns: `ID, Name, Kind, Algorithm`.
  - `get <kind> <name>` prints `SigningKeySetDetail` JSON.
  - `create` flags: `--kind asymmetric|symmetric` (req), `--name` (req), `--algorithm` (req).
  - `update <kind> <name> --name <new>`; `delete <kind> <name>`.
  - `sign <kind> <name>` flags `--payload` JSON (req), `--kid`, `--header` JSON → `SignJWTResult`.
  - `secret <kind> <name>` (GET set secret), `rotate-secret <kind> <name>`.
  - Nested `keys`: `generate <kind> <name>`, `activate <kind> <name> <key-id>`, `revoke <kind> <name> <key-id>`, `secret <kind> <name> <key-id>`.
  - Commit: `feat: implement signing-key-sets commands`.

- [ ] **Task 12 — `cmd/trust_exchanges.go`** (`SearchTrustExchanges`/`GetTrustExchange`/`CreateTrustExchange`/`UpdateTrustExchange`/`DeleteTrustExchange`).
  - `list` columns: `ID, Name, Provider, Output, Active`. Filters `--provider`, `--output-mode`, `--active`.
  - `create`/`update <id>` flags: `--name` (req), `--description`, `--provider github|google|auth0|custom_oidc` (req), `--issuer` (req), `--discovery-url`, `--jwks-url`, `--allowed-audiences` CSV (req), `--subject-exact`, `--subject-prefix`, `--output-mode claims|jwt` (req), `--output-key-spec-id`, and claim-rule/mapping as JSON flags `--claim-rules` (JSON → `map[string]TrustExchangeClaimRule`), `--claim-mapping` (JSON → `TrustExchangeClaimMapping`). Set `Type:"oidc"`, `Active:true`.
  - `get <id>`, `delete <id>`.
  - Commit: `feat: implement trust-exchanges commands`.

Each of Tasks 8–12 includes at least one command test (parse/flag-mapping) in a `_test.go` file, mirroring `TestTPCreate_BuildsOIDCInput`.

---

### Task 13: Distribution pipeline (mirror the working Sandbar CLI)

**Reference (copy + adapt, do not invent):** `~/mataki/sandbar/sandbar-cli` is a production Mataki CLI with a working release pipeline. Copy its `.goreleaser.yaml`, `.releaserc.json`, `Makefile`, `.github/workflows/release.yml`, `.github/workflows/deploy.yml`, and `~/mataki/sandbar/sandbar-web/public/install.sh`, substituting `sandbar`→`microwave`, `sandbar-cloud`→`microwave-sh`, `sandbar.cloud`→`microwave.sh`, tap `homebrew-tap`→`homebrew-microwave`. Sandbar's choices that the CLI.md generic templates miss and that you MUST keep: GoReleaser `release.mode: append` (semantic-release cuts the GitHub Release; GoReleaser appends archives), and a `mataki-robot` GitHub App token for the cross-repo Homebrew push.

**Files:**
- Create in `microwave-cli`: `.goreleaser.yaml`, `.releaserc.json`, `Makefile`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.github/workflows/deploy.yml`
- Create in `microwave-web` (separate repo, `~/mataki/microwave/microwave-web`): `public/install.sh`
- The tap `microwave-sh/homebrew-microwave` (formula at `Formula/microwave.rb`) is regenerated by GoReleaser — do not hand-edit it.

- [ ] **Step 1: `.goreleaser.yaml`** (Sandbar-adapted — note `directory: Formula` because microwave's tap keeps formulae under `Formula/`, and `mode: append`)

```yaml
version: 2

builds:
  - main: .
    binary: microwave
    ldflags:
      - -s -w -X github.com/microwave-sh/microwave-cli/internal/version.Version={{.Version}}
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "microwave_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

brews:
  - repository:
      owner: microwave-sh
      name: homebrew-microwave
    directory: Formula
    name: microwave
    homepage: "https://microwave.sh"
    description: "Command-line interface for the Microwave API"
    install: |
      bin.install "microwave"

release:
  github:
    owner: microwave-sh
    name: microwave-cli
  mode: append
```

- [ ] **Step 2: `.releaserc.json`** (verbatim from Sandbar)

```json
{
  "branches": ["main"],
  "plugins": [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    "@semantic-release/github"
  ]
}
```

- [ ] **Step 3: `Makefile`** — follow CLI.md's [Makefile](#) example with `BINARY := microwave` and `build` ldflags `-s -w -X github.com/microwave-sh/microwave-cli/internal/version.Version=$(VERSION)`.

- [ ] **Step 4: `.github/workflows/release.yml`** (copy Sandbar's verbatim, adapting names) — on push to `main`, mint a `mataki-robot` App token, run `cycjimmy/semantic-release-action@v6` with `extra_plugins: @semantic-release/exec@6` + `@semantic-release/git@10`, `GITHUB_TOKEN` = app token, `concurrency: release`.

- [ ] **Step 5: `.github/workflows/deploy.yml`** (copy Sandbar's verbatim, adapting names) — on `release: published`, mint a `mataki-robot` App token scoped to `owner: microwave-sh` `repositories: microwave-cli` + `homebrew-microwave` (cross-repo write for the formula push), checkout `fetch-depth: 0`, `setup-go` from `go.mod`, `goreleaser/goreleaser-action@v7` `version: "~> v2"` `args: release --clean`, `GITHUB_TOKEN` = app token, `concurrency: deploy` `cancel-in-progress: false`.

- [ ] **Step 6: `.github/workflows/ci.yml`** — follow CLI.md's [ci.yml](#): on PR to main, `go vet` + `go test -race -count=1 ./...` + `golangci-lint`. (Sandbar relies on pre-commit; add CI here for PR safety.)

- [ ] **Step 7: `microwave-web/public/install.sh`** — copy `~/mataki/sandbar/sandbar-web/public/install.sh`, substituting `sandbar`→`microwave`, repo `sandbar-cloud/sandbar-cli`→`microwave-sh/microwave-cli`, env `SANDBAR_VERSION`→`MICROWAVE_VERSION`, `SANDBAR_INSTALL_DIR`→`MICROWAVE_INSTALL_DIR`, install message domain. The archive URL pattern `microwave_${VERSION}_${OS}_${ARCH}.tar.gz` must match the `.goreleaser.yaml` `name_template`. Commit this in the `microwave-web` repo (its own worktree/branch), not `microwave-cli`. Hosted at `https://microwave.sh/install.sh`.

- [ ] **Step 8: Verify GoReleaser builds locally**

```bash
cd /Users/sethyates/mataki/microwave/microwave-cli   # or its worktree
go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean
ls dist/   # 4 archives (darwin/linux × amd64/arm64) + checksums.txt
```
Expected: 4 archives + checksums.

- [ ] **Step 9: Commit (CLI repo)**

```bash
git add .goreleaser.yaml .releaserc.json Makefile .github
git commit -m "build: add goreleaser, semantic-release, and CI/deploy workflows"
```

**Prereqs to flag in the final report (infra, not code):** the `mataki-robot` GitHub App must be installed on both `microwave-sh/microwave-cli` and `microwave-sh/homebrew-microwave` with `contents: write`, and `GH_APP_ID` + `GH_PRIVATE_KEY` available to the CLI repo (org secrets) — same setup Sandbar already uses.

---

### Task 14: README + end-to-end validation

**Files:**
- Modify: `README.md`
- Modify: only if validation reveals defects from earlier tasks.

- [ ] **Step 1: Rewrite `README.md`** per CLI.md D22 structure: Install (Homebrew `brew install microwave-sh/microwave/microwave` + `curl -sSL https://microwave.sh/install.sh | sh`) → Quickstart (`microwave login`, `microwave key-specs list`, `microwave keys list`) → Commands (table of every command from the target tree) → Configuration (env vars table: `MICROWAVE_TOKEN`, `MICROWAVE_API_URL`, `MICROWAVE_AUTH_URL`, `MICROWAVE_NO_UPDATE_CHECK`, `XDG_CONFIG_HOME`, `NO_COLOR`) → CI/CD (`MICROWAVE_TOKEN` secret usage). Do NOT reference workspace/spec/sdk/docs/collections.

- [ ] **Step 2: Full build, vet, lint, test**

```bash
go build -o microwave .
go vet ./...
go test -race -count=1 ./...
golangci-lint run ./...  # if available
```
Expected: all green.

- [ ] **Step 3: Help-surface smoke (no server)**

```bash
./microwave --help
for g in permission-sets key-specs keys signing-key-sets trust-exchanges trust-providers; do ./microwave $g --help; done
./microwave version && ./microwave completion zsh
```
Expected: every group shows its real subcommands; no workspace/spec/sdk/docs/collections anywhere.

- [ ] **Step 4: Live smoke (if a key + reachable API are available)**

```bash
export MICROWAVE_TOKEN=<management key>
export MICROWAVE_API_URL=<api base>   # if not the default
./microwave whoami
./microwave key-specs list
./microwave keys list -o json | head
```
Expected: `whoami` shows subject/scopes; list commands return data or empty tables. If no key/API is reachable in this environment, record this as not-run and provide these manual steps in the final report.

- [ ] **Step 5: Final commit**

```bash
git add README.md
git commit -m "docs: rewrite README for the AKaaS command set"
git status --short
```

---

## Self-Review

- **Spec coverage:** Every server route in the "command map" table maps to a CLI command (Tasks 7–12); mandatory commands + completion (Task 6); `--output`/`--debug`/update-check/signal-context (Tasks 1,3,6); charm output stack (Task 2); full distribution (Task 13); README D22 (Task 14). Removed surfaces (workspace/spec/sdk/docs/collections/console) are deleted in Task 1.
- **Auth alignment:** CLI uses `Authorization: Bearer <management key>` + `API-Version: 2026-05-27` — matches the server's management-key path. `whoami` verifies the stored key (no `/me` endpoint exists). `login` stores a pasted key (no device/OIDC issuance server-side).
- **Placeholder scan:** Foundation tasks (1–7) carry complete code; Tasks 8–12 give concrete flag/column/client-method specs against the Task 7 exemplar (full code pattern) — implementable deterministically. Distribution (Task 13) copies the working `sandbar-cli` pipeline (GoReleaser `mode: append`, `mataki-robot` App token cross-repo push, `homebrew-microwave` tap with `Formula/` dir, install.sh in `microwave-web/public/`). One ASSUMPTION flagged: API hostname (`.sh` vs `.dev`).
- **Type consistency:** Client method names in `akaas.go` (Task 5) match the calls in command tasks (e.g. `SearchTrustProviders`, `MintTrustProviderToken`, `IssueKey`, `SigningKeySetSecret`). DTO field names mirror the server dto (verify-against-server instruction included where the exploration was less certain: search envelope shape, `AsymmetricPublicMaterial` tags, embedded-vs-flat response structs).
- **Open verification items for the implementer (resolve by reading the server, not guessing):** (1) `platform/search` request shape — nested `filter` object vs `filters` array; (2) exact response-struct field parity for trust-exchange/provider (embedded Input vs flat); (3) brand color hex from `microwave-app`.

## Execution Handoff

Plan saved to `microwave-cli/docs/superpowers/plans/2026-05-29-cli-akaas-rewrite.md`.

**1. Subagent-Driven (recommended)** — fresh subagent per task, two-stage review between tasks.
**2. Inline Execution** — execute here with checkpoints.

Tasks 1 and 6 are larger (scaffold + skeletons); Tasks 8–12 are mechanical given the Task 7 exemplar. Suggest a worktree in `microwave-cli` before starting.
