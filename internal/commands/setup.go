package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// memoryInstallCmd is the official one-line installer for codebase-memory-mcp.
// It downloads a binary and auto-registers the MCP server with Claude Code and
// other detected agents.
const memoryInstallCmd = `curl -fsSL https://raw.githubusercontent.com/DeusData/codebase-memory-mcp/main/install.sh | bash`

// Setup configures graphify and/or codebase-memory-mcp. With no target it does
// both; "graphify" or "memory" targets just one. Every step is idempotent and
// safe to re-run.
func Setup(args []string) int {
	target := "all"
	rest := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		target = args[0]
		rest = args[1:]
	}

	switch target {
	case "all":
		g := setupGraphify(rest)
		m := setupMemory(rest)
		if g != 0 || m != 0 {
			return 1
		}
		return 0
	case "graphify":
		return setupGraphify(rest)
	case "memory":
		return setupMemory(rest)
	default:
		ui.Error("unknown setup target %q (use: graphify | memory)", target)
		return 1
	}
}

// setupGraphify installs the graphify skill+hook for a platform and reports the
// LLM backend situation.
func setupGraphify(args []string) int {
	ui.Header("Setup · graphify")

	if !tools.Exists("graphify") {
		ui.Fail("graphify not installed")
		ui.Info("install the graphify CLI first, then re-run %s", ui.Bold("dev setup graphify"))
		return 1
	}
	ui.OK("graphify present")

	platform := flagValue(args, "--platform", "claude")
	ui.Step("installing graphify skill + hook for %q", platform)
	// Per-platform form (e.g. `graphify claude install`) writes the skill,
	// CLAUDE.md section and PreToolUse hook. Fall back to the generic form.
	if err := tools.Run("graphify", platform, "install"); err != nil {
		if err2 := tools.Run("graphify", "install", "--platform", platform); err2 != nil {
			ui.Warn("could not auto-install skill — run %s manually", ui.Bold("graphify "+platform+" install"))
		} else {
			ui.OK("skill installed (%s)", platform)
		}
	} else {
		ui.OK("skill + hook installed (%s)", platform)
	}

	reportBackend()
	return 0
}

// reportBackend tells the user which LLM backend graphify will pick up. graphify
// auto-detects from whichever API key is set.
func reportBackend() {
	backends := []struct{ name, env string }{
		{"claude", "ANTHROPIC_API_KEY"},
		{"openai", "OPENAI_API_KEY"},
		{"gemini", "GEMINI_API_KEY"},
		{"deepseek", "DEEPSEEK_API_KEY"},
		{"kimi", "MOONSHOT_API_KEY"},
	}
	for _, b := range backends {
		if os.Getenv(b.env) != "" {
			ui.OK("LLM backend ready: %s %s", b.name, ui.Gray("("+b.env+" set)"))
			return
		}
	}
	if os.Getenv("OPENAI_BASE_URL") != "" || os.Getenv("ANTHROPIC_BASE_URL") != "" {
		ui.OK("self-hosted LLM endpoint configured %s", ui.Gray("(BASE_URL set)"))
		return
	}
	if tools.Exists("ollama") {
		ui.Info("no API key set, but ollama found — use %s", ui.Bold("graphify ... --backend ollama"))
		return
	}
	ui.Warn("no LLM backend detected")
	ui.Info("%s", ui.Gray("set ANTHROPIC_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY for semantic extraction & community naming"))
}

// setupMemory installs (with consent) and configures codebase-memory-mcp.
func setupMemory(args []string) int {
	ui.Header("Setup · codebase-memory-mcp")

	if memoryAvailable() {
		ui.OK("%s already installed", memoryBin)
	} else {
		ui.Warn("%s not installed", memoryBin)
		if !confirmInstall(args) {
			ui.Info("install it manually with:")
			fmt.Printf("    %s\n", ui.Bold(memoryInstallCmd))
			return 1
		}
		ui.Step("running the official installer (downloads a binary, registers MCP)")
		if err := tools.Run("bash", "-c", memoryInstallCmd); err != nil {
			ui.Error("installer failed: %v", err)
			return 1
		}
		if !memoryAvailable() {
			ui.Warn("installer finished but %s isn't on PATH yet — open a new shell, then re-run", memoryBin)
			return 1
		}
		ui.OK("%s installed", memoryBin)
	}

	ui.Step("enabling auto_index")
	if err := tools.Run(memoryBin, "config", "set", "auto_index", "true"); err != nil {
		ui.Warn("could not set auto_index — set it later with %s", ui.Bold("codebase-memory-mcp config set auto_index true"))
	} else {
		ui.OK("auto_index enabled")
	}

	ui.Info("MCP server is auto-registered with detected agents (Claude Code, Codex, ...)")
	ui.Info("index a project with %s", ui.Bold("dev memory"))
	return 0
}

// confirmInstall returns true if the user passed -y/--install or answers yes to
// the prompt. Network installs are opt-in, never silent.
func confirmInstall(args []string) bool {
	for _, a := range args {
		if a == "--install" || a == "-y" {
			return true
		}
	}
	fmt.Print("Run the official installer now? [y/N] ")
	ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	ans = strings.ToLower(strings.TrimSpace(ans))
	return ans == "y" || ans == "yes"
}

// flagValue returns the value of --name <v> or --name=v, or def if absent.
func flagValue(args []string, name, def string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"=")
		}
	}
	return def
}
