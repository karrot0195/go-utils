# go-utils

A collection of Go utilities shipping two CLI tools:

- **`openroute`** — fetches and displays model listings from the [OpenRouter](https://openrouter.ai) API in a benchmark-sorted table.
- **`vultrcli`** — manages Vultr cloud instances (create/destroy workers) via the `vultr-cli` binary.

Built with **Go 1.26.9**. Uses `github.com/joho/godotenv` for environment loading and `github.com/sirupsen/logrus` for logging; otherwise standard library only.

## Quick Start

```bash
# Build both CLI tools
make build

# Fetch free OpenRouter models
./bin/openroute --type free

# Manage Vultr instances (requires vultr-cli installed)
./bin/vultrcli --action create --size small
```

## Usage

### `openroute`

```
openroute [--type free|all]
```

| Flag     | Default | Description                                   |
| -------- | ------- | --------------------------------------------- |
| `--type` | `free`  | Filter models: `free` (id ends with `:free`) or `all` |

### `vultrcli`

```
vultrcli --action create|delete|destroy [--size high|medium|small] [--img <snapshot>] [--ssh-name <key>] [--api-key <key>]
```

| Flag         | Default                  | Description                                                                     |
| ------------ | ------------------------ | ------------------------------------------------------------------------------- |
| `--action`   | *(required)*             | Action to perform: `create`, `delete`, `destroy`                                |
| `--size`     | `small`                  | VM size for create: `high` (8C/16GB), `medium` (4C/8GB), `small` (1C/2GB)       |
| `--img`      | `retranscode-worker`     | Snapshot image name for create action                                           |
| `--ssh-name` | `vultr-dev-access`       | SSH key name (resolved to ID via vultr-cli)                                     |
| `--api-key`  | `$API_KEY` env var       | API key for job-stats endpoint (required for `destroy`)                         |

Environment variables (can be set in `.env`):

| Variable           | Default                              | Description                          |
| ------------------ | ------------------------------------ | ------------------------------------ |
| `WORKER_LABEL`     | `retranscode-worker`                 | Prefix for worker instance labels    |
| `JOB_STATS_URL`    | `https:/retranscoding.cloud/job-stats` | Job stats endpoint URL               |
| `DEFAULT_REGION`   | `sjc`                                | Default Vultr region                 |
| `DEFAULT_SSH_NAME` | `vultr-dev-access`                   | Default SSH key name                 |
| `API_TIMEOUT`      | `15s`                                | HTTP timeout for API calls           |
| `MAX_RESPONSE_SIZE`| `1048576`                            | Max response size for job-stats (bytes) |

### `openroute` Examples

```bash
# Fetch free models (default)
./bin/openroute --type free

# Fetch all models
./bin/openroute --type all
```

### `vultrcli` Examples

```bash
# Create a small worker instance
./bin/vultrcli --action create --size small

# Create a high-spec worker with custom snapshot
./bin/vultrcli --action create --size high --img my-snapshot --ssh-name my-key

# Destroy ready workers (checks job-stats first)
./bin/vultrcli --action destroy --api-key $API_KEY

# Clean build artifacts
make clean
```

## Features

### `openroute`

- **API fetching** — downloads the model list from `https://openrouter.ai/api/v1/models`.
- **Intelligent caching** — results are pretty-printed and saved to a dated cache
  file in the OS temp directory (e.g. `151026-model_list.json`). If today's cache
  already exists, it is reused and no network call is made.
- **Filtering** — `free` keeps only models whose id ends with `:free`; `all` keeps everything.
- **Benchmark sorting** — models are ranked by combined `coding_index + intelligence_index`
  (Artificial Analysis benchmarks), descending.
- **Tabular output** — columns for id, name, coding index, intelligence index,
  combined sum, and per-million-token pricing.
- **Price formatting** — OpenRouter per-token prices are scaled to per-million-token
  cost for readability.

### `vultrcli`

- **Instance lifecycle** — create and destroy Vultr cloud instances via `vultr-cli`.
- **Snapshot-based creation** — creates instances from a named snapshot image.
- **Worker labeling** — instances labeled with `<WORKER_LABEL>-<N>` for tracking.
- **Safe destruction** — `destroy` action checks job-stats endpoint first; refuses to
  destroy if pending jobs exist (accepted/draft > 0 or jobs array non-empty).
- **Readiness checks** — only destroys instances with `status=active`, `power_status=running`, `server_status=ok`.
- **Environment-driven config** — all settings via env vars or `.env` file (via `godotenv`).
- **Structured logging** — uses `logrus` for JSON-formatted output.

## Project Structure

```
go-utils/
├── assets/
│   └── cmd_openroute.png      # CLI diagram
├── bin/                       # build output (./bin/openroute, ./bin/vultrcli)
├── cmd/
│   ├── openroute/
│   │   └── main.go            # openroute CLI entry point
│   └── vultrcli/
│       └── main.go            # vultrcli CLI entry point
├── packages/
│   ├── config/
│   │   └── env.go             # environment configuration (.env + godotenv)
│   ├── log/
│   │   └── logger.go          # logging utilities (logrus wrapper)
│   ├── openroute/             # OpenRouter API integration
│   │   ├── client.go          # fetch, cache, parse, filter
│   │   ├── model.go           # data structures
│   │   └── printer.go         # tabular output & sorting
│   └── vultr/
│       └── cli.go             # Vultr CLI wrapper (instance/snapshot/ssh-key ops)
├── scripts/                   # (empty)
├── makefile                   # build / clean targets
├── go.mod                     # module go-utils (Go 1.26.9)
├── go.sum                     # dependency checksums
└── README.md
```

## Packages

### `packages/openroute`

Core package for interacting with the OpenRouter API.

| Function        | Description                                                              |
| --------------- | ------------------------------------------------------------------------ |
| `FetchModels()` | Downloads the model list, pretty-prints it, writes a dated cache file, and returns the raw JSON bytes plus the cache path. Reuses today's cache when present. |
| `ParseModels()` | Unmarshals raw JSON into a `ModelResponse`.                              |
| `FilterModels()`| Returns models matching the given type (`free` or `all`).                 |
| `PrintModels()` | Filters, sorts by combined benchmark score, and prints a table.           |

**Data structures** (`model.go`):

- `Model` — id, name, description, pricing (prompt/completion), context length,
  and Artificial Analysis benchmarks (agentic, coding, intelligence indices).
- `ModelResponse` — top-level envelope wrapping `data []Model`.

### `packages/vultr`

Wrapper around the `vultr-cli` binary for programmatic instance management.

| Type / Function         | Description                                                              |
| ----------------------- | ------------------------------------------------------------------------ |
| `VultrCmd`              | Main client struct holding context.                                      |
| `NewVultrCmd(ctx)`      | Creates a new VultrCmd (uses `context.Background()` if nil).             |
| `GetInstances()`        | Returns all Vultr instances via `vultr-cli instance list -o json`.       |
| `RemoveInstanceByIP()`  | Deletes the instance matching the given primary IP.                      |
| `Instance`              | Instance struct with ID, Label, Status, ServerStatus, PowerStatus, IPs.  |
| `Instance.PrimaryIP()`  | Returns `MainIP` if set, falls back to `IP`.                             |

### `packages/config`

Environment configuration loader.

| Function         | Description                                                              |
| ---------------- | ------------------------------------------------------------------------ |
| `NewEnv()`       | Creates a new Env instance.                                              |
| `Env.LoadEnv()`  | Loads `.env` file via `godotenv`.                                        |
| `Env.Get(key, default...)` | Returns env var value or default.                                |

### `packages/log`

Logging utilities.

| Function      | Description                                                              |
| ------------- | ------------------------------------------------------------------------ |
| `NewLogger()` | Returns a `*Logger` wrapping a `logrus.Logger` instance.                 |
| `Logger`      | Embeds `*logrus.Logger` for full logrus API access.                      |

## Example Output

Illustrative sample of the table format (values are examples, not live data):

```
ID                                                     NAME                                                   CODING_INDEX    INTELL_INDEX    SUM             PROMPT          COMPLETION      CONTEXT
anthropic/claude-3.5-sonnet:free                        Claude 3.5 Sonnet (self-moderated)                      92.00           88.00           180.00          3.0000          15.0000         200000
...
```

## Requirements

- Go 1.26.9+
- Dependencies: `github.com/joho/godotenv`, `github.com/sirupsen/logrus`
- **For `openroute`**: Network access to `openrouter.ai` on first run (subsequent runs use the cache)
- **For `vultrcli`**: `vultr-cli` binary installed and authenticated; network access to Vultr API and job-stats endpoint
