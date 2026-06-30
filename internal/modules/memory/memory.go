// Package memory is the dev module wrapping codebase-memory-mcp, a local code
// intelligence index. The tool is optional — every entry point degrades
// gracefully when the binary is absent.
package memory

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

const bin = "codebase-memory-mcp"

// installCmd is the official one-line installer. It downloads a static binary
// and auto-registers the MCP server with Claude Code and other detected agents.
const installCmd = `curl -fsSL https://raw.githubusercontent.com/DeusData/codebase-memory-mcp/main/install.sh | bash`

// Module implements core.Module for codebase-memory-mcp.
type Module struct{}

// New returns the memory module.
func New() core.Module { return Module{} }

func (Module) Name() string    { return "memory" }
func (Module) Summary() string { return "local code memory / MCP index" }
func (Module) Default() string { return "index" }

func (Module) Commands() []core.Command {
	return []core.Command{
		{Name: "index", Desc: "index the current repo (incremental)", Run: func([]string) int { return index() }},
		{Name: "status", Desc: "show index status", Run: func([]string) int { return status() }},
	}
}

func (Module) Doctor() []core.Check {
	return core.ChecksFor([]tools.Tool{
		{Name: "codebase-memory-mcp", Bin: bin, Optional: true, Hint: "run `dev setup memory`"},
	})
}

// Sync re-indexes the repo. index_repository is incremental, so this doubles as
// the reindex step.
func (Module) Sync() int { return index() }

// Setup installs (with consent) and configures codebase-memory-mcp.
func (Module) Setup(args []string) int {
	ui.Header("Setup · codebase-memory-mcp")

	if tools.Exists(bin) {
		ui.OK("%s already installed", bin)
	} else {
		ui.Warn("%s not installed", bin)
		if !confirmInstall(args) {
			ui.Info("install it manually with:")
			fmt.Printf("    %s\n", ui.Bold(installCmd))
			return 1
		}
		ui.Step("running the official installer (downloads a binary, registers MCP)")
		if err := tools.Run("bash", "-c", installCmd); err != nil {
			ui.Error("installer failed: %v", err)
			return 1
		}
		if !tools.Exists(bin) {
			ui.Warn("installer finished but %s isn't on PATH yet — open a new shell, then re-run", bin)
			return 1
		}
		ui.OK("%s installed", bin)
	}

	ui.Step("enabling auto_index")
	if err := tools.Run(bin, "config", "set", "auto_index", "true"); err != nil {
		ui.Warn("could not set auto_index — set it later with %s", ui.Bold(bin+" config set auto_index true"))
	} else {
		ui.OK("auto_index enabled")
	}

	ui.Info("MCP server is auto-registered with detected agents (Claude Code, Codex, ...)")
	ui.Info("index a project with %s", ui.Bold("dev memory"))
	return 0
}

// index runs `codebase-memory-mcp cli index_repository` for the working dir.
func index() int {
	if !tools.Exists(bin) {
		ui.Warn("%s not found — skipping memory step", bin)
		ui.Info("%s", ui.Gray("run `dev setup memory` to install it"))
		return 0
	}

	wd, err := os.Getwd()
	if err != nil {
		ui.Error("cannot determine working directory: %v", err)
		return 1
	}

	payload, _ := json.Marshal(map[string]string{"repo_path": wd})
	ui.Step("%s cli index_repository", bin)
	if err := tools.Run(bin, "cli", "index_repository", string(payload)); err != nil {
		ui.Error("memory index failed: %v", err)
		return 1
	}
	ui.OK("memory indexed")
	return 0
}

// status runs `codebase-memory-mcp cli index_status`.
func status() int {
	if !tools.Exists(bin) {
		ui.Warn("%s not installed — run %s", bin, ui.Bold("dev setup memory"))
		return 1
	}
	ui.Step("%s cli index_status", bin)
	if err := tools.Run(bin, "cli", "index_status"); err != nil {
		ui.Error("memory status failed: %v", err)
		return 1
	}
	return 0
}

// confirmInstall returns true if the user passed -y/--install or confirms the
// prompt. Network installs are opt-in, never silent.
func confirmInstall(args []string) bool {
	if core.HasFlag(args, "--install", "-y") {
		return true
	}
	fmt.Print("Run the official installer now? [y/N] ")
	ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	ans = strings.ToLower(strings.TrimSpace(ans))
	return ans == "y" || ans == "yes"
}
