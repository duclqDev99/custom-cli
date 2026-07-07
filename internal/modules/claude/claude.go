// Package claude is the dev module for Claude Code environment tweaks. Its first
// job is the status line: a per-prompt readout of model · context usage · token
// count · session cost, rendered by a small Node script that Claude Code invokes
// with session JSON on stdin.
// Docs: https://docs.claude.com/en/docs/claude-code/statusline
package claude

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/duclqDev99/custom-cli/internal/core"
	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// statuslineJS is the Node script Claude Code runs for each status line render.
// It ships with the binary so `dev claude statusline` works on any machine.
//
//go:embed statusline.js
var statuslineJS []byte

// scriptName is the file the script is written to under ~/.claude.
const scriptName = "statusline.js"

// Module implements core.Module for Claude Code tweaks.
type Module struct{}

// New returns the claude module.
func New() core.Module { return Module{} }

func (Module) Name() string    { return "claude" }
func (Module) Summary() string { return "Claude Code status line (context · cost · quota)" }
func (Module) Default() string { return "statusline" }

func (Module) Commands() []core.Command {
	return []core.Command{
		{Name: "statusline", Desc: "install the status line (context · token · cost · 5h/7d quota); --remove to undo", Run: cmdStatusline},
	}
}

func (Module) Doctor() []core.Check {
	checks := core.ChecksFor([]tools.Tool{
		{Name: "Node.js", Bin: "node", VersionArg: []string{"--version"}, Optional: true, Hint: "needed to render the status line"},
	})

	dir, err := claudeDir()
	if err != nil {
		return append(checks, core.Check{Name: "status line", OK: false, Optional: true, Detail: err.Error()})
	}
	const hint = "run `dev claude statusline`"
	scriptPath := filepath.Join(dir, scriptName)
	checks = append(checks, core.Check{
		Name: "status line script", OK: core.FileExists(scriptPath), Optional: true,
		Detail: detailOr(core.FileExists(scriptPath), scriptPath, hint),
	})
	configured := settingsHasStatusLine(dir)
	checks = append(checks, core.Check{
		Name: "status line configured", OK: configured, Optional: true,
		Detail: detailOr(configured, filepath.Join(dir, "settings.json"), hint),
	})
	return checks
}

// Sync is a no-op: the status line is a global Claude Code setting, not per-project.
func (Module) Sync() int {
	ui.Info("%s", ui.Gray("status line is a global Claude Code setting — nothing to sync"))
	return 0
}

// Setup installs the status line as part of `dev setup [claude]`.
func (Module) Setup(args []string) int {
	ui.Header("Setup · claude (status line)")
	return cmdStatusline(args)
}

// cmdStatusline installs (or, with --remove, uninstalls) the Claude Code status
// line: it writes the Node script under ~/.claude and points settings.json at it.
func cmdStatusline(args []string) int {
	dir, err := claudeDir()
	if err != nil {
		ui.Error("cannot locate ~/.claude: %v", err)
		return 1
	}
	if core.HasFlag(args, "--remove", "--uninstall") {
		return removeStatusline(dir)
	}
	return installStatusline(dir)
}

// installStatusline writes the script and wires up settings.json.
func installStatusline(dir string) int {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		ui.Error("cannot create %s: %v", dir, err)
		return 1
	}

	// 1) write the Node script (backing up any existing, differing copy).
	scriptPath := filepath.Join(dir, scriptName)
	switch existing, err := os.ReadFile(scriptPath); {
	case err == nil && string(existing) == string(statuslineJS):
		ui.OK("script already up to date (%s)", ui.Gray(scriptPath))
	case err == nil:
		if bak, err := backup(scriptPath, existing); err == nil {
			ui.Info("backed up your script → %s", ui.Gray(bak))
		}
		fallthrough
	default:
		if err := os.WriteFile(scriptPath, statuslineJS, 0o755); err != nil {
			ui.Error("cannot write %s: %v", scriptPath, err)
			return 1
		}
		ui.OK("wrote status line script (%s)", ui.Gray(scriptPath))
	}

	// 2) point settings.json at it, via an absolute node path when resolvable.
	node := nodePath()
	command := node + " " + scriptPath
	if err := patchSettings(dir, map[string]any{
		"type":    "command",
		"command": command,
		"padding": 0,
	}); err != nil {
		ui.Error("cannot update settings.json: %v", err)
		return 1
	}
	ui.OK("configured settings.json → %s", ui.Gray(command))

	if !tools.Exists("node") {
		ui.Warn("node isn't on PATH — install Node.js or the status line won't render")
	}
	ui.Info("open a new Claude Code prompt to see: model · context · token · cost")
	return 0
}

// removeStatusline deletes the statusLine block from settings.json and removes
// the installed script.
func removeStatusline(dir string) int {
	if err := patchSettings(dir, nil); err != nil {
		ui.Error("cannot update settings.json: %v", err)
		return 1
	}
	ui.OK("removed statusLine from settings.json")

	scriptPath := filepath.Join(dir, scriptName)
	if core.FileExists(scriptPath) {
		if err := os.Remove(scriptPath); err != nil {
			ui.Warn("could not remove %s: %v", scriptPath, err)
		} else {
			ui.OK("removed %s", ui.Gray(scriptPath))
		}
	}
	return 0
}

// patchSettings reads ~/.claude/settings.json, sets (or, when block is nil,
// deletes) the "statusLine" key, and writes it back — preserving every other
// setting. The prior file is backed up before it is overwritten.
func patchSettings(dir string, block map[string]any) error {
	path := filepath.Join(dir, "settings.json")
	settings := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("%s is not valid JSON (leaving it untouched): %w", path, err)
		}
		if _, err := backup(path, data); err != nil {
			return fmt.Errorf("could not back up %s: %w", path, err)
		}
	}

	if block == nil {
		delete(settings, "statusLine")
	} else {
		settings["statusLine"] = block
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// settingsHasStatusLine reports whether settings.json defines a statusLine.
func settingsHasStatusLine(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		return false
	}
	var s struct {
		StatusLine json.RawMessage `json:"statusLine"`
	}
	return json.Unmarshal(data, &s) == nil && len(s.StatusLine) > 0
}

// backup writes data next to path as "<path>.bak" so an edit is recoverable.
func backup(path string, data []byte) (string, error) {
	bak := path + ".bak"
	return bak, os.WriteFile(bak, data, 0o644)
}

// nodePath resolves an absolute node path so the status line works regardless of
// the shell PATH Claude Code inherits, falling back to the bare name.
func nodePath() string {
	if p, err := exec.LookPath("node"); err == nil {
		return p
	}
	return "node"
}

// claudeDir returns ~/.claude, where Claude Code keeps its user settings.
func claudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// detailOr returns path as a check's detail when ok, else the remediation hint.
func detailOr(ok bool, path, hint string) string {
	if ok {
		return path
	}
	return hint
}
