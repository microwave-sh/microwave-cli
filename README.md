# Microwave CLI

Microwave CLI — manage API keys, JWKS signing, and OIDC federation.

## Install

**Homebrew (recommended):**

```sh
brew install microwave-sh/microwave/microwave
```

The tap is `microwave-sh/homebrew-microwave`.

**curl installer:**

```sh
curl -sSL https://microwave.sh/install.sh | sh
```

**Direct download:**

Download a pre-built binary for your platform from
[GitHub Releases](https://github.com/microwave-sh/microwave-cli/releases).

## Quickstart

```sh
# 1. Authenticate — grab a management key from https://app.microwave.sh
microwave login

# 2. List your key specs
microwave key-specs list

# 3. List issued keys
microwave keys list

# 4. Confirm identity
microwave whoami
```

## Commands

### Top-level

| Command | Description |
|---|---|
| `microwave login [<key>]` | Store a management API key |
| `microwave logout` | Clear stored credentials |
| `microwave whoami` | Print the authenticated identity |
| `microwave version` | Print version |
| `microwave completion <shell>` | Print shell completion script (bash/zsh/fish/powershell) |

### permission-sets

| Command | Description |
|---|---|
| `permission-sets list` | Search permission sets |
| `permission-sets create --name=STRING` | Create a permission set |
| `permission-sets update --name=STRING <id>` | Update a permission set |
| `permission-sets delete <id>` | Delete a permission set |

### key-specs

| Command | Description |
|---|---|
| `key-specs list` | Search key specs |
| `key-specs create --name=STRING --format=STRING` | Create a key spec |
| `key-specs update --name=STRING --format=STRING <id>` | Update a key spec |
| `key-specs delete <id>` | Delete a key spec |
| `key-specs events <id>` | List key-spec events |
| `key-specs widget-session --subject=STRING <id>` | Create a widget session token |
| `key-specs keys issue --subject=STRING --name=STRING <spec-id>` | Issue a key from a spec |
| `key-specs keys list <spec-id>` | Search keys for a spec |
| `key-specs keys revoke-by-subject --subject=STRING <spec-id>` | Revoke all keys for a subject |

### keys

| Command | Description |
|---|---|
| `keys list` | Search issued keys |
| `keys get <id>` | Get an issued key |
| `keys update <id>` | Update a key |
| `keys revoke <id>` | Revoke a key |
| `keys rotate <id>` | Rotate a key |
| `keys events <id>` | List key events |
| `keys verify <key>` | Verify a key |

### signing-key-sets

| Command | Description |
|---|---|
| `signing-key-sets list` | Search signing key sets |
| `signing-key-sets get <kind> <name>` | Get a signing key set |
| `signing-key-sets create --kind=STRING --name=STRING --algorithm=STRING` | Create a signing key set |
| `signing-key-sets update --name=STRING <kind> <name>` | Rename a signing key set |
| `signing-key-sets delete <kind> <name>` | Delete a signing key set |
| `signing-key-sets sign --payload=STRING <kind> <name>` | Sign a JWT payload (asymmetric) |
| `signing-key-sets secret <kind> <name>` | Reveal symmetric secret state |
| `signing-key-sets rotate-secret <kind> <name>` | Rotate symmetric secret |
| `signing-key-sets keys generate <kind> <name>` | Generate a new signing key |
| `signing-key-sets keys activate <kind> <name> <key-id>` | Activate a signing key (symmetric) |
| `signing-key-sets keys revoke <kind> <name> <key-id>` | Revoke a signing key |
| `signing-key-sets keys secret <kind> <name> <key-id>` | Reveal a key secret (symmetric) |

### trust-exchanges

| Command | Description |
|---|---|
| `trust-exchanges list` | Search trust exchanges |
| `trust-exchanges get <id>` | Get a trust exchange |
| `trust-exchanges create --name=STRING --provider=STRING --issuer=STRING --allowed-audiences=STRING --output-mode=STRING` | Create a trust exchange |
| `trust-exchanges update --name=STRING --provider=STRING --issuer=STRING --allowed-audiences=STRING --output-mode=STRING <id>` | Update a trust exchange |
| `trust-exchanges delete <id>` | Delete a trust exchange |

### trust-providers

| Command | Description |
|---|---|
| `trust-providers list` | Search trust providers |
| `trust-providers get <id>` | Get a trust provider |
| `trust-providers create --name=STRING --signing-key-set-id=STRING --allowed-audiences=STRING` | Create a trust provider |
| `trust-providers update --name=STRING --signing-key-set-id=STRING --allowed-audiences=STRING <id>` | Update a trust provider |
| `trust-providers delete <id>` | Delete a trust provider |
| `trust-providers mint <id>` | Mint a token from a trust provider |
| `trust-providers discovery <id>` | Print federation (discovery/JWKS/token) URLs |

## Configuration

All commands accept `-o table` (default) or `-o json` for output format and `--debug` for verbose logging.

### Environment variables

| Variable | Description | Default |
|---|---|---|
| `MICROWAVE_TOKEN` | Management API key (required for all authenticated commands) | — |
| `MICROWAVE_API_URL` | Override the management API base URL | `https://api.microwave.sh` |
| `MICROWAVE_AUTH_URL` | Override the auth/token endpoint URL | `https://auth.microwave.sh` |
| `MICROWAVE_NO_UPDATE_CHECK` | Set to any non-empty value to disable the update check | — |
| `XDG_CONFIG_HOME` | Base directory for the config file | `~/.config` |
| `NO_COLOR` | Disable ANSI color output | — |

### Config file

Credentials and defaults are stored in:

```text
~/.config/microwave/config.toml
```

Run `microwave login` to populate it, or set `MICROWAVE_TOKEN` in the environment.

## CI/CD

Use `MICROWAVE_TOKEN` as a CI secret. Pass `--output json` to get machine-readable output:

```sh
# In GitHub Actions or any CI environment:
microwave keys list --output json | jq '.items[].id'
```

Example GitHub Actions step:

```yaml
- name: List key specs
  env:
    MICROWAVE_TOKEN: ${{ secrets.MICROWAVE_TOKEN }}
  run: microwave key-specs list --output json
```

## License

Apache 2.0 — see [LICENSE](LICENSE).
