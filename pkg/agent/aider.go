package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/task"
)

// AiderAgent implements the Agent interface for Aider CLI
type AiderAgent struct {
	config *config.AgentConfig
}

// NewAiderAgent creates a new Aider agent instance
func NewAiderAgent(cfg *config.AgentConfig) (*AiderAgent, error) {
	if cfg == nil {
		cfg = config.DefaultConfig("aider")
	}
	return &AiderAgent{config: cfg}, nil
}

// Name returns the agent name
func (a *AiderAgent) Name() string {
	return "aider"
}

// Run executes the task using Aider CLI
func (a *AiderAgent) Run(ctx context.Context, t *task.Task, cfg *config.AgentConfig) (*Result, error) {
	if cfg == nil {
		cfg = a.config
	}

	result := &Result{
		Agent:     a.Name(),
		Task:      t.Name,
		StartTime: time.Now(),
	}

	// Convert to absolute path
	absRepoPath, err := filepath.Abs(t.Repo)
	if err != nil {
		absRepoPath = t.Repo
	}

	// Build command arguments for non-interactive (CI-safe) mode
	// Aider uses: aider --yes-always --auto-commits --no-stream --message "task" [files...]
	args := []string{
		"--yes-always",              // No confirmations (CI-safe)
		"--auto-commits",            // Auto-commit changes
		"--no-stream",               // No TTY streaming (CI-safe)
		"--message", t.Instructions, // One-shot task
	}

	// Add model if specified
	if cfg.Model.Name != "" {
		args = append(args, "--model", cfg.Model.Name)
	}

	// Note: Aider accepts file paths as positional arguments at the end.
	// For now, we let Aider work with the whole repo (it has smart file selection).
	// Future enhancement: allow task.Files to specify explicit file list.

	fmt.Println("   Invoking Aider CLI...")
	fmt.Printf("   Command: aider --yes-always --auto-commits --no-stream --model %s\n", cfg.Model.Name)
	fmt.Println()
	fmt.Println("--- Agent Output (Live) ---")

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, "aider", args...)
	cmd.Dir = absRepoPath
	cmd.Env = os.Environ()

	// Capture output in real-time
	var outputBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &outputBuf)

	// Run the command
	err = cmd.Run()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = outputBuf.String()

	fmt.Println()
	fmt.Println("--- End Agent Output ---")
	fmt.Println()

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("aider execution failed: %v", err)
		return result, err
	}

	result.Success = true
	return result, nil
}
