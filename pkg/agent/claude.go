package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/task"
)

// ClaudeCodeAgent implements the Agent interface for Claude Code CLI
type ClaudeCodeAgent struct {
	config *config.AgentConfig
}

// NewClaudeCodeAgent creates a new Claude Code agent instance
func NewClaudeCodeAgent(cfg *config.AgentConfig) (*ClaudeCodeAgent, error) {
	if cfg == nil {
		cfg = config.DefaultConfig("claude-code")
	}
	return &ClaudeCodeAgent{config: cfg}, nil
}

// Name returns the agent name
func (a *ClaudeCodeAgent) Name() string {
	return "claude-code"
}

// Run executes the task using Claude Code CLI
func (a *ClaudeCodeAgent) Run(ctx context.Context, t *task.Task, cfg *config.AgentConfig) (*Result, error) {
	if cfg == nil {
		cfg = a.config
	}

	result := &Result{
		Agent:     a.Name(),
		Task:      t.Name,
		StartTime: time.Now(),
	}

	// Convert to absolute path first
	absRepoPath, err := filepath.Abs(t.Repo)
	if err != nil {
		// Fallback to relative path if absolute resolution fails
		absRepoPath = t.Repo
	}

	// Prepend explicit working directory context to instructions
	// This is REQUIRED for CMB's execution environment. While standalone tests show
	// cmd.Dir should work, in CMB's context (even with correct arg order) Claude
	// creates files in the project root without explicit path guidance.
	enhancedInstructions := fmt.Sprintf(`IMPORTANT: You are working in the directory: %s

All file paths you create should be relative to this directory or use absolute paths within this directory.
When creating files, ensure they are created in: %s

---

%s`, absRepoPath, absRepoPath, t.Instructions)

	// Build command arguments for non-interactive (headless) mode
	// IMPORTANT: The prompt must come IMMEDIATELY after -p flag, not at the end!
	args := []string{
		"-p", enhancedInstructions, // Prompt must be right after -p
		"--dangerously-skip-permissions", // Auto-approve tool use (safe in sandbox)
		"--model", cfg.Model.Name,
	}

	// Add system prompt if provided
	if cfg.Prompts.System != "" {
		args = append(args, "--system-prompt", cfg.Prompts.System)
	}

	fmt.Println("   Invoking Claude Code CLI...")
	fmt.Printf("   Command: claude --model %s\n", cfg.Model.Name)
	fmt.Println()

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, "claude", args...)

	// Set working directory (belt-and-suspenders with explicit path above)
	cmd.Dir = absRepoPath
	cmd.Env = os.Environ()

	// Capture output
	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	// Start the command
	if err = cmd.Start(); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to start command: %v", err)
		return result, err
	}

	// Show progress while waiting
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	brandText := "codematic bench...."
	windowSize := 9
	position := 0
	// Add padding to make it scroll smoothly
	paddedText := "   " + brandText + "   "
	lastOutputLen := 0

	fmt.Print(" ")

	for {
		select {
		case err = <-done:
			// Clear the line and show completion
			fmt.Print("\r   [DONE] Agent completed                                                    \n")

			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			result.Output = outputBuf.String()

			// Show final output
			if output := outputBuf.String(); output != "" {
				fmt.Println()
				fmt.Println("--- Agent Output ---")
				fmt.Print(output)
				fmt.Println("--- End Agent Output ---")
			}
			fmt.Println()
			goto finish

		case <-ticker.C:
			elapsed := time.Since(result.StartTime)

			// Check if there's new output
			currentOutput := outputBuf.String()
			if len(currentOutput) > lastOutputLen {
				// Clear current line and show progress
				fmt.Print("\r   [OUTPUT] New output received...                                          \n   ")
				lastOutputLen = len(currentOutput)
			} else {
				// Create scrolling window
				start := position % len(paddedText)
				end := start + windowSize
				var display string
				if end <= len(paddedText) {
					display = paddedText[start:end]
				} else {
					// Wrap around
					display = paddedText[start:] + paddedText[:end-len(paddedText)]
				}

				fmt.Printf("\r   [%s] Working... (elapsed: %ds)   ",
					display, int(elapsed.Seconds()))

				position++
			}
		}
	}

finish:
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("claude-code execution failed: %v\nOutput: %s", err, result.Output)
		return result, err
	}

	result.Success = true
	return result, nil
}
