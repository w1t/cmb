package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary task file
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "test-task.yaml")

	validYAML := `name: "test-task"
description: "Test task"
language: "go"
repo: "/tmp/test-repo"
instructions: "Do something"
success_criteria:
  - "Tests pass"
evaluation:
  run_tests: "go test"
`

	if err := os.WriteFile(taskFile, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test loading valid task
	task, err := LoadFromFile(taskFile)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if task.Name != "test-task" {
		t.Errorf("Name = %q, want %q", task.Name, "test-task")
	}
	if task.Description != "Test task" {
		t.Errorf("Description = %q, want %q", task.Description, "Test task")
	}
	if task.Language != "go" {
		t.Errorf("Language = %q, want %q", task.Language, "go")
	}
}

func TestLoadFromFile_InvalidFile(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("LoadFromFile() expected error for nonexistent file, got nil")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `name: "test
this is not valid yaml
`

	if err := os.WriteFile(taskFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadFromFile(taskFile)
	if err == nil {
		t.Error("LoadFromFile() expected error for invalid YAML, got nil")
	}
}

func TestLoadMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two valid task files
	for i := 1; i <= 2; i++ {
		taskFile := filepath.Join(tmpDir, "task"+string(rune('0'+i))+".yaml")
		yaml := `name: "test-task"
repo: "/tmp/repo"
instructions: "Do something"
`
		if err := os.WriteFile(taskFile, []byte(yaml), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	paths := []string{
		filepath.Join(tmpDir, "task1.yaml"),
		filepath.Join(tmpDir, "task2.yaml"),
	}

	tasks, err := LoadMultiple(paths)
	if err != nil {
		t.Fatalf("LoadMultiple() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("LoadMultiple() returned %d tasks, want 2", len(tasks))
	}
}
