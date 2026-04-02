package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/task"
	"github.com/creack/pty"
)

// OpenCodeAgent implements the Agent interface for OpenCode CLI
type OpenCodeAgent struct {
	config *config.AgentConfig
}

// NewOpenCodeAgent creates a new OpenCode agent instance
func NewOpenCodeAgent(cfg *config.AgentConfig) (*OpenCodeAgent, error) {
	if cfg == nil {
		cfg = config.DefaultConfig("opencode")
	}
	return &OpenCodeAgent{config: cfg}, nil
}

// Name returns the agent name
func (a *OpenCodeAgent) Name() string {
	return "opencode"
}

// Run executes the task using OpenCode CLI
func (a *OpenCodeAgent) Run(ctx context.Context, t *task.Task, cfg *config.AgentConfig) (*Result, error) {
	if cfg == nil {
		cfg = a.config
	}

	result := &Result{
		Agent:     a.Name(),
		Task:      t.Name,
		StartTime: time.Now(),
	}

	// Build command arguments for non-interactive (headless) mode
	// OpenCode uses: opencode run [message] --model <model>
	args := []string{
		"run",          // Non-interactive run command
		t.Instructions, // Task instructions as message
		"--model", fmt.Sprintf("%s/%s", cfg.Model.Provider, cfg.Model.Name),
	}

	// Add system prompt if provided (note: might need to be in message for OpenCode)
	// OpenCode doesn't have --system flag for run command

	fmt.Println("   Invoking OpenCode CLI...")
	fmt.Printf("   Command: opencode run --model %s/%s\n", cfg.Model.Provider, cfg.Model.Name)
	fmt.Println()
	fmt.Println("--- Agent Output (Live) ---")

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, "opencode", args...)
	cmd.Dir = t.Repo

	// Start command with a pseudo-TTY to force unbuffered output
	var outputBuf bytes.Buffer
	ptmx, err := pty.Start(cmd)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to start command with pty: %v", err)
		return result, err
	}
	defer ptmx.Close()

	// Copy PTY output to both stdout and our buffer in real-time
	done := make(chan error)
	go func() {
		_, copyErr := io.Copy(io.MultiWriter(os.Stdout, &outputBuf), ptmx)
		done <- copyErr
	}()

	// Wait for command to finish
	err = cmd.Wait()
	<-done // Wait for copy to finish

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = outputBuf.String()

	fmt.Println()
	fmt.Println("--- End Agent Output ---")
	fmt.Println()

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("opencode execution failed: %v", err)
		return result, err
	}

	result.Success = true
	return result, nil
}
