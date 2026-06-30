package commands

import (
	"encoding/json"
	"os"

	"github.com/duclq/dev/internal/tools"
	"github.com/duclq/dev/internal/ui"
)

const memoryBin = "codebase-memory-mcp"

// memoryAvailable reports whether the codebase-memory-mcp binary is installed.
func memoryAvailable() bool { return tools.Exists(memoryBin) }

// Memory indexes the current repository into codebase-memory-mcp. It degrades
// gracefully when the optional binary is not installed.
func Memory(args []string) int {
	return indexMemory()
}

// indexMemory runs `codebase-memory-mcp cli index_repository` for the working
// directory. Re-running is incremental — the tool detects changes itself, so
// this doubles as the "reindex" step used by dev sync.
func indexMemory() int {
	if !memoryAvailable() {
		ui.Warn("%s not found — skipping memory step", memoryBin)
		ui.Info("%s", ui.Gray("run `dev setup memory` to install it"))
		return 0
	}

	wd, err := os.Getwd()
	if err != nil {
		ui.Error("cannot determine working directory: %v", err)
		return 1
	}

	payload, _ := json.Marshal(map[string]string{"repo_path": wd})
	ui.Step("%s cli index_repository", memoryBin)
	if err := tools.Run(memoryBin, "cli", "index_repository", string(payload)); err != nil {
		ui.Error("memory index failed: %v", err)
		return 1
	}
	ui.OK("memory indexed")
	return 0
}
