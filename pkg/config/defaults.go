package config

import (
	"fmt"
	"path/filepath"
)

// defaultReviewTemplate is the standard review prompt template
const defaultReviewTemplate = `You are reviewing code changes made by another agent ({AGENT}).

Original Task: {TASK}

Code Changes:
{DIFF}

Please provide a detailed code review with the following structure:

SCORES (1-5, where 5 is excellent):
- Code Quality: [score]/5
- Correctness: [score]/5
- Maintainability: [score]/5

DETAILED REVIEW:
1. Code Quality & Best Practices
   - What's good
   - What needs improvement

2. Correctness & Bugs
   - Potential issues
   - Edge cases

3. Security & Performance
   - Security concerns
   - Performance implications

4. Suggestions for Improvement
   - Specific recommendations
   - Code examples if applicable

OVERALL ASSESSMENT: [APPROVE / REQUEST CHANGES / REJECT]

Be constructive and specific in your feedback.`

// DefaultConfig returns the default configuration for a given agent
// Note: Model names vary by agent:
// - Claude Code uses aliases: "sonnet", "opus", "haiku"
// - OpenCode uses full names: "claude-sonnet-4-5-20250929"
// - Other agents may have their own naming conventions
//
// For agents with YAML config files (codex, aider), this will attempt to load
// from config/{agent}-default.yaml first, falling back to hardcoded defaults.
func DefaultConfig(agentName string) *AgentConfig {
	// Try to load from YAML file first for agents that have config files
	configPath := filepath.Join("config", fmt.Sprintf("%s-default.yaml", agentName))
	if cfg, err := Load(configPath); err == nil {
		return cfg
	}

	// Fall back to hardcoded defaults
	switch agentName {
	case "opencode":
		return &AgentConfig{
			Name:  "opencode-default",
			Agent: "opencode",
			Model: ModelConfig{
				Provider:    "anthropic",
				Name:        "claude-sonnet-4",
				Temperature: 0.0,
				MaxTokens:   4096,
			},
			Prompts: PromptsConfig{
				System: "You are an expert software engineer. Focus on maintainability and type safety.",
				Review: defaultReviewTemplate,
			},
			Context: ContextConfig{
				MaxFiles: 20,
			},
		}
	case "claude-code":
		return &AgentConfig{
			Name:  "claude-code-default",
			Agent: "claude-code",
			Model: ModelConfig{
				Provider:    "anthropic",
				Name:        "sonnet", // Claude Code alias for latest Sonnet
				Temperature: 0.0,
				MaxTokens:   4096,
			},
			Prompts: PromptsConfig{
				System: "You are Claude, an AI assistant focused on helping with code. Be thorough, precise, and explain your reasoning.",
				Review: defaultReviewTemplate,
			},
			Context: ContextConfig{
				MaxFiles:         15,
				IncludeTestFiles: true,
			},
		}
	case "kiro":
		return &AgentConfig{
			Name:  "kiro-default",
			Agent: "kiro",
			Model: ModelConfig{
				Provider:    "aws-bedrock",
				Name:        "claude-sonnet-4",
				Temperature: 0.0,
				MaxTokens:   8192,
			},
			Prompts: PromptsConfig{
				System:       "You are a senior software engineer using spec-driven development. Always start with a clear specification before implementation.",
				SpecTemplate: "## Specification\n- Goal: {goal}\n- Success Criteria: {criteria}\n- Constraints: {constraints}",
				Review:       defaultReviewTemplate,
			},
			Settings: Settings{
				SpecFirst:    true,
				MaxPlanDepth: 3,
			},
			Context: ContextConfig{
				MaxFiles:       50,
				PersistContext: true,
			},
		}
	default:
		// Generic default
		return &AgentConfig{
			Name:  agentName + "-default",
			Agent: agentName,
			Model: ModelConfig{
				Provider:    "anthropic",
				Name:        "claude-sonnet-4",
				Temperature: 0.0,
				MaxTokens:   4096,
			},
			Context: ContextConfig{
				MaxFiles: 20,
			},
		}
	}
}
