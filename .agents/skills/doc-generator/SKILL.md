---
name: doc-generator
description: Generate or update project documentation (such as README.md, FLOW.md, or provider specs) based on the repository tree, codebase features, tech stack, and workflows. Use when the user asks to create, update, or maintain documentation files.
---

# Documentation Generator Skill

A skill for generating, updating, and structuring project documentation files like `README.md`, `FLOW.md`, or architecture specifications.

## When to Use
Use this skill when the user asks to create or update project documentation, workflow files, structure docs, or markdown guides based on the active codebase layout and design.

## Workflow

1. **Analyze Workspace & Scope**: Review the repository tree, source files, directory layout (e.g., `docs/`, `internal/`, `bin/`), and build configurations (`makefile`).
2. **Determine Target Document**: Identify whether you are updating `README.md`, creating a flow document (`docs/FLOW.md`), or documenting a provider spec.
3. **Synthesize Content**: 
   - Ensure accurate stack details and setup/build steps (`make build`, `./bin/...`).
   - Match exact directory trees and schema references.
4. **Write/Update Markdown**: Generate clear, scannable markdown files using proper headings, code blocks, and lists.
   - Auto-load images from `./assets` (e.g., `<img src="./assets/linux.png" />` as in README.md).

## Current Project Analysis

This skill is designed for the `go-utils` project. Here's what I've discovered:

### Project Structure
- **Root**: `/home/kai/Workspaces/go-utils`
- **Module**: `go-utils` (Go 1.26.9)
- **Build System**: Makefile with `build` and `clean` targets
- **Main Binary**: `cmd/openroute/main.go` - CLI tool for fetching OpenRouter model data

### Key Packages
- `packages/openroute/` - Core package for interacting with OpenRouter API
  - `client.go` - HTTP client for fetching model data with caching
  - `model.go` - Data structures for OpenRouter model responses
  - `printer.go` - Tabular output formatting
- `packages/envs/` - Environment configuration (empty)
- `packages/log/` - Logging utilities (empty)
- `packages/vultr/` - Vultr API integration (empty)

### Build Instructions
```bash
make build
./bin/openroute --type free
```

### Documentation Targets
This skill can generate:
- `README.md` - Project overview and setup instructions
- `docs/FLOW.md` - Workflow documentation
- Provider specifications for each package

### Assets
- `./assets/cmd_openroute.png` - Binary diagram image

### Dependencies
- Go 1.26.9+
- Standard library only (no external dependencies)
- Uses `net/http`, `encoding/json`, and standard Go tooling

### Features
- Model fetching from OpenRouter API with intelligent caching
- Free vs. all model filtering
- Tabular output with benchmarks (coding index, intelligence index)
- Price formatting (scaled to per-million tokens)
- Sorting by combined benchmark scores

## Usage Examples

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

## Next Steps
When generating documentation, this skill will:
1. Verify build instructions work
2. Document package interfaces and dependencies
3. Include code examples and usage patterns
4. Reference supporting assets where appropriate
5. Maintain consistency with existing documentation style