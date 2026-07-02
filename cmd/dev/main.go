// Command dev is one CLI for an AI-assisted development workflow. Cross-cutting
// verbs (init/setup/sync/doctor/ai) orchestrate a registry of tool modules;
// each tool also gets its own namespace, e.g. `dev graphify update`.
//
// To add a tool: implement core.Module in internal/modules/<tool> and add it to
// the `mods` slice below. Everything else (doctor/setup/sync/help) picks it up.
package main

import (
	"fmt"
	"os"

	"github.com/duclqDev99/custom-cli/internal/core"
	"github.com/duclqDev99/custom-cli/internal/modules/graphify"
	"github.com/duclqDev99/custom-cli/internal/modules/memory"
	"github.com/duclqDev99/custom-cli/internal/modules/uipro"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "0.1.0"

// alias maps a top-level shortcut to a module verb (e.g. `dev graph`).
type alias struct{ module, verb string }

// aliasOrder keeps shortcut help output stable.
var aliasOrder = []string{"graph", "update", "stats", "clean", "ui"}
var aliases = map[string]alias{
	"graph":  {"graphify", "graph"},
	"update": {"graphify", "update"},
	"stats":  {"graphify", "stats"},
	"clean":  {"graphify", "clean"},
	"ui":     {"uipro", "init"},
}

func main() {
	mods := []core.Module{graphify.New(), memory.New(), uipro.New()}

	args := os.Args[1:]
	if len(args) == 0 {
		usage(mods)
		return
	}

	switch args[0] {
	case "-h", "--help", "help":
		usage(mods)
		return
	case "-v", "--version", "version":
		fmt.Printf("dev %s\n", version)
		return
	case "init":
		os.Exit(core.Init(mods))
	case "setup":
		os.Exit(core.Setup(mods, args[1:]))
	case "sync":
		os.Exit(core.Sync(mods))
	case "doctor":
		os.Exit(core.Doctor(mods))
	case "ai":
		os.Exit(core.AI(args[1:]))
	}

	// Shortcut alias → module verb.
	if a, ok := aliases[args[0]]; ok {
		os.Exit(dispatchModule(mods, a.module, append([]string{a.verb}, args[1:]...)))
	}

	// Module namespace → `dev <module> <verb>`.
	for _, m := range mods {
		if m.Name() == args[0] {
			os.Exit(dispatchModule(mods, m.Name(), args[1:]))
		}
	}

	ui.Error("unknown command %q", args[0])
	fmt.Printf("run %s to see available commands\n", ui.Bold("dev help"))
	os.Exit(1)
}

// dispatchModule resolves `dev <module> [verb] [args]`, falling back to the
// module's default verb (or its help) when no verb is given.
func dispatchModule(mods []core.Module, name string, args []string) int {
	var m core.Module
	for _, x := range mods {
		if x.Name() == name {
			m = x
			break
		}
	}
	if m == nil {
		ui.Error("unknown module %q", name)
		return 1
	}

	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		moduleUsage(m)
		return 0
	}

	verb := m.Default()
	if len(args) > 0 && args[0][0] != '-' {
		verb = args[0]
		args = args[1:]
	}
	if verb == "" {
		moduleUsage(m)
		return 0
	}

	for _, c := range m.Commands() {
		if c.Name == verb {
			return c.Run(args)
		}
	}
	ui.Error("unknown %s command %q", name, verb)
	moduleUsage(m)
	return 1
}

func usage(mods []core.Module) {
	fmt.Printf("%s — one CLI for your AI dev workflow\n\n", ui.Bold("dev"))
	fmt.Printf("%s dev <command> [args]\n\n", ui.Bold("Usage:"))

	fmt.Println(ui.Bold("Workflow:"))
	for _, c := range [][2]string{
		{"init", "check & set up your dev environment"},
		{"setup", "configure all tools (or: dev setup <tool>)"},
		{"sync", "refresh every tool for this project"},
		{"doctor", "health of every tool & dependency"},
		{"ai", "check AI CLIs, API keys & proxy"},
	} {
		fmt.Printf("  %-9s %s\n", ui.Cyan(c[0]), ui.Gray(c[1]))
	}

	fmt.Printf("\n%s\n", ui.Bold("Tools:"))
	for _, m := range mods {
		fmt.Printf("  %-9s %s\n", ui.Cyan(m.Name()), ui.Gray(m.Summary()))
	}

	fmt.Printf("\n%s\n", ui.Bold("Shortcuts:"))
	for _, k := range aliasOrder {
		a := aliases[k]
		fmt.Printf("  %-9s %s\n", ui.Cyan(k), ui.Gray("→ dev "+a.module+" "+a.verb))
	}

	fmt.Printf("\n%s\n", ui.Gray("  dev <tool> --help · dev help · dev version"))
}

func moduleUsage(m core.Module) {
	fmt.Printf("%s — %s\n\n", ui.Bold("dev "+m.Name()), m.Summary())
	fmt.Println(ui.Bold("Commands:"))
	for _, c := range m.Commands() {
		suffix := ""
		if c.Name == m.Default() {
			suffix = ui.Gray(" (default)")
		}
		fmt.Printf("  %-9s %s%s\n", ui.Cyan(c.Name), ui.Gray(c.Desc), suffix)
	}
}
