package agent

import "time"

// Result represents the outcome of an agent executing a task
type Result struct {
	Agent      string        `json:"agent"`
	Task       string        `json:"task"`
	Success    bool          `json:"success"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	Evaluation *EvalResult   `json:"evaluation,omitempty"`
	Error      string        `json:"error,omitempty"`
	Output     string        `json:"output,omitempty"`
	ConfigUsed string        `json:"config_used,omitempty"`
}

// EvalResult contains the evaluation metrics for a task run
type EvalResult struct {
	TestsPassed   bool    `json:"tests_passed"`
	TestOutput    string  `json:"test_output,omitempty"`
	FilesModified int     `json:"files_modified"`
	LinesAdded    int     `json:"lines_added"`
	LinesDeleted  int     `json:"lines_deleted"`
	Diff          string  `json:"diff,omitempty"` // Git diff of changes
	CodeQuality   float64 `json:"code_quality,omitempty"`
	Autonomy      float64 `json:"autonomy,omitempty"`
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
}

// Metrics returns a summary of key metrics
func (r *Result) Metrics() map[string]interface{} {
	m := map[string]interface{}{
		"success":  r.Success,
		"duration": r.Duration.Seconds(),
	}

	if r.Evaluation != nil {
		m["tests_passed"] = r.Evaluation.TestsPassed
		m["files_modified"] = r.Evaluation.FilesModified
		m["lines_changed"] = r.Evaluation.LinesAdded + r.Evaluation.LinesDeleted
		if r.Evaluation.EstimatedCost > 0 {
			m["estimated_cost"] = r.Evaluation.EstimatedCost
		}
	}

	return m
}
