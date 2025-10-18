# nqcli

This is just an internal tool to execute Gremlin and Cypher queries. This is specifically made for DTF SDR. It uses a GraphQL API exposed by the DTF Backend to run queries against the database. 

You need a token to authenticate against the API. You can get it from the SDR UI.

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

Clone the repository and install the binary:

```bash
    git clone https://github.com/ankit-lilly/nqcli.git
    cd nqcli
    make build
```

If you just want to compile locally:

```bash
make build
```

---

## Configuration

`nqcli` reads its configuration from environment variables (and optionally a `.env` file if present):

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
```

If `--type` is omitted it defaults to `gremlin`. The command validates that the type is either `gremlin` or `cypher` before execution.

Errors are printed in color to stderr; successful responses are pretty-printed JSON on stdout.

---

## Web UI

Launch the embedded HTTP server for a browser-based experience:

```bash
./nq server --addr :8080
```

