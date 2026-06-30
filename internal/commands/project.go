package commands

import (
	"os"
	"path/filepath"

	"github.com/duclq/dev/internal/tools"
	"github.com/duclq/dev/internal/ui"
)

// graphDir is where graphify writes its output.
const graphDir = "graphify-out"

// graphJSON is the path to the built knowledge graph.
func graphJSON() string { return filepath.Join(graphDir, "graph.json") }

// fileExists reports whether a path exists.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// inGitRepo reports whether the working directory is inside a git work tree.
func inGitRepo() bool {
	out, err := tools.Capture("git", "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// reportProject prints the state of the current project directory.
func reportProject() {
	wd, _ := os.Getwd()
	ui.Info("dir: %s", ui.Gray(wd))

	if inGitRepo() {
		branch, _ := tools.Capture("git", "rev-parse", "--abbrev-ref", "HEAD")
		ui.OK("git repository (%s)", branch)
	} else {
		ui.Warn("not a git repository")
	}

	if fileExists(graphJSON()) {
		ui.OK("knowledge graph present (%s)", graphJSON())
	} else {
		ui.Warn("no graph yet — run %s", ui.Bold("dev graph"))
	}
}
