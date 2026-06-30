// Command dev is one CLI for an AI-assisted development workflow:
// it wraps graphify, codebase memory, git and friends behind a handful
// of memorable verbs.
package main

import (
	"fmt"
	"os"

	"github.com/duclqDev99/custom-cli/internal/commands"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "0.1.0"

type command struct {
	name string
	desc string
	run  func(args []string) int
}

func main() {
	cmds := []command{
		{"init", "check & set up your dev environment", commands.Init},
		{"setup", "configure graphify & codebase-memory-mcp", commands.Setup},
		{"sync", "graphify update → memory reindex → git status", commands.Sync},
		{"graph", "build the graph, or update it if it exists", commands.Graph},
		{"update", "re-extract & update the graph", commands.Update},
		{"memory", "index the codebase memory (optional tool)", commands.Memory},
		{"doctor", "show the health of every dependency", commands.Doctor},
		{"ai", "check AI CLIs, API keys & proxy", commands.AI},
		{"stats", "show knowledge-graph statistics", commands.Stats},
		{"clean", "remove generated graphify-out/ artifacts", commands.Clean},
	}

	args := os.Args[1:]
	if len(args) == 0 {
		usage(cmds)
		return
	}

	switch args[0] {
	case "-h", "--help", "help":
		usage(cmds)
		return
	case "-v", "--version", "version":
		fmt.Printf("dev %s\n", version)
		return
	}

	for _, c := range cmds {
		if c.name == args[0] {
			os.Exit(c.run(args[1:]))
		}
	}

	ui.Error("unknown command %q", args[0])
	fmt.Printf("run %s to see available commands\n", ui.Bold("dev help"))
	os.Exit(1)
}

func usage(cmds []command) {
	fmt.Printf("%s — one CLI for your AI dev workflow\n\n", ui.Bold("dev"))
	fmt.Printf("%s dev <command> [flags]\n\n", ui.Bold("Usage:"))
	fmt.Println(ui.Bold("Commands:"))
	for _, c := range cmds {
		fmt.Printf("  %-8s %s\n", ui.Cyan(c.name), ui.Gray(c.desc))
	}
	fmt.Printf("\n%s\n", ui.Gray("  dev help · dev version"))
}
