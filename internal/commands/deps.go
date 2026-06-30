package commands

import (
	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// coreDeps are the tools dev orchestrates or expects to find.
var coreDeps = []tools.Tool{
	{Name: "Graphify", Bin: "graphify", Hint: "https://github.com/ — install the graphify CLI"},
	{Name: "Codebase Memory MCP", Bin: "codebase-memory-mcp", Optional: true, Hint: "run `dev setup memory`"},
	{Name: "Git", Bin: "git", VersionArg: []string{"--version"}},
	{Name: "Node.js", Bin: "node", VersionArg: []string{"--version"}},
	{Name: "Python", Bin: "python3", VersionArg: []string{"--version"}},
	{Name: "Docker", Bin: "docker", VersionArg: []string{"--version"}, Optional: true, Hint: "install Docker Desktop"},
	{Name: "Claude Code", Bin: "claude", VersionArg: []string{"--version"}, Optional: true, Hint: "npm i -g @anthropic-ai/claude-code"},
}

// serviceDeps are background services doctor reports on (all optional).
var serviceDeps = []tools.Tool{
	{Name: "Redis", Bin: "redis-cli", VersionArg: []string{"--version"}, Optional: true, Hint: "brew install redis"},
	{Name: "PostgreSQL", Bin: "psql", VersionArg: []string{"--version"}, Optional: true, Hint: "brew install postgresql"},
}

// checkDeps prints the status of each tool and returns the count of missing
// required (non-optional) tools.
func checkDeps(deps []tools.Tool) (missingRequired int) {
	for _, t := range deps {
		path, ok := t.Found()
		if ok {
			detail := t.Version()
			if detail == "" {
				detail = path
			}
			ui.OK("%s %s", t.Name, ui.Gray(detail))
			continue
		}

		hint := ""
		if t.Hint != "" {
			hint = ui.Gray(" (" + t.Hint + ")")
		}
		if t.Optional {
			ui.Warn("%s%s%s", t.Name, ui.Gray(" — not installed"), hint)
		} else {
			ui.Fail("%s%s%s", t.Name, ui.Gray(" — missing"), hint)
			missingRequired++
		}
	}
	return missingRequired
}

// pluralize returns "s" when n != 1.
func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
