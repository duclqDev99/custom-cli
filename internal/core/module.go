// Package core defines the module contract every integrated tool implements,
// plus the cross-cutting orchestrators (doctor, setup, sync, init) that loop
// over the registered modules.
//
// To add a new tool: create internal/modules/<tool> implementing Module and
// register it in cmd/dev/main.go. No other file needs to change.
package core

// Check is a single health-check result rendered by `dev doctor`.
type Check struct {
	Name     string // what was checked
	OK       bool   // did it pass?
	Optional bool   // if true, absence is a warning rather than a failure
	Detail   string // version/path when OK, remediation hint when not
}

// Command is one verb under a module: `dev <module> <name>`.
type Command struct {
	Name string
	Desc string
	Run  func(args []string) int
}

// Module is one tool dev orchestrates (graphify, memory, ...).
type Module interface {
	// Name is the namespace used as `dev <name> ...`.
	Name() string
	// Summary is a one-line description for help output.
	Summary() string
	// Commands are the verbs available under this module.
	Commands() []Command
	// Default is the verb run when `dev <name>` is given no verb ("" = help).
	Default() string
	// Doctor returns the health checks this module contributes to `dev doctor`.
	Doctor() []Check
	// Setup performs `dev setup <name>` and returns an exit code.
	Setup(args []string) int
	// Sync performs this module's step of `dev sync` (0 = ok or skipped).
	Sync() int
}
