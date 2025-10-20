package ui

import (
	"fmt"
	"os"
	"os/exec"
)

// hasGum checks if gum is installed
func hasGum() bool {
	_, err := exec.LookPath("gum")
	return err == nil
}

// Spin runs a function with styled output
// Falls back to simple output if gum not installed
func Spin(title string, fn func() error) error {
	Info(title + "...")
	return fn()
}

// SpinCommand runs a command with a gum spinner
func SpinCommand(title string, name string, args ...string) error {
	if !hasGum() {
		fmt.Printf("→ %s...\n", title)
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Build gum spin command: gum spin --title "..." -- command args
	gumArgs := []string{"spin", "--spinner", "dot", "--title", title, "--show-error", "--", name}
	gumArgs = append(gumArgs, args...)

	cmd := exec.Command("gum", gumArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// SpinFunc runs a shell command string with a spinner
func SpinFunc(title string, shellCmd string) error {
	if !hasGum() {
		fmt.Printf("→ %s...\n", title)
		cmd := exec.Command("sh", "-c", shellCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("gum", "spin", "--spinner", "dot", "--title", title, "--show-error", "--", "sh", "-c", shellCmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Success prints a success message
func Success(msg string) {
	if hasGum() {
		exec.Command("gum", "style", "--foreground", "212", "✓ "+msg).Run()
	} else {
		fmt.Printf("✓ %s\n", msg)
	}
}

// Error prints an error message
func Error(msg string) {
	if hasGum() {
		exec.Command("gum", "style", "--foreground", "196", "✗ "+msg).Run()
	} else {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	}
}

// Info prints an info message
func Info(msg string) {
	if hasGum() {
		exec.Command("gum", "style", "--foreground", "86", "→ "+msg).Run()
	} else {
		fmt.Printf("→ %s\n", msg)
	}
}
