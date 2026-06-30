package core

import (
	"os"

	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// AI reports installed AI CLIs, configured API keys, and proxy overrides. It is
// a cross-cutting utility, not tied to any single module.
func AI(args []string) int {
	ui.Header("AI CLIs")
	RenderChecks(ChecksFor([]tools.Tool{
		{Name: "Claude Code", Bin: "claude", VersionArg: []string{"--version"}, Optional: true},
		{Name: "Gemini CLI", Bin: "gemini", Optional: true},
		{Name: "Codex CLI", Bin: "codex", Optional: true},
		{Name: "OpenAI CLI", Bin: "openai", Optional: true},
	}))

	ui.Header("API keys")
	for _, k := range []struct{ name, env string }{
		{"Anthropic", "ANTHROPIC_API_KEY"},
		{"OpenAI", "OPENAI_API_KEY"},
		{"Google Gemini", "GEMINI_API_KEY"},
		{"Google", "GOOGLE_API_KEY"},
	} {
		if v := os.Getenv(k.env); v != "" {
			ui.OK("%s %s", k.name, ui.Gray("("+mask(v)+")"))
		} else {
			ui.Warn("%s %s", k.name, ui.Gray("("+k.env+" not set)"))
		}
	}

	ui.Header("Proxy / base URLs")
	found := false
	for _, p := range []string{
		"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY",
		"ANTHROPIC_BASE_URL", "OPENAI_BASE_URL", "OPENAI_API_BASE",
	} {
		if v := os.Getenv(p); v != "" {
			ui.Info("%s = %s", p, v)
			found = true
		}
	}
	if !found {
		ui.Info("%s", ui.Gray("no proxy / base-url overrides set"))
	}
	return 0
}

// mask hides the middle of a secret, keeping a recognizable prefix/suffix.
func mask(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "…" + s[len(s)-4:]
}
