# nqcli

nqcli is an internal CLI and lightweight web UI for running Gremlin and Cypher queries against the DTF SDR Neptune database through the AppSync GraphQL API. Authenticate with a token from the SDR UI before issuing requests.

## Quick Start

```bash
curl -fsSL https://raw.githubusercontent.com/ankit-lilly/nqcli/main/scripts/install.sh | bash
nq --help
```

## Features

- Run Gremlin (default) or Cypher queries from stdin or a file.
- Switch query language with `--type gremlin|cypher`.
- Launch a local browser UI via `nq server`.
- Pretty-printed JSON results on success; colorized errors on failure.

## Installation

### One-line installer (macOS & Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/ankit-lilly/nqcli/main/scripts/install.sh | bash
```

The script detects your OS/architecture, downloads the matching release, removes macOS quarantine attributes when needed, and installs `nq` to `/usr/local/bin`. Override the destination with `INSTALL_DIR=/your/path` or pin a specific release with `VERSION=vX.Y.Z`.

### Build from source

```bash
git clone https://github.com/ankit-lilly/nqcli.git
cd nqcli
make build
```

Set `VERSION=vX.Y.Z` when running `make build` to embed a custom version string in the binary.

## Configuration

`nqcli` reads configuration from environment variables and (optionally) a `.env` file. Pass `--env-file /path/to/file` to the CLI or `server` subcommand to load a specific file. When the flag is omitted the tool looks for `.env` in the current directory and then falls back to `~/.env`.

| Variable        | Description                                 | Default                               |
| --------------- | ------------------------------------------- | ------------------------------------- |
| `NEPTUNE_URL`   | AppSync GraphQL endpoint                    | Hard-coded demo URL (replace in prod) |
| `NEPTUNE_TOKEN` | Bearer token used for the `Authorization` header | `"jwt token"` (logs a warning)        |

Example `.env` file:

```dotenv
NEPTUNE_URL=https://your-appsync-id.appsync-api.us-east-1.amazonaws.com/graphql
NEPTUNE_TOKEN=eyJhbGciOi...your token...
```

> **Note:** The defaults are placeholders. Provide real values before running queries against production Neptune instances.

## CLI Usage

Once environment variables are set, use the binary directly:

```bash
# Pipe a Gremlin query (default type)
echo 'g.V().hasLabel("Person")' | nq

# Or just pass in the query as argument:

nq 'g.V().hasLabel("Person")'

# Pipe a Cypher query
echo 'MATCH (n) RETURN n LIMIT 5' | nq --type cypher

nq --type cyper 'MATCH (s:Study) return s.name'

# Execute from a file
nq path/to/query.gql --type gremlin

# Check the installed version
nq --version
```

If `--type` is omitted the command defaults to `gremlin` and validates that the supplied value is one of the supported options.

## Web UI

```bash
nq server --addr :8080
```

The server launches an interactive web UI at the provided address (default `0.0.0.0:8080`).

<img width="3456" height="2234" alt="Screenshot of the nqcli web UI" src="https://github.com/user-attachments/assets/24135c61-7d18-42b3-ab1b-a0bd1d2ea333" />

## Docker

Build a lightweight image and run the server without installing Go locally:

```bash
docker build -t nqcli:latest .
```

Inject configuration with individual variables or an env file:

```bash
# Using environment variables
docker run --rm -p 8080:8080 \
  -e NEPTUNE_URL=https://your-appsync-id.appsync-api.us-east-1.amazonaws.com/graphql \
  -e NEPTUNE_TOKEN=eyJhbGciOi... \
  nqcli:latest

# Or share a .env file (see .env.example)
docker run --rm -p 8080:8080 --env-file .env nqcli:latest
```

Override the address or run CLI commands by appending arguments when invoking the container:

```bash
docker run --rm nqcli:latest server --addr 0.0.0.0:9090
docker run --rm --env-file .env nqcli:latest --type cypher "MATCH (n) RETURN n LIMIT 5"
```

When using the CLI mode, pass input via stdin or mount a query file into the container.

## Screenshots

Additional examples live under `./screenshots`:

![CLI output](./screenshots/cli-screenshot.png)
![Web UI dark mode](./screenshots/dark-web.png)
![Web UI light mode](./screenshots/light-web.png)

## Limitations

- Only Gremlin and Cypher queries are supported.
- Tokens must be managed manually. After updating `NEPTUNE_TOKEN`, restart the server (`Ctrl+C` and rerun the `nq server ...` command) so the new value is loaded.
