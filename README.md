# desec-cli

A command-line client for the [deSEC](https://desec.io) DNS API.

## Installation

### From release binaries

Download the latest release from the [releases page](https://github.com/swagner-de/desec-cli/releases).

### From source

```bash
go install github.com/swagner-de/desec-cli@latest
```

## Configuration

### Config file

Run the interactive setup:

```bash
desec-cli config init
```

This creates `~/.config/desec-cli/config.yaml`:

```yaml
token: "your-api-token"
output: table
```

### Environment variable

```bash
export DESEC_TOKEN="your-api-token"
```

The environment variable takes precedence over the config file.

## Usage

### Domain management

```bash
desec-cli domain list
desec-cli domain get example.com
desec-cli domain create example.com
desec-cli domain create example.com --zonefile "$(<zonefile.txt)"
desec-cli domain export example.com
desec-cli domain delete example.com
```

### DNS records

```bash
desec-cli record list example.com
desec-cli record list example.com --type A
desec-cli record list example.com --subname www

desec-cli record get example.com www A

desec-cli record create example.com --subname www --type A --ttl 3600 \
  --record 1.2.3.4 --record 5.6.7.8

desec-cli record update example.com www A --ttl 300
desec-cli record update example.com www A --record 9.10.11.12

desec-cli record delete example.com www A
```

Use `@` for the zone apex:

```bash
desec-cli record create example.com --subname @ --type MX --ttl 3600 \
  --record "10 mail.example.com."
```

#### Bulk operations

```bash
desec-cli record bulk example.com --file records.json
cat records.json | desec-cli record bulk example.com
```

The JSON file should contain an array of RRset objects:

```json
[
  {"subname": "www", "type": "A", "ttl": 3600, "records": ["1.2.3.4"]},
  {"subname": "mail", "type": "A", "ttl": 3600, "records": ["5.6.7.8"]}
]
```

### Token management

```bash
desec-cli token list
desec-cli token get <token-id>
desec-cli token create --name "my-token"
desec-cli token create --name "restricted" --perm-create-domain --subnet 10.0.0.0/8
desec-cli token update <token-id> --name "renamed"
desec-cli token delete <token-id>
```

### Token policies

Restrict tokens to specific RRsets:

```bash
# List policies
desec-cli token-policy list <token-id>

# Create a default deny-all policy (required as the first policy)
desec-cli token-policy create <token-id>

# Allow writing TXT records for ACME challenges
desec-cli token-policy create <token-id> \
  --domain example.com \
  --subname _acme-challenge \
  --type TXT \
  --perm-write

# Delete a policy
desec-cli token-policy delete <token-id> <policy-id>
```

### Dynamic DNS

```bash
desec-cli dyndns update --hostname example.dedyn.io --ipv4 1.2.3.4
desec-cli dyndns update --hostname example.dedyn.io --ipv6 2001:db8::1
desec-cli dyndns update --hostname example.dedyn.io --ipv4 1.2.3.4 --ipv6 2001:db8::1
```

## Output formats

All commands support `--output` / `-o` with `table` (default), `json`, or `yaml`:

```bash
desec-cli domain list -o json
desec-cli record list example.com -o yaml
```

Set the default in your config file:

```yaml
output: json
```

## License

[MIT](LICENSE)
