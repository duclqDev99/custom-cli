package commands

import (
	"os"

	"github.com/duclq/dev/internal/tools"
	"github.com/duclq/dev/internal/ui"
)

// AI reports installed AI CLIs, configured API keys, and proxy overrides.
func AI(args []string) int {
	ui.Header("AI CLIs")
	checkDeps([]tools.Tool{
		{Name: "Claude Code", Bin: "claude", VersionArg: []string{"--version"}, Optional: true},
		{Name: "Gemini CLI", Bin: "gemini", Optional: true},
		{Name: "Codex CLI", Bin: "codex", Optional: true},
		{Name: "OpenAI CLI", Bin: "openai", Optional: true},
	})

	ui.Header("API keys")
	keys := []struct{ name, env string }{
		{"Anthropic", "ANTHROPIC_API_KEY"},
		{"OpenAI", "OPENAI_API_KEY"},
		{"Google Gemini", "GEMINI_API_KEY"},
		{"Google", "GOOGLE_API_KEY"},
	}
	for _, k := range keys {
		if v := os.Getenv(k.env); v != "" {
			ui.OK("%s %s", k.name, ui.Gray("("+mask(v)+")"))
		} else {
			ui.Warn("%s %s", k.name, ui.Gray("("+k.env+" not set)"))
		}
	}

	ui.Header("Proxy / base URLs")
	overrides := []string{
		"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY",
		"ANTHROPIC_BASE_URL", "OPENAI_BASE_URL", "OPENAI_API_BASE",
	}
	found := false
	for _, p := range overrides {
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
