# desec-cli Design Spec

A Go CLI client for the deSEC DNS API (https://desec.io/api/v1).

## Overview

`desec-cli` provides full coverage of the deSEC API: domain management, DNS record (RRset) management, token management, token policies, and dynamic DNS updates.

- **Language:** Go
- **Module path:** `github.com/swagner-de/desec-cli`
- **CLI framework:** Cobra + Viper
- **Output formats:** table (default), JSON, YAML via `--output` / `-o` flag

## Command Structure

```
desec-cli
├── domain
│   ├── list                    # List all domains
│   ├── get <domain>            # Get domain details
│   ├── create <domain>         # Create a domain (optional --zonefile)
│   ├── delete <domain>         # Delete a domain (confirms unless --yes)
│   └── export <domain>         # Export as zonefile
├── record
│   ├── list <domain>           # List RRsets (--type, --subname filters)
│   ├── get <domain> <subname> <type>   # Get specific RRset
│   ├── create <domain>         # Create RRset (--subname, --type, --ttl, --records)
│   ├── update <domain> <subname> <type> # Modify RRset (--ttl, --records)
│   ├── delete <domain> <subname> <type> # Delete RRset
│   └── bulk <domain>           # Bulk create/update from JSON file or stdin
├── token
│   ├── list                    # List all tokens
│   ├── get <id>                # Get token details
│   ├── create                  # Create token (--name, --subnets, --perms, --max-age, etc.)
│   ├── update <id>             # Modify token
│   └── delete <id>             # Delete token (confirms unless --yes)
├── token-policy
│   ├── list <token-id>         # List policies for a token
│   ├── get <token-id> <policy-id>
│   ├── create <token-id>       # Create policy (--domain, --subname, --type, --perm-write)
│   ├── update <token-id> <policy-id>
│   └── delete <token-id> <policy-id>
├── dyndns
│   └── update                  # Update IP (--hostname, --ip4, --ip6)
└── config
    └── init                    # Interactive setup: prompts for token, writes config
```

## Configuration & Authentication

**Config file:** `~/.config/desec-cli/config.yaml`

```yaml
token: "your-api-token"
output: table
```

**Precedence:** Environment variable (`DESEC_TOKEN`) overrides config file.

**API base URL:** Hardcoded to `https://desec.io/api/v1` (not self-hosted).

## Project Structure

```
desec-cli/
├── cmd/
│   ├── root.go              # Root command, config loading, global flags (--output)
│   ├── domain.go            # domain subcommands
│   ├── record.go            # record subcommands
│   ├── token.go             # token subcommands
│   ├── token_policy.go      # token-policy subcommands
│   ├── dyndns.go            # dyndns subcommands
│   └── config.go            # config init command
├── internal/
│   ├── client/
│   │   └── client.go        # HTTP client wrapping the deSEC API
│   ├── output/
│   │   └── output.go        # Table/JSON/YAML formatting
│   └── config/
│       └── config.go        # Config file + env var loading via Viper
├── main.go                  # Entry point
├── go.mod
└── go.sum
```

## Architecture

### HTTP Client (`internal/client`)

A struct holding token and base URL. One method per API call (e.g., `ListDomains()`, `CreateRRset()`). Returns typed Go structs. Handles pagination by following `Link` headers internally.

### Output Formatter (`internal/output`)

A `Print(format string, data any)` function that switches on `table`/`json`/`yaml`. Table rendering uses `tablewriter`. Each resource type has its own column definitions.

### Commands (`cmd/`)

Each file wires up Cobra commands. Commands parse flags, call the client, and pass results to the output formatter. Minimal logic in command files.

### Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/spf13/viper` — config management
- `github.com/olekukonez/tablewriter` — table output
- `gopkg.in/yaml.v3` — YAML output
- stdlib `encoding/json`, `net/http`, `fmt`, `os`

## Error Handling

- **API errors:** Parsed from JSON response body, displayed as readable messages with HTTP status.
- **Auth errors:** Early exit if no token found, with guidance to set `DESEC_TOKEN` or run `desec-cli config init`.
- **Destructive operations:** `domain delete` and `token delete` prompt for confirmation unless `--yes` / `-y` is passed.

## Record Values

Passed as repeatable `--record` flags:

```
desec-cli record create example.com --subname www --type A --ttl 3600 --record 1.2.3.4 --record 5.6.7.8
```

Zone apex is represented as `@` on the CLI, mapped to empty string for the API.

## Bulk Operations

`desec-cli record bulk` reads JSON from stdin or a file (`--file`):

```
cat records.json | desec-cli record bulk example.com
desec-cli record bulk example.com --file records.json
```

Operations are atomic (all succeed or all fail) per the deSEC API.
