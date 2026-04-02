package compare

import (
	"context"
	"fmt"

	"github.com/codematicbench/cmb/pkg/agent"
	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/metrics"
	"github.com/codematicbench/cmb/pkg/runner"
	"github.com/codematicbench/cmb/pkg/task"
)

// Comparison represents a comparison between multiple agents on a task
type Comparison struct {
	Task            string
	Agents          []string
	Runs            int
	PreserveSandbox bool                                 // Whether to preserve sandboxes after execution
	Results         map[string][]*agent.Result           // agent name -> results
	Statistics      map[string]*metrics.AggregateMetrics // agent name -> stats
}

// NewComparison creates a new comparison instance
func NewComparison(taskName string, agentNames []string, runs int, preserveSandbox bool) *Comparison {
	return &Comparison{
		Task:            taskName,
		Agents:          agentNames,
		Runs:            runs,
		PreserveSandbox: preserveSandbox,
		Results:         make(map[string][]*agent.Result),
		Statistics:      make(map[string]*metrics.AggregateMetrics),
	}
}

// Run executes the comparison across all agents
func (c *Comparison) Run(ctx context.Context, t *task.Task, configs map[string]*config.AgentConfig) error {
	r := runner.NewRunner(true) // Use sandbox
	r.SetPreserveSandbox(c.PreserveSandbox)

	for _, agentName := range c.Agents {
		// Get config for agent
		cfg, ok := configs[agentName]
		if !ok {
			cfg = config.DefaultConfig(agentName)
		}

		// Create agent
		ag, err := agent.New(agentName, cfg)
		if err != nil {
			return fmt.Errorf("failed to create agent %s: %w", agentName, err)
		}

		// Run multiple times
		results, err := r.RunMultiple(ctx, t, ag, cfg, c.Runs)
		if err != nil {
			return fmt.Errorf("failed to run agent %s: %w", agentName, err)
		}

		c.Results[agentName] = results
	}

	// Calculate statistics
	c.calculateStats()

	return nil
}

// calculateStats computes aggregate statistics for each agent
func (c *Comparison) calculateStats() {
	for agentName, results := range c.Results {
		c.Statistics[agentName] = metrics.AggregateResults(results)
	}
}

// Display formats the comparison results for output
func (c *Comparison) Display() string {
	comparison := &metrics.ComparisonMetrics{
		Agents:     c.Agents,
		Statistics: c.Statistics,
	}
	return metrics.FormatComparisonTable(comparison, c.Task, c.Runs)
}
