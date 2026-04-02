package metrics

import (
	"testing"
	"time"

	"github.com/codematicbench/cmb/pkg/agent"
)

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		wantMin      float64
		wantMax      float64
	}{
		{
			name:         "Claude Opus",
			model:        "claude-opus-4",
			inputTokens:  1000,
			outputTokens: 500,
			wantMin:      0.05,  // Approximate
			wantMax:      0.055, // Approximate
		},
		{
			name:         "Claude Sonnet",
			model:        "claude-sonnet-4",
			inputTokens:  1000,
			outputTokens: 500,
			wantMin:      0.010,
			wantMax:      0.012,
		},
		{
			name:         "GPT-4",
			model:        "gpt-4",
			inputTokens:  1000,
			outputTokens: 500,
			wantMin:      0.024,
			wantMax:      0.026,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateCost(tt.model, tt.inputTokens, tt.outputTokens)
			if cost < tt.wantMin || cost > tt.wantMax {
				t.Errorf("CalculateCost() = %v, want between %v and %v", cost, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestParseTestOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   TestMetrics
	}{
		{
			name:   "pytest format",
			output: "5 passed, 2 failed in 1.23s",
			want:   TestMetrics{Passed: 5, Failed: 2, Total: 7},
		},
		{
			name:   "go test format",
			output: "PASS\nok  \tgithub.com/test\t0.123s",
			want:   TestMetrics{Passed: 1, Failed: 0, Total: 1},
		},
		{
			name:   "empty output",
			output: "",
			want:   TestMetrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTestOutput(tt.output)
			if got.Passed != tt.want.Passed || got.Failed != tt.want.Failed {
				t.Errorf("ParseTestOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateAutonomy(t *testing.T) {
	tests := []struct {
		name          string
		interventions int
		totalSteps    int
		want          float64
	}{
		{
			name:          "fully autonomous",
			interventions: 0,
			totalSteps:    10,
			want:          1.0,
		},
		{
			name:          "half autonomous",
			interventions: 5,
			totalSteps:    10,
			want:          0.5,
		},
		{
			name:          "no steps",
			interventions: 0,
			totalSteps:    0,
			want:          1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateAutonomy(tt.interventions, tt.totalSteps)
			if got != tt.want {
				t.Errorf("CalculateAutonomy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAggregateResults(t *testing.T) {
	results := []*agent.Result{
		{
			Success:  true,
			Duration: 10 * time.Second,
			Evaluation: &agent.EvalResult{
				FilesModified: 3,
				LinesAdded:    50,
				LinesDeleted:  20,
				EstimatedCost: 0.05,
			},
		},
		{
			Success:  true,
			Duration: 20 * time.Second,
			Evaluation: &agent.EvalResult{
				FilesModified: 5,
				LinesAdded:    100,
				LinesDeleted:  30,
				EstimatedCost: 0.10,
			},
		},
		{
			Success:  false,
			Duration: 5 * time.Second,
			Evaluation: &agent.EvalResult{
				FilesModified: 0,
				LinesAdded:    0,
				LinesDeleted:  0,
				EstimatedCost: 0.02,
			},
		},
	}

	metrics := AggregateResults(results)

	if metrics.TotalRuns != 3 {
		t.Errorf("TotalRuns = %d, want 3", metrics.TotalRuns)
	}

	if metrics.SuccessRate != 2.0/3.0 {
		t.Errorf("SuccessRate = %v, want %v", metrics.SuccessRate, 2.0/3.0)
	}

	expectedAvgDuration := (10 + 20 + 5) * time.Second / 3
	if metrics.AvgDuration != expectedAvgDuration {
		t.Errorf("AvgDuration = %v, want %v", metrics.AvgDuration, expectedAvgDuration)
	}

	if metrics.MinDuration != 5*time.Second {
		t.Errorf("MinDuration = %v, want %v", metrics.MinDuration, 5*time.Second)
	}

	if metrics.MaxDuration != 20*time.Second {
		t.Errorf("MaxDuration = %v, want %v", metrics.MaxDuration, 20*time.Second)
	}

	expectedAvgLines := (70.0 + 130.0 + 0.0) / 3.0
	if metrics.AvgLinesChanged != expectedAvgLines {
		t.Errorf("AvgLinesChanged = %v, want %v", metrics.AvgLinesChanged, expectedAvgLines)
	}

	expectedTotalCost := 0.17
	if metrics.TotalCost != expectedTotalCost {
		t.Errorf("TotalCost = %v, want %v", metrics.TotalCost, expectedTotalCost)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			want:     "500ms",
		},
		{
			name:     "seconds",
			duration: 5 * time.Second,
			want:     "5.0s",
		},
		{
			name:     "minutes",
			duration: 2 * time.Minute,
			want:     "2.0m",
		},
		{
			name:     "hours",
			duration: 1*time.Hour + 30*time.Minute,
			want:     "1.5h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBestAgent(t *testing.T) {
	// Test case 1: Clear winner by success rate
	comparison := &ComparisonMetrics{
		Agents: []string{"agent1", "agent2", "agent3"},
		Statistics: map[string]*AggregateMetrics{
			"agent1": {SuccessRate: 0.8, AvgCost: 0.10},
			"agent2": {SuccessRate: 0.95, AvgCost: 0.15},
			"agent3": {SuccessRate: 0.9, AvgCost: 0.12},
		},
	}

	best := GetBestAgent(comparison)
	if best != "agent2" {
		t.Errorf("GetBestAgent() = %v, want agent2 (highest success rate)", best)
	}

	// Test case 2: Tie-breaking by cost
	comparison.Statistics["agent2"].SuccessRate = 0.9
	comparison.Statistics["agent3"].SuccessRate = 0.9

	best = GetBestAgent(comparison)
	if best != "agent3" {
		t.Errorf("GetBestAgent() = %v, want agent3 (tied success, lower cost)", best)
	}
}

func TestCalculateCostEfficiency(t *testing.T) {
	tests := []struct {
		name        string
		successRate float64
		avgCost     float64
		wantMin     float64
	}{
		{
			name:        "high success, low cost",
			successRate: 0.9,
			avgCost:     0.01,
			wantMin:     80.0,
		},
		{
			name:        "low success, low cost",
			successRate: 0.5,
			avgCost:     0.01,
			wantMin:     20.0,
		},
		{
			name:        "zero cost",
			successRate: 0.9,
			avgCost:     0.0,
			wantMin:     100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCostEfficiency(tt.successRate, tt.avgCost)
			if got < tt.wantMin {
				t.Errorf("CalculateCostEfficiency() = %v, want >= %v", got, tt.wantMin)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "no truncation needed",
			input:  "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "truncation needed",
			input:  "very long string",
			maxLen: 10,
			want:   "very lo...",
		},
		{
			name:   "exact length",
			input:  "exactly10!",
			maxLen: 10,
			want:   "exactly10!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}
