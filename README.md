# nqcli

nqcli is an internal CLI and lightweight web UI for running Gremlin and Cypher queries against the DTF SDR Neptune database through the AppSync GraphQL API. Requests are signed with AWS IAM (SigV4).

## Quick Start

```bash
curl -fsSL https://raw.githubusercontent.com/ankit-lilly/nqcli/main/scripts/install.sh | bash
nq --help
```

The installer detects your OS/architecture, downloads the matching release,
and installs `nq` to `/usr/local/bin`. Override the destination with
`INSTALL_DIR=/your/path` or pin a specific release with `VERSION=vX.Y.Z`.

## Features

- Run Gremlin (default) or Cypher queries from stdin or a file.
- Switch query language with `--type gremlin|cypher`.
- Pretty-printed JSON results on success; colorized errors on failure.

## Configuration

`nqcli` can read configuration from environment variables or an optional `.env` file. You usually do not need a `.env` because the AppSync URL is discovered automatically from the active AWS profile. If you do want a file, pass `--env-file /path/to/file` to the CLI. When omitted the tool looks for `.env` in the current directory and then falls back to `~/.env`.

| Variable                     | Description                                                                 | Default   |
| ---------------------------- | --------------------------------------------------------------------------- | --------- |
| `NEPTUNE_URL`                | AppSync GraphQL endpoint (overrides discovery)                              |           |
| `NEPTUNE_APPSYNC_API_NAME`   | AppSync API name to select when discovering the endpoint                    |           |
| `NEPTUNE_APPSYNC_API_ID`     | AppSync API ID to select when discovering the endpoint                      |           |

When `NEPTUNE_URL` is unset, the CLI calls `appsync:ListGraphqlApis` for the
current `--aws-profile` (or `AWS_PROFILE`) and region to resolve the URL. The
result is cached in `~/.cache/nqcli/appsync_cache.json` (keyed by
profile+region). Delete the cache file to force a refresh.

Example `.env` file (only needed if you want to override discovery):

```dotenv
NEPTUNE_URL=https://your-appsync-id.appsync-api.us-east-1.amazonaws.com/graphql
```

> **Note:** If multiple AppSync APIs exist in the account/region, set
> `NEPTUNE_APPSYNC_API_NAME` or `NEPTUNE_APPSYNC_API_ID` to disambiguate.

### IAM authentication

`nqcli` signs AppSync requests with AWS SigV4, so you must provide AWS credentials with access to the AppSync API. Use an AWS profile, environment variables, or IAM role credentials in your execution environment.

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

Use `--aws-profile` or `--aws-region` to control which AWS credentials are used when signing requests.

## Web UI

```bash
nq server --addr :8080
```

The server launches an interactive web UI at the provided address (default `0.0.0.0:8080`).

## Screenshots

Additional examples live under `./screenshots`:

![CLI output](./screenshots/cli-screenshot.png)
![Web UI dark mode](./screenshots/dark-web.png)
![Web UI light mode](./screenshots/light-web.png)

## Limitations

- Only Gremlin and Cypher queries are supported.
