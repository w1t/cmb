package config

// AgentConfig represents the complete configuration for an agent
type AgentConfig struct {
	Name     string        `yaml:"name"`
	Agent    string        `yaml:"agent"`
	Model    ModelConfig   `yaml:"model"`
	Prompts  PromptsConfig `yaml:"prompts"`
	Settings Settings      `yaml:"settings,omitempty"`
	Context  ContextConfig `yaml:"context,omitempty"`
}

// ModelConfig defines the LLM model configuration
type ModelConfig struct {
	Provider    string  `yaml:"provider"`
	Name        string  `yaml:"name"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
}

// PromptsConfig contains custom prompts for the agent
type PromptsConfig struct {
	System       string `yaml:"system,omitempty"`
	SpecTemplate string `yaml:"spec_template,omitempty"`
	Review       string `yaml:"review,omitempty"`
}

// Settings contains agent-specific settings
type Settings struct {
	Mode         string            `yaml:"mode,omitempty"`
	AutoCommit   bool              `yaml:"auto_commit,omitempty"`
	SpecFirst    bool              `yaml:"spec_first,omitempty"`
	MaxPlanDepth int               `yaml:"max_plan_depth,omitempty"`
	Extra        map[string]string `yaml:",inline"` // Catch-all for agent-specific settings
}

// ContextConfig defines context window settings
type ContextConfig struct {
	MaxFiles          int      `yaml:"max_files,omitempty"`
	IncludeTestFiles  bool     `yaml:"include_test_files,omitempty"`
	PersistContext    bool     `yaml:"persist_context,omitempty"`
	DocumentationURLs []string `yaml:"documentation_urls,omitempty"`
}

// Validate checks if the configuration is valid
func (c *AgentConfig) Validate() error {
	if c.Agent == "" {
		return &ValidationError{Field: "agent", Message: "agent name is required"}
	}
	if c.Model.Provider == "" {
		return &ValidationError{Field: "model.provider", Message: "model provider is required"}
	}
	if c.Model.Name == "" {
		return &ValidationError{Field: "model.name", Message: "model name is required"}
	}
	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Field + ": " + e.Message
}
