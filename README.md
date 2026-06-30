# dev

One CLI for an AI-assisted development workflow. Instead of remembering the
flags of `graphify`, `codebase-memory-mcp`, `git`, Docker and friends, you run
a handful of memorable verbs and `dev` orchestrates the rest.

```
dev init      check & set up your dev environment
dev setup     configure graphify & codebase-memory-mcp
dev sync      graphify update → memory reindex → git status
dev graph     build the graph, or update it if it exists
dev update    re-extract & update the graph
dev memory    index the codebase memory (optional tool)
dev doctor    show the health of every dependency
dev ai        check AI CLIs, API keys & proxy
dev stats     show knowledge-graph statistics
dev clean     remove generated graphify-out/ artifacts
```

## Install

Requires Go 1.22+.

```sh
# from a clone, install to ~/.local/bin
make install

# or straight from the module path
go install github.com/duclq/dev/cmd/dev@latest
```

Make sure the install dir is on your `PATH`:

```sh
# ~/.local/bin (make install) or ~/go/bin (go install)
export PATH="$HOME/.local/bin:$HOME/go/bin:$PATH"
```

### Homebrew (optional)

Once the repo is published and tagged you can ship it through a personal tap:

```sh
brew tap duclq/tap
brew install dev
```

(The tap holds a formula that runs `go install` / downloads a release binary.)

## Commands

| Command       | What it does                                                                 |
| ------------- | ---------------------------------------------------------------------------- |
| `dev init`    | Verifies graphify, git, node, python, docker, claude, etc. and reports project state. |
| `dev setup`   | Configures both tools (`dev setup graphify` / `dev setup memory` for one).    |
| `dev doctor`  | Full health check of core deps + services (redis, postgres) + project.       |
| `dev graph`   | Runs `graphify extract .` the first time, `graphify update .` afterwards.     |
| `dev update`  | Always runs `graphify update .`.                                             |
| `dev sync`    | `graphify` → memory re-index → `git status` in one shot.                      |
| `dev memory`  | Runs `codebase-memory-mcp cli index_repository` for the current dir (incremental). Skips gracefully if the tool isn't installed. |
| `dev ai`      | Lists detected AI CLIs (Claude, Gemini, Codex, OpenAI), API keys, proxies.    |
| `dev stats`   | Node / edge / community counts from `graphify-out/graph.json`.                |
| `dev clean`   | Removes `graphify-out/` (`-f`/`--force` to skip the prompt).                  |

## Setup: what `dev setup` automates

The two tools each take several manual steps; `dev setup` does them for you
(idempotent — safe to re-run).

### `dev setup graphify`

1. Checks that the `graphify` binary is present.
2. Installs the skill + hook for your agent: `graphify <platform> install`
   (writes the `CLAUDE.md` section + PreToolUse hook). Default platform is
   `claude`; override with `dev setup graphify --platform codex` (or `cursor`,
   `gemini`, ...).
3. Detects the LLM backend graphify will use. graphify auto-selects from
   whichever key is set — `ANTHROPIC_API_KEY` / `OPENAI_API_KEY` /
   `GEMINI_API_KEY` / `DEEPSEEK_API_KEY` / `MOONSHOT_API_KEY` — or a self-hosted
   endpoint (`OPENAI_BASE_URL` / `ANTHROPIC_BASE_URL`), or local `ollama`. Warns
   if none is configured.

### `dev setup memory`

1. If `codebase-memory-mcp` isn't installed, runs the official one-line
   installer **only after you confirm** (or pass `--install` / `-y`):
   `curl -fsSL .../install.sh | bash`. The installer downloads a static binary
   and **auto-registers the MCP server** with Claude Code and other detected
   agents — no manual `~/.claude/.mcp.json` editing.
2. Enables incremental indexing: `codebase-memory-mcp config set auto_index true`.

No API keys, Redis, or Postgres are needed — codebase-memory-mcp is a single
static binary with embedded SQLite that runs 100% locally.

```sh
dev setup            # both tools
dev setup graphify   # just graphify
dev setup memory -y  # install memory without the prompt
```

### Graceful by design

Optional tools (`codebase-memory-mcp`, Docker, Redis, PostgreSQL) never crash a
command — they warn and continue. Only genuinely required tools (graphify, git,
node, python) count toward a failing `dev doctor`.

## Layout

```
cmd/dev/main.go          # command dispatch + usage
internal/ui/             # colored output helpers
internal/tools/          # binary lookup + process running
internal/commands/       # one file per command
```

## Development

```sh
make build     # → bin/dev
make run ARGS="doctor"
make vet
make fmt
```

Set `NO_COLOR=1` to disable ANSI colors.
