package config

import "testing"

func TestAgentConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  AgentConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AgentConfig{
				Agent: "opencode",
				Model: ModelConfig{
					Provider: "anthropic",
					Name:     "claude-sonnet-4",
				},
			},
			wantErr: false,
		},
		{
			name: "missing agent",
			config: AgentConfig{
				Model: ModelConfig{
					Provider: "anthropic",
					Name:     "claude-sonnet-4",
				},
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			config: AgentConfig{
				Agent: "opencode",
				Model: ModelConfig{
					Name: "claude-sonnet-4",
				},
			},
			wantErr: true,
		},
		{
			name: "missing model name",
			config: AgentConfig{
				Agent: "opencode",
				Model: ModelConfig{
					Provider: "anthropic",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		agent    string
		wantName string
	}{
		{"opencode", "opencode-default"},
		{"claude-code", "claude-code-default"},
		{"kiro", "kiro-default"},
		{"unknown", "unknown-default"},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			cfg := DefaultConfig(tt.agent)
			if cfg.Name != tt.wantName {
				t.Errorf("DefaultConfig().Name = %q, want %q", cfg.Name, tt.wantName)
			}
			if cfg.Agent != tt.agent {
				t.Errorf("DefaultConfig().Agent = %q, want %q", cfg.Agent, tt.agent)
			}
			if cfg.Model.Provider == "" {
				t.Error("DefaultConfig().Model.Provider is empty")
			}
			if cfg.Model.Name == "" {
				t.Error("DefaultConfig().Model.Name is empty")
			}
		})
	}
}
