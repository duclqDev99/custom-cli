// Package uipro is the dev module wrapping uipro-cli, which installs the
// UI/UX Pro Max design skill for AI coding assistants.
package uipro

import (
	"github.com/duclqDev99/custom-cli/internal/core"
	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

const (
	bin = "uipro"
	// pkg is the npm package name: the package is `uipro-cli`, the installed
	// CLI is `uipro`. https://www.npmjs.com/package/uipro-cli
	pkg = "uipro-cli"
)

// Module implements core.Module for uipro-cli.
type Module struct{}

// New returns the uipro module.
func New() core.Module { return Module{} }

func (Module) Name() string    { return "uipro" }
func (Module) Summary() string { return "UI/UX design skill for AI assistants (init/update)" }
func (Module) Default() string { return "init" }

func (Module) Commands() []core.Command {
	return []core.Command{
		{Name: "init", Desc: "install the UX/UI skill for an assistant (--ai claude)", Run: cmdInit},
		{Name: "update", Desc: "upgrade the skill to the newest release", Run: func(a []string) int { return run("update", a...) }},
		{Name: "versions", Desc: "list available skill releases", Run: func(a []string) int { return run("versions", a...) }},
	}
}

func (Module) Doctor() []core.Check {
	return core.ChecksFor([]tools.Tool{
		{Name: "uipro", Bin: bin, Optional: true, Hint: "auto-installs on first use, or: npm install -g " + pkg},
	})
}

// Sync is a no-op: the skill is installed per assistant, not per project.
func (Module) Sync() int {
	ui.Info("%s", ui.Gray("skill install is global — nothing to sync"))
	return 0
}

// Setup installs uipro-cli (npm) and the UX/UI skill for a platform.
func (Module) Setup(args []string) int {
	ui.Header("Setup · uipro (UI/UX skill)")
	if !ensure() {
		return 1
	}
	ui.OK("uipro present")
	return cmdInit(args)
}

// ensure resolves uipro, auto-installing it via npm on first use when missing.
func ensure() bool {
	if tools.Exists(bin) {
		return true
	}
	ui.Step("uipro not installed — installing %s from npm", pkg)
	if !tools.Exists("npm") {
		ui.Fail("npm not found — install Node.js first")
		ui.Info("then: %s", ui.Bold("npm install -g "+pkg))
		return false
	}
	if err := tools.Run("npm", "install", "-g", pkg); err != nil {
		ui.Fail("npm install -g %s failed: %v", pkg, err)
		ui.Info("install manually: %s", ui.Bold("npm install -g "+pkg))
		return false
	}
	if !tools.Exists(bin) {
		ui.Warn("installed, but %s isn't on PATH yet — open a new shell, then re-run", bin)
		return false
	}
	ui.OK("uipro installed")
	return true
}

// cmdInit installs the UX/UI skill for an assistant, defaulting to claude when
// no --ai was given. Other flags (--force, --offline, ...) pass through.
func cmdInit(args []string) int {
	if !ensure() {
		return 1
	}
	if core.Flag(args, "--ai", "") == "" {
		args = append([]string{"--ai", "claude"}, args...)
	}
	ai := core.Flag(args, "--ai", "claude")
	ui.Step("installing UI/UX skill for %q", ai)
	if err := tools.Run(bin, append([]string{"init"}, args...)...); err != nil {
		ui.Error("uipro init failed: %v", err)
		return 1
	}
	ui.OK("UI/UX skill installed (%s)", ai)
	return 0
}

// run executes `uipro <sub> [extra...]`, installing uipro first if needed.
func run(sub string, extra ...string) int {
	if !ensure() {
		return 1
	}
	if err := tools.Run(bin, append([]string{sub}, extra...)...); err != nil {
		ui.Error("uipro %s failed: %v", sub, err)
		return 1
	}
	ui.OK("uipro %s done", sub)
	return 0
}
