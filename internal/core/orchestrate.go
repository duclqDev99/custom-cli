package core

import (
	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// Doctor runs the environment checks plus every module's checks.
func Doctor(mods []Module) int {
	ui.Header("Environment")
	missing := RenderChecks(EnvChecks())

	for _, m := range mods {
		ui.Header(m.Name())
		missing += RenderChecks(m.Doctor())
	}

	ui.Header("Project")
	reportProject()

	ui.Header("Summary")
	if missing > 0 {
		ui.Fail("%d required check%s failing", missing, plural(missing))
		return 1
	}
	ui.OK("everything looks good")
	return 0
}

// Init is a friendly environment report plus next steps.
func Init(mods []Module) int {
	ui.Header("Initializing dev environment")
	missing := RenderChecks(EnvChecks())

	for _, m := range mods {
		ui.Header(m.Name())
		missing += RenderChecks(m.Doctor())
	}

	ui.Header("Project")
	reportProject()

	ui.Header("Next steps")
	ui.Info("configure the tools: %s", ui.Bold("dev setup"))
	if !FileExists(GraphJSON()) {
		ui.Info("build your first graph: %s", ui.Bold("dev graph"))
	}
	ui.Info("keep everything in sync: %s", ui.Bold("dev sync"))

	if missing > 0 {
		ui.Warn("%d required check%s failing — fix, then re-run %s",
			missing, plural(missing), ui.Bold("dev init"))
		return 1
	}
	ui.OK("ready to go")
	return 0
}

// Setup configures every module, or a single named one (`dev setup <name>`).
func Setup(mods []Module, args []string) int {
	target := ""
	rest := args
	if len(args) > 0 && !isFlag(args[0]) {
		target = args[0]
		rest = args[1:]
	}

	rc, ran := 0, false
	for _, m := range mods {
		if target == "" || target == m.Name() {
			if m.Setup(rest) != 0 {
				rc = 1
			}
			ran = true
		}
	}
	if !ran {
		ui.Error("unknown setup target %q (try one of: %s)", target, moduleNames(mods))
		return 1
	}
	return rc
}

// Sync runs each module's sync step, then prints git status.
func Sync(mods []Module) int {
	total := len(mods) + 1
	rc := 0
	for i, m := range mods {
		ui.Step("%d/%d %s", i+1, total, m.Name())
		if m.Sync() != 0 {
			rc = 1
		}
	}

	ui.Step("%d/%d git status", total, total)
	if InGitRepo() {
		_ = tools.Run("git", "status", "--short", "--branch")
	} else {
		ui.Warn("not a git repository")
	}

	if rc == 0 {
		ui.OK("sync complete")
	}
	return rc
}

func isFlag(s string) bool { return len(s) > 0 && s[0] == '-' }

func moduleNames(mods []Module) string {
	names := make([]string, len(mods))
	for i, m := range mods {
		names[i] = m.Name()
	}
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}
