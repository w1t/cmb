package task

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads a task definition from a YAML file
func LoadFromFile(path string) (*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var task Task
	if err := yaml.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task YAML: %w", err)
	}

	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("task validation failed: %w", err)
	}

	return &task, nil
}

// LoadMultiple loads multiple task definitions from files
func LoadMultiple(paths []string) ([]*Task, error) {
	tasks := make([]*Task, 0, len(paths))
	for _, path := range paths {
		task, err := LoadFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load task from %s: %w", path, err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
