package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// runTests executes the test command and returns pass/fail status and output
func runTests(workdir, testCmd string) (bool, string, error) {
	// Parse the test command (simple shell command parsing)
	parts := strings.Fields(testCmd)
	if len(parts) == 0 {
		return false, "", fmt.Errorf("empty test command")
	}

	// Create command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = workdir

	// Run and capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Test passed if command succeeded (exit code 0)
	passed := err == nil

	return passed, outputStr, nil
}
