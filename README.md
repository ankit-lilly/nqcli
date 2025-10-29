# nqcli

This is just an internal tool to execute Gremlin and Cypher queries. This is specifically made for DTF SDR. It uses a GraphQL API exposed by the DTF Backend to run queries against the database.  You need a token to authenticate against the API. You can get it from the SDR UI.

---

## Features


You can either use it as CLI:

```sh
    echo "g.V().hasLabel('Study').elementMap()" | ./nq
```

OR 

save your gremlin/cypher query in a file (extension doesn't matter)"

```sh
    ./nq path/to/query.groovy 
```
OR 

```sh
    ./nq --type cypher "MATCH (n) RETURN n LIMIT 5"
```

Note that default query type is gremlin. You can override it with `--type` flag.

```sh
    echo "MATCH (n) RETURN n LIMIT 5" | ./nq --type cypher
```

## Installation

### One-line installer

You can install the latest release binary directly (macOS & Linux) with:

```bash
curl -fsSL https://raw.githubusercontent.com/ankit-lilly/nqcli/main/scripts/install.sh | bash
```

The script detects your OS/architecture, downloads the matching release asset, clears the macOS quarantine attribute when necessary, and installs `nq` into `/usr/local/bin` (override with `INSTALL_DIR=/your/path`). Set `VERSION=vX.Y.Z` to pin a specific release.


### Build from source

```bash
    git clone https://github.com/ankit-lilly/nqcli.git
    cd nqcli
    make build
```

If you just want to compile locally:

```bash
make build
```

Set `VERSION=vX.Y.Z` when invoking `make build` to embed a specific version string in the compiled binary.

## Configuration

`nqcli` reads its configuration from environment variables (and optionally a `.env` file if present):

Use `--env-file /path/to/file` (available on both CLI and `server` subcommand) to explicitly load a `.env` file. When the flag is omitted the tool searches for `.env` in the current directory first, then falls back to `~/.env`.

| Variable         | Description                                              | Default                                 |
| ---------------- | -------------------------------------------------------- | --------------------------------------- |
| `NEPTUNE_URL`    | AppSync GraphQL endpoint                                 | Hard-coded demo URL (replace in prod)   |
| `NEPTUNE_TOKEN`  | Bearer token for the `Authorization` header              | `"jwt token"` (logs a warning)          |

Example `.env`:

```dotenv
NEPTUNE_URL=https://your-appsync-id.appsync-api.us-east-1.amazonaws.com/graphql
NEPTUNE_TOKEN=eyJhbGciOi...your token...
```

> **Note:** The defaults are placeholders. Set real values before running queries against production Neptune instances.

---

## CLI Usage

Run the binary directly after setting the env vars. You can either pipe a query or point to a file:

```bash
# Pipe a Gremlin query (default type)
echo 'g.V().hasLabel("Person")' | ./nq

# Pipe a Cypher query
echo 'MATCH (n) RETURN n LIMIT 5' | ./nq --type cypher

# Execute from file
./nq path/to/query.gql --type gremlin

# Check the installed version
./nq --version
```

If `--type` is omitted it defaults to `gremlin`. The command validates that the type is either `gremlin` or `cypher` before execution.

Errors are printed in color to stderr; successful responses are pretty-printed JSON on stdout.

---

## Web UI

Launch the embedded HTTP server for a browser-based experience:

```bash
./nq server --addr :8080
```


<img width="3456" height="2234" alt="image" src="https://github.com/user-attachments/assets/24135c61-7d18-42b3-ab1b-a0bd1d2ea333" />


---

## Docker

Build a tiny image and run the server without installing Go locally:

```bash
docker build -t nqcli:latest .
```

Provide configuration either with individual variables or an env file:

```bash
# Using env vars
docker run --rm -p 8080:8080 \
  -e NEPTUNE_URL=https://your-appsync-id.appsync-api.us-east-1.amazonaws.com/graphql \
  -e NEPTUNE_TOKEN=eyJhbGciOi... \
  nqcli:latest

# Or share a .env file (see .env.example)
docker run --rm -p 8080:8080 --env-file .env nqcli:latest
```

By default the container starts the web UI on `0.0.0.0:8080`. Override the address or run CLI commands by appending the desired arguments:

```bash
docker run --rm nqcli:latest server --addr 0.0.0.0:9090
docker run --rm --env-file .env nqcli:latest --type cypher "MATCH (n) RETURN n LIMIT 5"
```

When using the CLI mode, remember to pass input via stdin or files mounted into the container.


## Screenshots


![cli](./screenshots/cli-screenshot.png)

![web-darkmode](./screenshots/dark-web.png)
![web-lightmode](./screenshots/light-web.png)



## Limitations


- Only supports Gremlin and Cypher queries.
- You need to manage the token yourself. If you are using webui and the token is invalid/expired, you need to restart the server after updating the token in env vars.
So you just need to do ctrl+c and re-run the server command after updating the token.
