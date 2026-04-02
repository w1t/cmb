package task

import (
	"testing"
	"time"
)

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: Task{
				Name:         "test-task",
				Repo:         "/path/to/repo",
				Instructions: "Do something",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			task: Task{
				Repo:         "/path/to/repo",
				Instructions: "Do something",
			},
			wantErr: true,
		},
		{
			name: "missing repo",
			task: Task{
				Name:         "test-task",
				Instructions: "Do something",
			},
			wantErr: true,
		},
		{
			name: "missing instructions",
			task: Task{
				Name: "test-task",
				Repo: "/path/to/repo",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskValidate_DefaultTimeout(t *testing.T) {
	task := Task{
		Name:         "test-task",
		Repo:         "/path/to/repo",
		Instructions: "Do something",
	}

	err := task.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if task.Timeout != 10*time.Minute {
		t.Errorf("Expected default timeout of 10 minutes, got %v", task.Timeout)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Message: "is required",
	}

	expected := "validation error: name: is required"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
