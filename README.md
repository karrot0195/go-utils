# go-utils

A collection of Go utilities. Currently ships a single CLI tool, `openroute`,
which fetches and displays model listings from the [OpenRouter](https://openrouter.ai)
API in a benchmark-sorted table.

Built with **Go 1.26.9** and the standard library only — no external dependencies.

## Quick Start

```bash
make build
./bin/openroute --type free
```

## Usage

```
openroute [--type free|all]
```

| Flag     | Default | Description                                   |
| -------- | ------- | --------------------------------------------- |
| `--type` | `free`  | Filter models: `free` (id ends with `:free`) or `all` |

### Fetch Free Models

```bash
./bin/openroute --type free
```

### Fetch All Models

```bash
./bin/openroute --type all
```

### Clean Build

```bash
make clean
```

## Features

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

## Project Structure

```
go-utils/
├── assets/
│   └── cmd_openroute.png      # CLI diagram
├── bin/                       # build output (./bin/openroute)
├── cmd/
│   └── openroute/
│       └── main.go            # CLI entry point
├── packages/
│   ├── envs/                  # environment configuration (placeholder)
│   ├── log/                   # logging utilities (placeholder)
│   ├── openroute/             # OpenRouter API integration
│   │   ├── client.go          # fetch, cache, parse, filter
│   │   ├── model.go           # data structures
│   │   └── printer.go         # tabular output & sorting
│   └── vultr/                 # Vultr API integration (placeholder)
├── scripts/                   # (empty)
├── makefile                   # build / clean targets
├── go.mod                     # module go-utils (Go 1.26.9)
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

### Placeholder Packages

- `packages/envs` — environment configuration (empty).
- `packages/log` — logging utilities (empty).
- `packages/vultr` — Vultr API integration (empty).

## Example Output

Illustrative sample of the table format (values are examples, not live data):

```
ID                                                     NAME                                                   CODING_INDEX    INTELL_INDEX    SUM             PROMPT          COMPLETION      CONTEXT
anthropic/claude-3.5-sonnet:free                        Claude 3.5 Sonnet (self-moderated)                      92.00           88.00           180.00          3.0000          15.0000         200000
...
```

## Requirements

- Go 1.26.9+
- Standard library only (no external dependencies)
- Network access to `openrouter.ai` on first run (subsequent runs use the cache)
