// Package graphify is the dev module wrapping the graphify code-graph CLI.
package graphify

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/duclqDev99/custom-cli/internal/core"
	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

const (
	bin = "graphify"
	// pkg is the PyPI distribution name: the package is `graphifyy` (double y),
	// the installed CLI is `graphify`. https://github.com/safishamsi/graphify
	pkg  = "graphifyy"
	repo = "https://github.com/safishamsi/graphify"
)

// binPath is the resolved graphify executable, set by ensure(). It stays the
// bare name for a normal PATH install and becomes an absolute path when the
// binary lives somewhere PATH doesn't cover yet (fresh uv/pipx install).
var binPath = bin

// Module implements core.Module for graphify.
type Module struct{}

// New returns the graphify module.
func New() core.Module { return Module{} }

func (Module) Name() string    { return "graphify" }
func (Module) Summary() string { return "code knowledge graph (extract/update/stats)" }
func (Module) Default() string { return "graph" }

func (Module) Commands() []core.Command {
	return []core.Command{
		{Name: "graph", Desc: "extract first time, update afterwards", Run: cmdGraph},
		{Name: "extract", Desc: "full re-extraction", Run: func(a []string) int { return run("extract", withBackend(a)...) }},
		{Name: "update", Desc: "incremental update", Run: func(a []string) int { return run("update", a...) }},
		{Name: "label", Desc: "name communities via the LLM backend", Run: func(a []string) int { return run("label", withBackend(a)...) }},
		{Name: "stats", Desc: "node / edge / community counts", Run: cmdStats},
		{Name: "clean", Desc: "remove graphify-out/", Run: cmdClean},
	}
}

func (Module) Doctor() []core.Check {
	checks := core.ChecksFor([]tools.Tool{
		{Name: "graphify", Bin: bin, Hint: "auto-installs on first use, or: uv tool install " + pkg},
	})
	return append(checks, backendCheck())
}

// Sync updates the graph (extracting first if none exists yet).
func (Module) Sync() int {
	if !ensure() {
		ui.Warn("graphify unavailable — skipping")
		return 0
	}
	if core.FileExists(core.GraphJSON()) {
		return run("update")
	}
	ui.Warn("no graph yet — extracting first")
	return run("extract", withBackend(nil)...)
}

// Setup installs the graphify skill+hook for a platform and reports the backend.
func (Module) Setup(args []string) int {
	ui.Header("Setup · graphify")

	if !ensure() {
		return 1
	}
	ui.OK("graphify present")

	platform := core.Flag(args, "--platform", "claude")
	ui.Step("installing graphify skill + hook for %q", platform)
	// `graphify <platform> install` writes the skill, CLAUDE.md section and
	// PreToolUse hook. Fall back to the generic `install --platform` form.
	if err := tools.Run(binPath, platform, "install"); err != nil {
		if err2 := tools.Run(binPath, "install", "--platform", platform); err2 != nil {
			ui.Warn("could not auto-install skill — run %s manually", ui.Bold("graphify "+platform+" install"))
		} else {
			ui.OK("skill installed (%s)", platform)
		}
	} else {
		ui.OK("skill + hook installed (%s)", platform)
	}

	if _, detail, ok := chosenBackend(); ok {
		ui.OK("LLM backend ready: %s", detail)
	} else {
		ui.Warn("no LLM backend — %s", detail)
	}
	return 0
}

// resolve finds the graphify executable: PATH first, then ~/.local/bin —
// where uv tool and pipx place scripts that a not-yet-reloaded shell PATH
// may not cover.
func resolve() (string, bool) {
	if p, err := exec.LookPath(bin); err == nil {
		return p, true
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".local", "bin", bin)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, true
		}
	}
	return "", false
}

// ensure resolves graphify, auto-installing it on first use when missing.
// Install order mirrors the upstream README: uv tool → pipx → pip.
func ensure() bool {
	if p, ok := resolve(); ok {
		binPath = p
		return true
	}
	ui.Step("graphify not installed — installing %s from PyPI", pkg)
	installers := [][]string{
		{"uv", "tool", "install", pkg},
		{"pipx", "install", pkg},
		{"python3", "-m", "pip", "install", pkg},
		{"python3", "-m", "pip", "install", "--break-system-packages", pkg},
	}
	for _, c := range installers {
		if !tools.Exists(c[0]) {
			continue
		}
		ui.Info("trying: %s", ui.Gray(strings.Join(c, " ")))
		if err := tools.Run(c[0], c[1:]...); err != nil {
			continue
		}
		if p, ok := resolve(); ok {
			binPath = p
			ui.OK("graphify installed (%s)", p)
			if _, err := exec.LookPath(bin); err != nil {
				ui.Warn("%s is not on PATH — add it to run graphify directly", filepath.Dir(p))
			}
			return true
		}
	}
	ui.Fail("could not install graphify automatically")
	ui.Info("install manually: %s (see %s)", ui.Bold("uv tool install "+pkg), repo)
	return false
}

// run executes `graphify <sub> . [extra...]` and reports the outcome.
func run(sub string, extra ...string) int {
	if !ensure() {
		return 1
	}
	cmdArgs := append([]string{sub, "."}, extra...)
	if err := tools.Run(binPath, cmdArgs...); err != nil {
		ui.Error("graphify %s failed: %v", sub, err)
		return 1
	}
	ui.OK("graphify %s done", sub)
	return 0
}

// cmdGraph extracts a fresh graph or updates the existing one.
func cmdGraph(args []string) int {
	if !ensure() {
		return 1
	}
	if core.FileExists(core.GraphJSON()) {
		ui.Step("graph exists → updating")
		return run("update", args...)
	}
	ui.Step("no graph yet → extracting")
	return run("extract", withBackend(args)...)
}

// withBackend appends an explicit `--backend claude-cli` when no API-key backend
// is configured but the `claude` CLI is available — so community naming works
// via the Claude Code subscription with no API key. A user-supplied --backend
// always wins. claude-cli is the only backend graphify won't auto-detect, so
// it's the only one we ever inject.
func withBackend(args []string) []string {
	for _, a := range args {
		if a == "--backend" || strings.HasPrefix(a, "--backend=") {
			return args // respect explicit override
		}
	}
	extra := backendArgs()
	if len(extra) > 0 {
		ui.Info("backend: %s", ui.Gray("claude-cli (Claude Code subscription, no API key)"))
	}
	return append(append([]string{}, args...), extra...)
}

// backendArgs returns the --backend flag to inject, or nil when graphify can
// resolve a backend itself (API key set) or none is available.
func backendArgs() []string {
	if name, _, ok := chosenBackend(); ok && name == "claude-cli" {
		return []string{"--backend", "claude-cli"}
	}
	return nil
}

// chosenBackend mirrors graphify's resolution: an API key wins (graphify
// auto-detects it), else a self-hosted/ollama endpoint, else the local `claude`
// CLI (claude-cli — subscription, no key). Returns ok=false when only code-only
// AST extraction is possible.
func chosenBackend() (name, detail string, ok bool) {
	for _, b := range []struct{ name, env string }{
		{"gemini", "GEMINI_API_KEY"},
		{"gemini", "GOOGLE_API_KEY"},
		{"kimi", "MOONSHOT_API_KEY"},
		{"claude", "ANTHROPIC_API_KEY"},
		{"openai", "OPENAI_API_KEY"},
		{"deepseek", "DEEPSEEK_API_KEY"},
	} {
		if os.Getenv(b.env) != "" {
			return b.name, b.name + " (" + b.env + ")", true
		}
	}
	if os.Getenv("ANTHROPIC_BASE_URL") != "" || os.Getenv("OPENAI_BASE_URL") != "" {
		return "self-hosted", "self-hosted (BASE_URL set)", true
	}
	if os.Getenv("OLLAMA_BASE_URL") != "" {
		return "ollama", "ollama (OLLAMA_BASE_URL)", true
	}
	if tools.Exists("claude") {
		return "claude-cli", "claude-cli (Claude Code subscription, no API key)", true
	}
	return "", "code-only AST works; set an API key or install Claude Code for semantic naming", false
}

// backendCheck reports the LLM backend graphify will use, for `dev doctor`.
func backendCheck() core.Check {
	_, detail, ok := chosenBackend()
	return core.Check{Name: "LLM backend", OK: ok, Optional: true, Detail: detail}
}

// cmdStats prints high-level statistics about the knowledge graph.
func cmdStats([]string) int {
	path := core.GraphJSON()
	if !core.FileExists(path) {
		ui.Warn("no graph found — run %s first", ui.Bold("dev graph"))
		return 1
	}

	data, err := os.ReadFile(path)
	if err != nil {
		ui.Error("read %s: %v", path, err)
		return 1
	}

	var g struct {
		Nodes []json.RawMessage `json:"nodes"`
		Edges []json.RawMessage `json:"edges"`
		Links []json.RawMessage `json:"links"`
	}
	if err := json.Unmarshal(data, &g); err != nil {
		ui.Error("parse %s: %v", path, err)
		return 1
	}

	edges := len(g.Edges)
	if edges == 0 {
		edges = len(g.Links)
	}

	ui.Header("Knowledge graph")
	ui.Info("file: %s (%s)", path, humanSize(len(data)))
	ui.OK("%d nodes", len(g.Nodes))
	ui.OK("%d edges", edges)
	if c := countCommunities(g.Nodes); c > 0 {
		ui.OK("%d communities", c)
	}
	return 0
}

// cmdClean removes the generated graphify-out/ directory.
func cmdClean(args []string) int {
	dir := core.GraphDir
	if !core.FileExists(dir) {
		ui.Info("nothing to clean (%s/ not found)", dir)
		return 0
	}

	if !core.HasFlag(args, "-f", "--force") {
		fmt.Printf("Remove %s/ ? [y/N] ", dir)
		ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(ans)); a != "y" && a != "yes" {
			ui.Info("aborted")
			return 0
		}
	}

	if err := os.RemoveAll(dir); err != nil {
		ui.Error("failed to remove %s/: %v", dir, err)
		return 1
	}
	ui.OK("removed %s/", dir)
	return 0
}

// countCommunities counts distinct community/cluster ids across nodes.
func countCommunities(nodes []json.RawMessage) int {
	seen := map[string]bool{}
	for _, n := range nodes {
		var m map[string]any
		if json.Unmarshal(n, &m) != nil {
			continue
		}
		for _, key := range []string{"community", "cluster", "communityId", "community_id"} {
			if v, ok := m[key]; ok && v != nil {
				seen[fmt.Sprint(v)] = true
				break
			}
		}
	}
	return len(seen)
}

// humanSize formats a byte count as a human-readable string.
func humanSize(n int) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := int64(n) / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
