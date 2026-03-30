# Microwave CLI

Command-line interface for the Microwave API

## Install

**Homebrew (recommended):**

```sh
brew install microwave-sh/microwave/microwave
```

**Go install:**

```sh
go install github.com/microwave-sh/microwave-cli@latest
```

## Quick Start

```sh
# Authenticate
microwave login

# Text analysis
microwave text analyze --input "The quick brown fox"

# Image resizing
microwave image resize --url "https://example.com/photo.jpg" --width 800

# Cryptographic hashing
microwave crypto hash --algo sha256 --input "hello world"

# IP geolocation
microwave geo lookup --ip "1.2.3.4"

# Address validation
microwave address validate --input "1600 Pennsylvania Ave NW, Washington, DC"

# Currency conversion
microwave financial fx --from USD --to EUR --amount 100

# Math / statistics
microwave math stats --input "1,2,3,4,5"

# Timezone conversion
microwave time convert --from "America/New_York" --to "Europe/London" --at "2026-06-01T09:00:00"

# Encoding / hashing
microwave encoding base64 encode --input "hello"
```

## Configuration

| Variable             | Description                                    | Default                      |
|----------------------|------------------------------------------------|------------------------------|
| `MICROWAVE_API_KEY`  | Your Microwave API key (required for most ops) | —                            |
| `MICROWAVE_BASE_URL` | Override the API base URL                      | `https://api.microwave.dev`  |

You can also pass the key directly with the `-k` / `--api-key` flag on any command.

## License

Apache 2.0 — see [LICENSE](LICENSE).
