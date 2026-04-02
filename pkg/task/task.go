package task

import "time"

// Task represents a repository-scale coding task for agents to complete
type Task struct {
	Name            string        `yaml:"name"`
	Description     string        `yaml:"description"`
	Language        string        `yaml:"language"`
	Repo            string        `yaml:"repo"`
	Instructions    string        `yaml:"instructions"`
	SuccessCriteria []string      `yaml:"success_criteria"`
	Evaluation      EvalConfig    `yaml:"evaluation"`
	Timeout         time.Duration `yaml:"timeout,omitempty"`
	InitialState    string        `yaml:"initial_state,omitempty"` // Git commit/branch
}

// EvalConfig defines how to evaluate task success
type EvalConfig struct {
	RunTests  string `yaml:"run_tests,omitempty"`
	CheckDiff bool   `yaml:"check_diff,omitempty"`
	CustomCmd string `yaml:"custom_cmd,omitempty"`
}

// Validate checks if the task definition is valid
func (t *Task) Validate() error {
	if t.Name == "" {
		return &ValidationError{Field: "name", Message: "task name is required"}
	}
	if t.Repo == "" {
		return &ValidationError{Field: "repo", Message: "repository path is required"}
	}
	if t.Instructions == "" {
		return &ValidationError{Field: "instructions", Message: "instructions are required"}
	}
	if t.Timeout == 0 {
		t.Timeout = 10 * time.Minute // Default timeout
	}
	return nil
}

// ValidationError represents a task validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Field + ": " + e.Message
}
