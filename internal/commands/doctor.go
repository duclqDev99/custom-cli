package commands

import (
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// Doctor reports the health of every dependency and the current project.
func Doctor(args []string) int {
	ui.Header("Core dependencies")
	missing := checkDeps(coreDeps)

	ui.Header("Services")
	checkDeps(serviceDeps)

	ui.Header("Project")
	reportProject()

	ui.Header("Summary")
	if missing > 0 {
		ui.Fail("%d required tool%s missing", missing, pluralize(missing))
		return 1
	}
	ui.OK("everything looks good")
	return 0
}

// Init checks the environment and points the user at first steps.
func Init(args []string) int {
	ui.Header("Initializing dev environment")
	missing := checkDeps(coreDeps)

	ui.Header("Project")
	reportProject()

	ui.Header("Next steps")
	ui.Info("configure the tools: %s", ui.Bold("dev setup"))
	if !fileExists(graphJSON()) {
		ui.Info("build your first graph: %s", ui.Bold("dev graph"))
	}
	ui.Info("keep everything in sync: %s", ui.Bold("dev sync"))

	if missing > 0 {
		ui.Warn("%d required tool%s missing — install, then re-run %s",
			missing, pluralize(missing), ui.Bold("dev init"))
		return 1
	}
	ui.OK("ready to go")
	return 0
}
