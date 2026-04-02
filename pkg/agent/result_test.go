package agent

import (
	"testing"
	"time"
)

func TestResult_Metrics(t *testing.T) {
	result := &Result{
		Agent:    "opencode",
		Task:     "test-task",
		Success:  true,
		Duration: 5 * time.Second,
		Evaluation: &EvalResult{
			TestsPassed:   true,
			FilesModified: 3,
			LinesAdded:    50,
			LinesDeleted:  20,
			EstimatedCost: 0.15,
		},
	}

	metrics := result.Metrics()

	if metrics["success"] != true {
		t.Errorf("Metrics success = %v, want true", metrics["success"])
	}

	if metrics["duration"] != 5.0 {
		t.Errorf("Metrics duration = %v, want 5.0", metrics["duration"])
	}

	if metrics["tests_passed"] != true {
		t.Errorf("Metrics tests_passed = %v, want true", metrics["tests_passed"])
	}

	if metrics["files_modified"] != 3 {
		t.Errorf("Metrics files_modified = %v, want 3", metrics["files_modified"])
	}

	if metrics["lines_changed"] != 70 {
		t.Errorf("Metrics lines_changed = %v, want 70", metrics["lines_changed"])
	}

	if metrics["estimated_cost"] != 0.15 {
		t.Errorf("Metrics estimated_cost = %v, want 0.15", metrics["estimated_cost"])
	}
}

func TestResult_Metrics_NoEvaluation(t *testing.T) {
	result := &Result{
		Agent:    "opencode",
		Task:     "test-task",
		Success:  false,
		Duration: 2 * time.Second,
	}

	metrics := result.Metrics()

	if metrics["success"] != false {
		t.Errorf("Metrics success = %v, want false", metrics["success"])
	}

	if metrics["duration"] != 2.0 {
		t.Errorf("Metrics duration = %v, want 2.0", metrics["duration"])
	}

	// Should not have evaluation metrics
	if _, ok := metrics["tests_passed"]; ok {
		t.Error("Metrics should not contain tests_passed without evaluation")
	}
}
