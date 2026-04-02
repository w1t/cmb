package agent

import (
	"context"
	"fmt"

	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/task"
)

// Agent represents a coding agent that can execute tasks
type Agent interface {
	Name() string
	Run(ctx context.Context, t *task.Task, cfg *config.AgentConfig) (*Result, error)
}

// New creates a new agent instance based on the agent name
func New(name string, cfg *config.AgentConfig) (Agent, error) {
	switch name {
	case "opencode":
		return NewOpenCodeAgent(cfg)
	case "claude-code":
		return NewClaudeCodeAgent(cfg)
	case "codex":
		return NewCodexAgent(cfg)
	case "aider":
		return NewAiderAgent(cfg)
	case "kiro":
		return NewKiroAgent(cfg)
	default:
		return nil, fmt.Errorf("unknown agent: %s (supported: opencode, claude-code, codex, aider, kiro)", name)
	}
}

// ValidateAgentAvailable checks if the agent CLI is available in PATH
func ValidateAgentAvailable(name string) error {
	var cmd string
	switch name {
	case "opencode":
		cmd = "opencode"
	case "claude-code":
		cmd = "claude"
	case "codex":
		cmd = "codex"
	case "aider":
		cmd = "aider"
	case "kiro":
		cmd = "kiro"
	default:
		return fmt.Errorf("unknown agent: %s", name)
	}

	// TODO: Actually check if command exists in PATH
	// For now, just return nil - we'll handle missing binaries when we try to run them
	_ = cmd
	return nil
}
