package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/duclq/dev/internal/ui"
)

// Clean removes the generated graphify-out/ directory. It asks for
// confirmation unless -f/--force is passed.
func Clean(args []string) int {
	force := false
	for _, a := range args {
		if a == "-f" || a == "--force" {
			force = true
		}
	}

	if !fileExists(graphDir) {
		ui.Info("nothing to clean (%s/ not found)", graphDir)
		return 0
	}

	if !force {
		fmt.Printf("Remove %s/ ? [y/N] ", graphDir)
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans != "y" && ans != "yes" {
			ui.Info("aborted")
			return 0
		}
	}

	if err := os.RemoveAll(graphDir); err != nil {
		ui.Error("failed to remove %s/: %v", graphDir, err)
		return 1
	}
	ui.OK("removed %s/", graphDir)
	return 0
}
