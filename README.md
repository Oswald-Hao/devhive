# DevHive

Multi-agent software development system — autonomous coding with verify-specialized agents, failure signatures, and structured handoff protocols.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![npm](https://img.shields.io/npm/v/@oswaldzsh/devhive?color=CB3837&logo=npm)](https://www.npmjs.com/package/@oswaldzsh/devhive)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## Architecture

```
SPECIFY → EXECUTE → VERIFY_L1 (static) → VERIFY_L2 (dynamic/semantic) → MERGE
```

- **Orchestrator Engine** — goroutine-based task queue, agent pool, and event bus
- **Execute Agent** — calls the AI model, produces structured code changes
- **Verifier Agents** — Static (rule engine), Dynamic (test runner), Semantic (spec alignment)
- **Convergence Gate** — loop detection and escalation when tasks don't converge
- **Signature Engine** — Pure Go weighted similarity matching for error pattern recognition
- **Checkpoint Store** — SQLite-based task state persistence

## Install

### npm (recommended)

```bash
npm install -g @oswaldzsh/devhive
```

### from source

```bash
git clone https://github.com/Oswald-Hao/devhive.git
cd devhive
go build -o ~/.devhive/bin/devhive ./cmd/dh/
export PATH="$HOME/.devhive/bin:$PATH"
```

### from GitHub Releases

Download the latest binary for your platform from [Releases](https://github.com/Oswald-Hao/devhive/releases).

## Quick Start

```bash
devhive --init                        # Create ~/.devhive/config.yaml
# Edit ~/.devhive/config.yaml with your API credentials
devhive                  # Start interactive chat
devhive --help           # Show all flags
devhive --resume         # Resume last session
echo "explain Go interfaces" | devhive --no-tui        # Single question
echo '{"task":"add login"}' | devhive --no-tui --json  # JSON output
```

Set your API credentials:

```bash
devhive --init
# Then edit ~/.devhive/config.yaml:
#   api:
#     base_url: "https://your-api.example.com"
#     auth_token: "your-token-here"
```

## Chat Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/clear` | Clear conversation history |
| `/model <name>` | Switch AI model |
| `/save <file>` | Save conversation to file |
| `/quit`, `/q` | Exit DevHive |

## CLI Flags

```
USAGE:
  devhive [flags]

FLAGS:
  -h, --help        Show help
  -v, --version     Show version
  --init            Generate ~/.devhive/config.yaml template
  --json            Output in JSON format
  --no-tui          Read a single prompt from stdin
  --resume          Resume the last session
  --model <name>    Override AI model
```

## Configuration

DevHive reads configuration from `~/.devhive/config.yaml`. Run `devhive --init` to create a template.

| Setting | Description |
|---------|-------------|
| `api.base_url` | API endpoint URL (required) |
| `api.auth_token` | API authentication token (required) |
| `api.default_model` | Default model name (optional) |

Environment variables (`ANTHROPIC_AUTH_TOKEN`, `LEJU_TOKEN`, `ANTHROPIC_BASE_URL`, `DEVHIVE_MODEL`) override config file values.

## Development

```bash
git clone https://github.com/Oswald-Hao/devhive.git
cd devhive
go build ./cmd/dh/        # Build CLI
go test ./...             # Run tests
```

### Release

```bash
# Bump version and package
VERSION=0.3.0 bash scripts/release.sh
```

## License

MIT
