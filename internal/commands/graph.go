package commands

import (
	"github.com/duclq/dev/internal/tools"
	"github.com/duclq/dev/internal/ui"
)

// Graph extracts a fresh graph or updates the existing one.
func Graph(args []string) int {
	if !tools.Exists("graphify") {
		ui.Error("graphify is not installed")
		return 1
	}
	if fileExists(graphJSON()) {
		ui.Step("graph exists → updating")
		return runGraphify("update")
	}
	ui.Step("no graph yet → extracting")
	return runGraphify("extract")
}

// Update re-extracts and updates the graph (no full rebuild).
func Update(args []string) int {
	if !tools.Exists("graphify") {
		ui.Error("graphify is not installed")
		return 1
	}
	ui.Step("updating graph")
	return runGraphify("update")
}

// runGraphify runs `graphify <sub> .` and reports the outcome.
func runGraphify(sub string) int {
	if err := tools.Run("graphify", sub, "."); err != nil {
		ui.Error("graphify %s failed: %v", sub, err)
		return 1
	}
	ui.OK("graphify %s done", sub)
	return 0
}
