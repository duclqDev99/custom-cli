// Package graphify is the dev module wrapping the graphify code-graph CLI.
package graphify

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/duclqDev99/custom-cli/internal/core"
	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

const bin = "graphify"

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
		{Name: "extract", Desc: "full re-extraction", Run: func([]string) int { return run("extract") }},
		{Name: "update", Desc: "incremental update", Run: func([]string) int { return run("update") }},
		{Name: "stats", Desc: "node / edge / community counts", Run: cmdStats},
		{Name: "clean", Desc: "remove graphify-out/", Run: cmdClean},
	}
}

func (Module) Doctor() []core.Check {
	checks := core.ChecksFor([]tools.Tool{
		{Name: "graphify", Bin: bin, Hint: "install the graphify CLI"},
	})
	return append(checks, backendCheck())
}

// Sync updates the graph (extracting first if none exists yet).
func (Module) Sync() int {
	if !tools.Exists(bin) {
		ui.Warn("graphify not installed — skipping")
		return 0
	}
	if core.FileExists(core.GraphJSON()) {
		return run("update")
	}
	ui.Warn("no graph yet — extracting first")
	return run("extract")
}

// Setup installs the graphify skill+hook for a platform and reports the backend.
func (Module) Setup(args []string) int {
	ui.Header("Setup · graphify")

	if !tools.Exists(bin) {
		ui.Fail("graphify not installed")
		ui.Info("install the graphify CLI first, then re-run %s", ui.Bold("dev setup graphify"))
		return 1
	}
	ui.OK("graphify present")

	platform := core.Flag(args, "--platform", "claude")
	ui.Step("installing graphify skill + hook for %q", platform)
	// `graphify <platform> install` writes the skill, CLAUDE.md section and
	// PreToolUse hook. Fall back to the generic `install --platform` form.
	if err := tools.Run(bin, platform, "install"); err != nil {
		if err2 := tools.Run(bin, "install", "--platform", platform); err2 != nil {
			ui.Warn("could not auto-install skill — run %s manually", ui.Bold("graphify "+platform+" install"))
		} else {
			ui.OK("skill installed (%s)", platform)
		}
	} else {
		ui.OK("skill + hook installed (%s)", platform)
	}

	if c := backendCheck(); c.OK {
		ui.OK("LLM backend ready: %s", c.Detail)
	} else {
		ui.Warn("no LLM backend — %s", c.Detail)
	}
	return 0
}

// run executes `graphify <sub> .` and reports the outcome.
func run(sub string) int {
	if !tools.Exists(bin) {
		ui.Error("graphify is not installed")
		return 1
	}
	if err := tools.Run(bin, sub, "."); err != nil {
		ui.Error("graphify %s failed: %v", sub, err)
		return 1
	}
	ui.OK("graphify %s done", sub)
	return 0
}

// cmdGraph extracts a fresh graph or updates the existing one.
func cmdGraph([]string) int {
	if !tools.Exists(bin) {
		ui.Error("graphify is not installed")
		return 1
	}
	if core.FileExists(core.GraphJSON()) {
		ui.Step("graph exists → updating")
		return run("update")
	}
	ui.Step("no graph yet → extracting")
	return run("extract")
}

// backendCheck reports which LLM backend graphify will auto-detect.
func backendCheck() core.Check {
	for _, b := range []struct{ name, env string }{
		{"claude", "ANTHROPIC_API_KEY"},
		{"openai", "OPENAI_API_KEY"},
		{"gemini", "GEMINI_API_KEY"},
		{"deepseek", "DEEPSEEK_API_KEY"},
		{"kimi", "MOONSHOT_API_KEY"},
	} {
		if os.Getenv(b.env) != "" {
			return core.Check{Name: "LLM backend", OK: true, Detail: b.name + " (" + b.env + ")"}
		}
	}
	if os.Getenv("OPENAI_BASE_URL") != "" || os.Getenv("ANTHROPIC_BASE_URL") != "" {
		return core.Check{Name: "LLM backend", OK: true, Detail: "self-hosted (BASE_URL set)"}
	}
	if tools.Exists("ollama") {
		return core.Check{Name: "LLM backend", OK: true, Detail: "ollama (local)"}
	}
	return core.Check{
		Name:     "LLM backend",
		Optional: true,
		Detail:   "set ANTHROPIC_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY",
	}
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
