// Package ui handles colored terminal output for the dev CLI.
package ui

import (
	"fmt"
	"os"
)

// useColor is disabled when NO_COLOR is set or the terminal is dumb.
var useColor = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

func paint(code, s string) string {
	if !useColor {
		return s
	}
	return code + s + reset
}

// Bold returns s in bold.
func Bold(s string) string { return paint(bold, s) }

// Dim returns s dimmed.
func Dim(s string) string { return paint(dim, s) }

// Red returns s in red.
func Red(s string) string { return paint(red, s) }

// Green returns s in green.
func Green(s string) string { return paint(green, s) }

// Yellow returns s in yellow.
func Yellow(s string) string { return paint(yellow, s) }

// Blue returns s in blue.
func Blue(s string) string { return paint(blue, s) }

// Cyan returns s in cyan.
func Cyan(s string) string { return paint(cyan, s) }

// Gray returns s in gray.
func Gray(s string) string { return paint(gray, s) }

// Header prints a section header with spacing.
func Header(title string) {
	fmt.Println()
	fmt.Println(Bold(Cyan("▌ " + title)))
}

// OK prints a green check line.
func OK(format string, a ...any) {
	fmt.Printf("  %s %s\n", Green("✓"), fmt.Sprintf(format, a...))
}

// Fail prints a red cross line.
func Fail(format string, a ...any) {
	fmt.Printf("  %s %s\n", Red("✗"), fmt.Sprintf(format, a...))
}

// Warn prints a yellow warning line.
func Warn(format string, a ...any) {
	fmt.Printf("  %s %s\n", Yellow("⚠"), fmt.Sprintf(format, a...))
}

// Info prints a neutral bullet line.
func Info(format string, a ...any) {
	fmt.Printf("  %s %s\n", Blue("•"), fmt.Sprintf(format, a...))
}

// Step prints a top-level progress arrow.
func Step(format string, a ...any) {
	fmt.Printf("%s %s\n", Cyan("→"), fmt.Sprintf(format, a...))
}

// Error prints a red error line to stderr.
func Error(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Red("error:"), fmt.Sprintf(format, a...))
}
