package core

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/duclqDev99/custom-cli/internal/tools"
	"github.com/duclqDev99/custom-cli/internal/ui"
)

// GraphDir is where graphify writes its output.
const GraphDir = "graphify-out"

// GraphJSON is the path to the built knowledge graph.
func GraphJSON() string { return filepath.Join(GraphDir, "graph.json") }

// FileExists reports whether a path exists.
func FileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// InGitRepo reports whether the working directory is inside a git work tree.
func InGitRepo() bool {
	out, err := tools.Capture("git", "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// EnvChecks reports on the language/runtime tools dev relies on but does not
// own — these are not modules, just expectations of the environment.
func EnvChecks() []Check {
	return ChecksFor([]tools.Tool{
		{Name: "Go", Bin: "go", VersionArg: []string{"version"}, Optional: true, Hint: goInstallHint()},
		{Name: "Git", Bin: "git", VersionArg: []string{"--version"}},
		{Name: "Node.js", Bin: "node", VersionArg: []string{"--version"}},
		{Name: "Python", Bin: "python3", VersionArg: []string{"--version"}},
		{Name: "Docker", Bin: "docker", VersionArg: []string{"--version"}, Optional: true, Hint: "install Docker Desktop"},
		{Name: "Claude Code", Bin: "claude", VersionArg: []string{"--version"}, Optional: true, Hint: "npm i -g @anthropic-ai/claude-code"},
		{Name: "Redis", Bin: "redis-cli", VersionArg: []string{"--version"}, Optional: true, Hint: "brew install redis"},
		{Name: "PostgreSQL", Bin: "psql", VersionArg: []string{"--version"}, Optional: true, Hint: "brew install postgresql"},
	})
}

// goInstallHint suggests a Go install command for the current platform; the
// full per-OS list lives in install.sh (`./install.sh --check`).
func goInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install go"
	case "linux":
		return "apt-get install golang-go / dnf install golang / pacman -S go"
	case "windows":
		return "winget install GoLang.Go"
	default:
		return "https://go.dev/dl/"
	}
}

// ChecksFor converts tool definitions into health checks. Modules use this to
// avoid reimplementing presence/version detection.
func ChecksFor(ts []tools.Tool) []Check {
	out := make([]Check, 0, len(ts))
	for _, t := range ts {
		path, ok := t.Found()
		c := Check{Name: t.Name, OK: ok, Optional: t.Optional}
		if ok {
			if c.Detail = t.Version(); c.Detail == "" {
				c.Detail = path
			}
		} else {
			c.Detail = t.Hint
		}
		out = append(out, c)
	}
	return out
}

// RenderChecks prints checks and returns the number of failing required ones.
func RenderChecks(checks []Check) (missing int) {
	for _, c := range checks {
		if c.OK {
			ui.OK("%s %s", c.Name, ui.Gray(c.Detail))
			continue
		}
		tail := ""
		if c.Detail != "" {
			tail = " " + ui.Gray("— "+c.Detail)
		}
		if c.Optional {
			ui.Warn("%s%s", c.Name, tail)
		} else {
			ui.Fail("%s%s", c.Name, tail)
			missing++
		}
	}
	return missing
}

// reportProject prints the state of the current project directory.
func reportProject() {
	wd, _ := os.Getwd()
	ui.Info("dir: %s", ui.Gray(wd))

	if InGitRepo() {
		branch, _ := tools.Capture("git", "rev-parse", "--abbrev-ref", "HEAD")
		ui.OK("git repository (%s)", branch)
	} else {
		ui.Warn("not a git repository")
	}

	if FileExists(GraphJSON()) {
		ui.OK("knowledge graph present (%s)", GraphJSON())
	} else {
		ui.Warn("no graph yet — run %s", ui.Bold("dev graph"))
	}
}
