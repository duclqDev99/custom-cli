package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/duclq/dev/internal/ui"
)

// Stats prints high-level statistics about the knowledge graph.
func Stats(args []string) int {
	path := graphJSON()
	if !fileExists(path) {
		ui.Warn("no graph found — run %s first", ui.Bold("dev graph"))
		return 1
	}

	data, err := os.ReadFile(path)
	if err != nil {
		ui.Error("read %s: %v", path, err)
		return 1
	}

	var g struct {
		Nodes []json.RawMessage `json:"nodes"`
		Edges []json.RawMessage `json:"edges"`
		Links []json.RawMessage `json:"links"`
	}
	if err := json.Unmarshal(data, &g); err != nil {
		ui.Error("parse %s: %v", path, err)
		return 1
	}

	edges := len(g.Edges)
	if edges == 0 {
		edges = len(g.Links)
	}

	ui.Header("Knowledge graph")
	ui.Info("file: %s (%s)", path, humanSize(len(data)))
	ui.OK("%d nodes", len(g.Nodes))
	ui.OK("%d edges", edges)
	if c := countCommunities(g.Nodes); c > 0 {
		ui.OK("%d communities", c)
	}
	return 0
}

// countCommunities counts distinct community/cluster ids across nodes.
func countCommunities(nodes []json.RawMessage) int {
	seen := map[string]bool{}
	for _, n := range nodes {
		var m map[string]any
		if json.Unmarshal(n, &m) != nil {
			continue
		}
		for _, key := range []string{"community", "cluster", "communityId", "community_id"} {
			if v, ok := m[key]; ok && v != nil {
				seen[fmt.Sprint(v)] = true
				break
			}
		}
	}
	return len(seen)
}

// humanSize formats a byte count as a human-readable string.
func humanSize(n int) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := int64(n) / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
