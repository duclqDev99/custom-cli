// Package tools wraps external binaries the dev CLI orchestrates.
package tools

import (
	"os"
	"os/exec"
	"strings"
)

// Tool describes an external dependency.
type Tool struct {
	Name       string   // human-readable name
	Bin        string   // executable looked up on PATH
	VersionArg []string // args that print a version (empty = skip)
	Optional   bool     // true if absence is a warning, not a failure
	Hint       string   // install hint shown when missing
}

// Found reports the resolved path and whether the binary is on PATH.
func (t Tool) Found() (string, bool) {
	path, err := exec.LookPath(t.Bin)
	if err != nil {
		return "", false
	}
	return path, true
}

// Version runs the version command and returns its first non-empty line.
func (t Tool) Version() string {
	if len(t.VersionArg) == 0 {
		return ""
	}
	out, err := exec.Command(t.Bin, t.VersionArg...).CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return ""
}

// Exists reports whether a binary is available on PATH.
func Exists(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// Run executes a command, streaming stdio in the current working directory.
func Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Capture runs a command and returns its trimmed combined output.
func Capture(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
