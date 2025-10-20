package run

import (
	"fmt"
	"os"
	"os/exec"
)

// OnHost executes a command on a specific host
func OnHost(host, command string) error {
	cmd := exec.Command("ssh", host, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// OnAll executes a command on all hosts in parallel
func OnAll(hosts []string, command string) error {
	errors := make(chan error, len(hosts))

	for _, host := range hosts {
		go func(h string) {
			fmt.Printf("\n→ Running on %s\n", h)
			errors <- OnHost(h, command)
		}(host)
	}

	var lastErr error
	for range hosts {
		if err := <-errors; err != nil {
			lastErr = err
		}
	}

	return lastErr
}
