package commands

import (
	"fmt"

	"github.com/duclq/dev/internal/tools"
	"github.com/duclq/dev/internal/ui"
)

// Sync runs the full refresh pipeline: graph → memory → git status.
func Sync(args []string) int {
	// 1. graph
	ui.Step("1/3 graphify")
	if tools.Exists("graphify") {
		sub := "update"
		if !fileExists(graphJSON()) {
			ui.Warn("no graph yet — extracting first")
			sub = "extract"
		}
		if err := tools.Run("graphify", sub, "."); err != nil {
			ui.Error("graphify %s failed: %v", sub, err)
			return 1
		}
		ui.OK("graph synced")
	} else {
		ui.Warn("graphify not installed — skipping graph step")
	}

	// 2. memory (graceful; index_repository is incremental, so this reindexes)
	ui.Step("2/3 memory reindex")
	indexMemory()

	// 3. git status
	ui.Step("3/3 git status")
	if inGitRepo() {
		_ = tools.Run("git", "status", "--short", "--branch")
	} else {
		ui.Warn("not a git repository")
	}

	fmt.Println()
	ui.OK("sync complete")
	return 0
}
