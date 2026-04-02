package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/codematicbench/cmb/pkg/agent"
	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/runner"
	"github.com/codematicbench/cmb/pkg/storage"
	"github.com/codematicbench/cmb/pkg/task"
)

// TestTaskLoadAndValidate tests the complete flow of loading and validating a task
func TestTaskLoadAndValidate(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "test-task.yaml")

	taskYAML := `name: "integration-test-task"
description: "Test task for integration"
language: "go"
repo: "/tmp/test-repo"
instructions: "Do something"
success_criteria:
  - "Tests pass"
evaluation:
  run_tests: "go test"
timeout: 300s
`

	if err := os.WriteFile(taskFile, []byte(taskYAML), 0644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}

	// Load task
	tsk, err := task.LoadFromFile(taskFile)
	if err != nil {
		t.Fatalf("Failed to load task: %v", err)
	}

	// Validate fields
	if tsk.Name != "integration-test-task" {
		t.Errorf("Task name = %q, want %q", tsk.Name, "integration-test-task")
	}

	if tsk.Timeout != 300*time.Second {
		t.Errorf("Task timeout = %v, want %v", tsk.Timeout, 300*time.Second)
	}

	if tsk.Evaluation.RunTests != "go test" {
		t.Errorf("Task evaluation.run_tests = %q, want %q", tsk.Evaluation.RunTests, "go test")
	}
}

// TestConfigLoadAndDefault tests configuration loading and defaults
func TestConfigLoadAndDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	configYAML := `name: "test-config"
agent: "opencode"
model:
  provider: "anthropic"
  name: "claude-sonnet-4"
  temperature: 0.5
  max_tokens: 2048
prompts:
  system: "Test system prompt"
`

	if err := os.WriteFile(configFile, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	cfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Model.Temperature != 0.5 {
		t.Errorf("Config temperature = %v, want 0.5", cfg.Model.Temperature)
	}

	// Test default config
	defaultCfg := config.DefaultConfig("opencode")
	if defaultCfg.Agent != "opencode" {
		t.Errorf("Default config agent = %q, want %q", defaultCfg.Agent, "opencode")
	}
}

// TestStorageSaveAndRetrieve tests SQLite storage operations
func TestStorageSaveAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := storage.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create test result
	result := &agent.Result{
		Agent:     "opencode",
		Task:      "test-task",
		Success:   true,
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		Duration:  5 * time.Minute,
		Evaluation: &agent.EvalResult{
			TestsPassed:   true,
			FilesModified: 2,
			LinesAdded:    10,
			LinesDeleted:  5,
		},
	}

	// Save result
	if err := store.SaveResult(result); err != nil {
		t.Fatalf("Failed to save result: %v", err)
	}

	// Retrieve results
	filters := map[string]interface{}{
		"task": "test-task",
	}
	results, err := store.GetResults(filters, 10)
	if err != nil {
		t.Fatalf("Failed to get results: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	retrieved := results[0]
	if retrieved.Agent != "opencode" {
		t.Errorf("Retrieved agent = %q, want %q", retrieved.Agent, "opencode")
	}
	if !retrieved.Success {
		t.Error("Retrieved result should be successful")
	}
	if retrieved.Evaluation == nil {
		t.Fatal("Retrieved result should have evaluation")
	}
	if retrieved.Evaluation.FilesModified != 2 {
		t.Errorf("Retrieved files modified = %d, want 2", retrieved.Evaluation.FilesModified)
	}
}

// TestAgentFactory tests agent creation
func TestAgentFactory(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		wantErr   bool
	}{
		{"opencode", "opencode", false},
		{"claude-code", "claude-code", false},
		{"kiro", "kiro", false},
		{"unknown", "unknown-agent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig(tt.agentName)
			ag, err := agent.New(tt.agentName, cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("agent.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && ag == nil {
				t.Error("agent.New() returned nil agent")
			}

			if !tt.wantErr && ag.Name() != tt.agentName {
				t.Errorf("agent.Name() = %q, want %q", ag.Name(), tt.agentName)
			}
		})
	}
}

// TestRunnerCreation tests runner initialization
func TestRunnerCreation(t *testing.T) {
	r := runner.NewRunner(true)
	if r == nil {
		t.Fatal("NewRunner() returned nil")
	}

	r2 := runner.NewRunner(false)
	if r2 == nil {
		t.Fatal("NewRunner(false) returned nil")
	}
}

// TestEndToEndFlow tests a complete workflow (mock)
// Note: This test doesn't actually run agents since they may not be installed
func TestEndToEndFlow(t *testing.T) {
	tmpDir := t.TempDir()

	// Create task file
	taskFile := filepath.Join(tmpDir, "task.yaml")
	taskYAML := `name: "e2e-test"
repo: "` + tmpDir + `"
instructions: "Test instructions"
evaluation:
  check_diff: true
`
	if err := os.WriteFile(taskFile, []byte(taskYAML), 0644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}

	// Load task
	tsk, err := task.LoadFromFile(taskFile)
	if err != nil {
		t.Fatalf("Failed to load task: %v", err)
	}

	// Create config
	cfg := config.DefaultConfig("opencode")

	// Create agent
	ag, err := agent.New("opencode", cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Verify agent is ready (but don't actually run it)
	if ag.Name() != "opencode" {
		t.Errorf("Agent name = %q, want %q", ag.Name(), "opencode")
	}

	// Create database
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := storage.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create mock result
	result := &agent.Result{
		Agent:     ag.Name(),
		Task:      tsk.Name,
		Success:   true,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Minute),
		Duration:  1 * time.Minute,
	}

	// Save result
	if err := store.SaveResult(result); err != nil {
		t.Fatalf("Failed to save result: %v", err)
	}

	// Retrieve results
	results, err := store.GetResults(map[string]interface{}{"task": "e2e-test"}, 10)
	if err != nil {
		t.Fatalf("Failed to get results: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestMultipleResults tests storing and querying multiple results
func TestMultipleResults(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := storage.NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create multiple results
	agents := []string{"opencode", "claude-code", "kiro"}
	for i, agentName := range agents {
		result := &agent.Result{
			Agent:     agentName,
			Task:      "multi-test",
			Success:   i%2 == 0, // Alternate success
			StartTime: time.Now().Add(-time.Duration(i) * time.Hour),
			EndTime:   time.Now(),
			Duration:  time.Duration(i+1) * time.Minute,
		}
		if err := store.SaveResult(result); err != nil {
			t.Fatalf("Failed to save result %d: %v", i, err)
		}
	}

	// Query all results
	allResults, err := store.GetResults(map[string]interface{}{}, 10)
	if err != nil {
		t.Fatalf("Failed to get all results: %v", err)
	}

	if len(allResults) != 3 {
		t.Errorf("Expected 3 results, got %d", len(allResults))
	}

	// Query successful results only
	successResults, err := store.GetResults(map[string]interface{}{"success": true}, 10)
	if err != nil {
		t.Fatalf("Failed to get successful results: %v", err)
	}

	if len(successResults) != 2 {
		t.Errorf("Expected 2 successful results, got %d", len(successResults))
	}

	// Query by agent
	opencodeResults, err := store.GetResults(map[string]interface{}{"agent": "opencode"}, 10)
	if err != nil {
		t.Fatalf("Failed to get opencode results: %v", err)
	}

	if len(opencodeResults) != 1 {
		t.Errorf("Expected 1 opencode result, got %d", len(opencodeResults))
	}
}

// Note: The following test would actually run an agent and requires proper setup
// It's commented out but serves as a template for manual testing
/*
func TestActualAgentExecution(t *testing.T) {
	// This test requires:
	// 1. Agent CLI installed (opencode, claude, kiro)
	// 2. Valid test repository
	// 3. API keys configured
	t.Skip("Skipping actual agent execution test - requires environment setup")

	ctx := context.Background()

	// Setup test repository
	// Create task
	// Run agent
	// Verify results
}
*/
