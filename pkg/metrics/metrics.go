package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/codematicbench/cmb/pkg/agent"
)

// TestMetrics contains parsed test execution results
type TestMetrics struct {
	Passed  int
	Failed  int
	Skipped int
	Total   int
}

// AggregateMetrics contains aggregated statistics across multiple runs
type AggregateMetrics struct {
	TotalRuns       int
	SuccessRate     float64
	AvgDuration     time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	AvgFilesModded  float64
	AvgLinesChanged float64
	AvgCost         float64
	TotalCost       float64
}

// ComparisonMetrics contains metrics for comparing multiple agents
type ComparisonMetrics struct {
	Agents     []string
	Statistics map[string]*AggregateMetrics
}

// CalculateCost estimates the API cost based on token usage and model
func CalculateCost(model string, inputTokens, outputTokens int) float64 {
	// Pricing as of 2026 (approximations)
	var inputCost, outputCost float64

	modelLower := strings.ToLower(model)
	switch {
	case strings.Contains(modelLower, "opus"):
		inputCost = 15.0 / 1_000_000  // $15 per 1M input tokens
		outputCost = 75.0 / 1_000_000 // $75 per 1M output tokens
	case strings.Contains(modelLower, "sonnet"):
		inputCost = 3.0 / 1_000_000   // $3 per 1M input tokens
		outputCost = 15.0 / 1_000_000 // $15 per 1M output tokens
	case strings.Contains(modelLower, "haiku"):
		inputCost = 0.25 / 1_000_000  // $0.25 per 1M input tokens
		outputCost = 1.25 / 1_000_000 // $1.25 per 1M output tokens
	case strings.Contains(modelLower, "gpt-4"):
		inputCost = 10.0 / 1_000_000  // $10 per 1M input tokens
		outputCost = 30.0 / 1_000_000 // $30 per 1M output tokens
	case strings.Contains(modelLower, "gpt-3.5"):
		inputCost = 0.5 / 1_000_000  // $0.5 per 1M input tokens
		outputCost = 1.5 / 1_000_000 // $1.5 per 1M output tokens
	default:
		// Generic estimate
		inputCost = 2.0 / 1_000_000
		outputCost = 6.0 / 1_000_000
	}

	return float64(inputTokens)*inputCost + float64(outputTokens)*outputCost
}

// ParseTestOutput extracts test statistics from test command output
func ParseTestOutput(output string) TestMetrics {
	metrics := TestMetrics{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		lineLower := strings.ToLower(line)

		// Common patterns across test frameworks
		if strings.Contains(lineLower, "passed") {
			fmt.Sscanf(line, "%d passed", &metrics.Passed)
		}
		if strings.Contains(lineLower, "failed") {
			fmt.Sscanf(line, "%d failed", &metrics.Failed)
		}
		if strings.Contains(lineLower, "skipped") {
			fmt.Sscanf(line, "%d skipped", &metrics.Skipped)
		}

		// Go test format: "PASS" or "FAIL"
		if strings.HasPrefix(line, "PASS") {
			metrics.Passed++
		}
		if strings.HasPrefix(line, "FAIL") {
			metrics.Failed++
		}

		// pytest format: "5 passed, 2 failed"
		if strings.Contains(lineLower, " passed,") {
			fmt.Sscanf(lineLower, "%d passed, %d failed", &metrics.Passed, &metrics.Failed)
		}
	}

	metrics.Total = metrics.Passed + metrics.Failed + metrics.Skipped
	return metrics
}

// CalculateCodeQuality estimates code quality based on static analysis
// TODO: Implement actual static analysis (golangci-lint, pylint, etc.)
func CalculateCodeQuality(workdir string, language string) float64 {
	// Placeholder: returns a fixed score
	// Future: integrate with linters and static analysis tools
	return 0.0
}

// CalculateAutonomy measures agent independence
// Returns a score from 0.0 (needs constant guidance) to 1.0 (fully autonomous)
// TODO: Implement based on human intervention tracking
func CalculateAutonomy(interventions int, totalSteps int) float64 {
	// Placeholder: simple calculation
	// Future: track prompts, errors, retries, and human corrections
	if totalSteps == 0 {
		return 1.0
	}
	return 1.0 - (float64(interventions) / float64(totalSteps))
}

// AggregateResults computes aggregate statistics from multiple results
func AggregateResults(results []*agent.Result) *AggregateMetrics {
	metrics := &AggregateMetrics{
		TotalRuns:   len(results),
		MinDuration: time.Hour * 24, // Start with large value
	}

	if len(results) == 0 {
		return metrics
	}

	var successCount int
	var totalDuration time.Duration
	var totalFilesModded int
	var totalLinesChanged int
	var totalCost float64

	for _, r := range results {
		if r.Success {
			successCount++
		}

		totalDuration += r.Duration
		if r.Duration < metrics.MinDuration {
			metrics.MinDuration = r.Duration
		}
		if r.Duration > metrics.MaxDuration {
			metrics.MaxDuration = r.Duration
		}

		if r.Evaluation != nil {
			totalFilesModded += r.Evaluation.FilesModified
			totalLinesChanged += r.Evaluation.LinesAdded + r.Evaluation.LinesDeleted
			totalCost += r.Evaluation.EstimatedCost
		}
	}

	metrics.SuccessRate = float64(successCount) / float64(len(results))
	metrics.AvgDuration = totalDuration / time.Duration(len(results))
	metrics.AvgFilesModded = float64(totalFilesModded) / float64(len(results))
	metrics.AvgLinesChanged = float64(totalLinesChanged) / float64(len(results))
	metrics.AvgCost = totalCost / float64(len(results))
	metrics.TotalCost = totalCost

	return metrics
}

// CompareAgents creates comparison metrics for multiple agents
func CompareAgents(resultsMap map[string][]*agent.Result) *ComparisonMetrics {
	comparison := &ComparisonMetrics{
		Agents:     make([]string, 0, len(resultsMap)),
		Statistics: make(map[string]*AggregateMetrics),
	}

	for agentName, results := range resultsMap {
		comparison.Agents = append(comparison.Agents, agentName)
		comparison.Statistics[agentName] = AggregateResults(results)
	}

	return comparison
}

// FormatDuration formats a duration for human-readable display
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// FormatComparisonTable creates a formatted table of agent comparison results
func FormatComparisonTable(comparison *ComparisonMetrics, taskName string, runsPerAgent int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Comparison Results: %s\n", taskName))
	sb.WriteString(fmt.Sprintf("Runs per agent: %d\n\n", runsPerAgent))

	// Header
	sb.WriteString("┌──────────────┬───────┬─────────────┬──────────────┬──────────────┬──────────────┬──────────────┐\n")
	sb.WriteString("│ Agent        │ Succ% │ Avg Duration│ Min Duration │ Max Duration │ Avg Lines Δ  │   Avg Cost   │\n")
	sb.WriteString("├──────────────┼───────┼─────────────┼──────────────┼──────────────┼──────────────┼──────────────┤\n")

	// Rows
	for _, agentName := range comparison.Agents {
		stats := comparison.Statistics[agentName]

		sb.WriteString(fmt.Sprintf("│ %-12s │ %5.1f │ %11s │ %12s │ %12s │ %12.0f │ $%11.4f │\n",
			truncate(agentName, 12),
			stats.SuccessRate*100,
			FormatDuration(stats.AvgDuration),
			FormatDuration(stats.MinDuration),
			FormatDuration(stats.MaxDuration),
			stats.AvgLinesChanged,
			stats.AvgCost,
		))
	}

	sb.WriteString("└──────────────┴───────┴─────────────┴──────────────┴──────────────┴──────────────┴──────────────┘\n")

	return sb.String()
}

// FormatSingleResult creates a formatted summary of a single task result
func FormatSingleResult(result *agent.Result) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Task: %s\n", result.Task))
	sb.WriteString(fmt.Sprintf("Agent: %s\n", result.Agent))
	sb.WriteString(fmt.Sprintf("Success: %v\n", result.Success))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", FormatDuration(result.Duration)))

	if result.Evaluation != nil {
		sb.WriteString("\nEvaluation:\n")
		sb.WriteString(fmt.Sprintf("  Tests Passed: %v\n", result.Evaluation.TestsPassed))
		sb.WriteString(fmt.Sprintf("  Files Modified: %d\n", result.Evaluation.FilesModified))
		sb.WriteString(fmt.Sprintf("  Lines Added: %d\n", result.Evaluation.LinesAdded))
		sb.WriteString(fmt.Sprintf("  Lines Deleted: %d\n", result.Evaluation.LinesDeleted))
		if result.Evaluation.EstimatedCost > 0 {
			sb.WriteString(fmt.Sprintf("  Estimated Cost: $%.4f\n", result.Evaluation.EstimatedCost))
		}
	}

	if result.Error != "" {
		sb.WriteString(fmt.Sprintf("\nError: %s\n", result.Error))
	}

	return sb.String()
}

// truncate truncates a string to maxLen, adding "..." if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetBestAgent returns the agent name with the highest success rate
// In case of tie, returns the one with lowest average cost
func GetBestAgent(comparison *ComparisonMetrics) string {
	if len(comparison.Agents) == 0 {
		return ""
	}

	bestAgent := comparison.Agents[0]
	bestStats := comparison.Statistics[bestAgent]

	for _, agentName := range comparison.Agents[1:] {
		stats := comparison.Statistics[agentName]

		// Higher success rate wins
		if stats.SuccessRate > bestStats.SuccessRate {
			bestAgent = agentName
			bestStats = stats
			continue
		}

		// If tied on success rate, lower cost wins
		if stats.SuccessRate == bestStats.SuccessRate && stats.AvgCost < bestStats.AvgCost {
			bestAgent = agentName
			bestStats = stats
		}
	}

	return bestAgent
}

// CalculateCostEfficiency returns a score combining success rate and cost
// Higher is better: success_rate^2 / cost
func CalculateCostEfficiency(successRate float64, avgCost float64) float64 {
	if avgCost == 0 {
		return successRate * 1000 // Avoid division by zero
	}
	// Square success rate to heavily penalize failures
	return (successRate * successRate) / avgCost
}
