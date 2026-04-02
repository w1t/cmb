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

// KiroAgent implements the Agent interface for Kiro CLI
type KiroAgent struct {
	config *config.AgentConfig
}

// NewKiroAgent creates a new Kiro agent instance
func NewKiroAgent(cfg *config.AgentConfig) (*KiroAgent, error) {
	if cfg == nil {
		cfg = config.DefaultConfig("kiro")
	}
	return &KiroAgent{config: cfg}, nil
}

// Name returns the agent name
func (a *KiroAgent) Name() string {
	return "kiro"
}

// Run executes the task using Kiro CLI
func (a *KiroAgent) Run(ctx context.Context, t *task.Task, cfg *config.AgentConfig) (*Result, error) {
	if cfg == nil {
		cfg = a.config
	}

	result := &Result{
		Agent:     a.Name(),
		Task:      t.Name,
		StartTime: time.Now(),
	}

	// Build command arguments for Kiro
	args := []string{
		"solve",
		"--model", cfg.Model.Name,
		"--task", t.Instructions,
	}

	fmt.Println("   Invoking Kiro CLI...")
	fmt.Printf("   Command: kiro solve --model %s\n", cfg.Model.Name)
	fmt.Println()
	fmt.Println("--- Agent Output (Live) ---")

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, "kiro", args...)
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
		result.Error = fmt.Sprintf("kiro execution failed: %v", err)
		return result, err
	}

	result.Success = true
	return result, nil
}
