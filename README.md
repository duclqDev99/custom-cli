# dev

One CLI for an AI-assisted development workflow. Instead of remembering the
flags of `graphify`, `codebase-memory-mcp`, `git`, Docker and friends, you run
a handful of memorable verbs and `dev` orchestrates the rest.

Commands come in two layers: **workflow** verbs that orchestrate every tool at
once, and per-tool **namespaces**.

```
# workflow (cross-cutting)
dev init                 check & set up your dev environment
dev setup [tool]         configure all tools, or just one
dev sync                 refresh every tool for this project
dev doctor               health of every tool & dependency
dev ai                   check AI CLIs, API keys & proxy

# tool namespaces
dev graphify <verb>      graph | extract | update | stats | clean
dev memory   <verb>      index | status
dev uipro    <verb>      init | update | versions

# shortcuts (aliases to a namespace verb)
dev graph    → dev graphify graph     dev update → dev graphify update
dev stats    → dev graphify stats     dev clean  → dev graphify clean
dev ui       → dev uipro init         dev memory → dev memory index
```

Run `dev help`, or `dev <tool> --help` for a tool's verbs.

## Install

Requires [Go 1.22+](https://go.dev/dl/). The installer (and every `make`
target) checks this first — when Go is missing you get an error plus the
install command for your OS (macOS, Debian/Ubuntu, Fedora, Arch, Alpine,
openSUSE, Windows).

```sh
# from a clone — checks Go, builds, installs to ~/.local/bin
./install.sh

# the same via make
make install

# or straight from the module path (needs Go already installed)
go install github.com/duclqDev99/custom-cli/cmd/dev@latest
```

Make sure the install dir is on your `PATH` — `install.sh` warns and prints
the exact line to add when it is not:

```sh
# ~/.local/bin (install.sh / make install) or ~/go/bin (go install)
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
| `dev setup`   | Configures all tools (`dev setup graphify` / `memory` / `uipro` for one).     |
| `dev doctor`  | Full health check of core deps + services (redis, postgres) + project.       |
| `dev graph`   | Runs `graphify extract .` the first time, `graphify update .` afterwards.     |
| `dev update`  | Always runs `graphify update .`.                                             |
| `dev sync`    | `graphify` → memory re-index → `git status` in one shot.                      |
| `dev memory`  | Runs `codebase-memory-mcp cli index_repository` for the current dir (incremental). Skips gracefully if the tool isn't installed. |
| `dev ai`      | Lists detected AI CLIs (Claude, Gemini, Codex, OpenAI), API keys, proxies.    |
| `dev stats`   | Node / edge / community counts from `graphify-out/graph.json`.                |
| `dev clean`   | Removes `graphify-out/` (`-f`/`--force` to skip the prompt).                  |
| `dev ui`      | Installs the UI/UX Pro Max skill for your assistant (`uipro init --ai claude`). |

## Setup: what `dev setup` automates

Each tool takes several manual steps; `dev setup` does them for you
(idempotent — safe to re-run).

### `dev setup graphify`

1. Checks that the `graphify` binary is present. If it isn't, it is installed
   automatically from PyPI (package [`graphifyy`](https://github.com/safishamsi/graphify),
   CLI `graphify`) — trying `uv tool install`, then `pipx`, then `pip`. The same
   auto-install runs on first use of any graph command (`dev graph`, `dev sync`, ...).
2. Installs the skill + hook for your agent: `graphify <platform> install`
   (writes the `CLAUDE.md` section + PreToolUse hook). Default platform is
   `claude`; override with `dev setup graphify --platform codex` (or `cursor`,
   `gemini`, ...).
3. Reports the LLM backend graphify will use (see below).

### graphify's LLM backend — and why no API key is needed

graphify only needs an LLM for **semantic extraction** (docs/images) and
**naming communities**. Plain code is pure local AST and needs **no key at all**.

When an LLM *is* needed, the backend is resolved in this order:

1. **An API key**, auto-detected by graphify: `ANTHROPIC_API_KEY` /
   `OPENAI_API_KEY` / `GEMINI_API_KEY` (or `GOOGLE_API_KEY`) /
   `DEEPSEEK_API_KEY` / `MOONSHOT_API_KEY`, or a self-hosted endpoint
   (`OPENAI_BASE_URL` / `ANTHROPIC_BASE_URL`), or `OLLAMA_BASE_URL`.
2. **`claude-cli`** — the locally-installed, already-authenticated `claude` CLI
   (Claude Code). This bills your **Pro/Max subscription**, not pay-as-you-go API
   credit, so **no API key is required**.

graphify never auto-selects `claude-cli` for a bare CLI run, so `dev` injects
`--backend claude-cli` for you when no API key is set but `claude` is on `PATH`.
That's why `dev graph` produces named communities on this machine with zero keys.
A `--backend` you pass yourself always wins:

```sh
dev graphify extract --backend ollama     # override
dev graph                                  # auto: claude-cli if no key, else your key
```

### `dev setup memory`

1. If `codebase-memory-mcp` isn't installed, runs the official one-line
   installer **only after you confirm** (or pass `--install` / `-y`):
   `curl -fsSL .../install.sh | bash`. The installer downloads a static binary
   and **auto-registers the MCP server** with Claude Code and other detected
   agents — no manual `~/.claude/.mcp.json` editing.
2. Enables incremental indexing: `codebase-memory-mcp config set auto_index true`.

No API keys, Redis, or Postgres are needed — codebase-memory-mcp is a single
static binary with embedded SQLite that runs 100% locally.

### `dev setup uipro`

1. Checks that the `uipro` binary is present. If it isn't, it is installed
   automatically from npm (package
   [`uipro-cli`](https://www.npmjs.com/package/uipro-cli), CLI `uipro`).
2. Installs the [UI/UX Pro Max](https://www.npmjs.com/package/uipro-cli) design
   skill for your assistant: `uipro init --ai claude`. Default platform is
   `claude`; override with `dev setup uipro --ai cursor` (or `windsurf`,
   `copilot`, ...). Extra flags like `--force` / `--offline` pass through.

```sh
dev setup            # all tools
dev setup graphify   # just graphify
dev setup memory -y  # install memory without the prompt
dev setup uipro      # UI/UX skill for Claude Code
```

### Graceful by design

Optional tools (`codebase-memory-mcp`, Docker, Redis, PostgreSQL) never crash a
command — they warn and continue. Only genuinely required tools (graphify, git,
node, python) count toward a failing `dev doctor`.

## Layout

```
cmd/dev/main.go              # registry + dispatch (workflow verbs, namespaces, aliases)
internal/core/              # Module interface, orchestrators (doctor/setup/sync/init), ai, env
internal/modules/graphify/  # graphify module — implements core.Module
internal/modules/memory/    # codebase-memory-mcp module — implements core.Module
internal/modules/uipro/     # uipro-cli (UI/UX skill) module — implements core.Module
internal/ui/                # colored output helpers
internal/tools/             # binary lookup + process running
```

### Adding a new tool

`dev` is built around a small `core.Module` interface, so integrating another
tool (Semgrep, Codex, n8n, ...) is a drop-in:

1. Create `internal/modules/<tool>/<tool>.go` implementing `core.Module`
   (`Name`, `Summary`, `Commands`, `Default`, `Doctor`, `Setup`, `Sync`).
2. Register it in `cmd/dev/main.go`:
   ```go
   mods := []core.Module{graphify.New(), memory.New(), yourtool.New()}
   ```

That's it — `dev doctor`, `dev setup`, `dev sync`, and `dev <tool> ...` all pick
it up automatically. No existing file needs to change.

## Development

```sh
make build     # → bin/dev
make run ARGS="doctor"
make vet
make fmt
```

Set `NO_COLOR=1` to disable ANSI colors.
