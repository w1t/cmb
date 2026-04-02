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

// CodexAgent implements the Agent interface for OpenAI's Codex CLI
type CodexAgent struct {
	config *config.AgentConfig
}

// NewCodexAgent creates a new Codex agent instance
func NewCodexAgent(cfg *config.AgentConfig) (*CodexAgent, error) {
	if cfg == nil {
		cfg = config.DefaultConfig("codex")
	}
	return &CodexAgent{config: cfg}, nil
}

// Name returns the agent name
func (a *CodexAgent) Name() string {
	return "codex"
}

// Run executes the task using Codex CLI
func (a *CodexAgent) Run(ctx context.Context, t *task.Task, cfg *config.AgentConfig) (*Result, error) {
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

	// Build command arguments for non-interactive (headless) mode
	// Codex uses: codex exec [options] [prompt]
	args := []string{
		"exec", // Non-interactive execution subcommand
		"--dangerously-bypass-approvals-and-sandbox", // Auto-approve (safe in our git worktree sandbox)
		"-C", absRepoPath, // Set working directory
	}

	// Add model if specified
	if cfg.Model.Name != "" {
		args = append(args, "--model", cfg.Model.Name)
	}

	// Add the task instructions as the prompt (must be last positional argument)
	args = append(args, t.Instructions)

	fmt.Println("   Invoking Codex CLI...")
	fmt.Printf("   Command: codex exec -C %s --model %s\n", absRepoPath, cfg.Model.Name)
	fmt.Println()
	fmt.Println("--- Agent Output (Live) ---")

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, "codex", args...)
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
		result.Error = fmt.Sprintf("codex execution failed: %v", err)
		return result, err
	}

	result.Success = true
	return result, nil
}
