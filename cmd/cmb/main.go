package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/codematicbench/cmb/pkg/agent"
	"github.com/codematicbench/cmb/pkg/compare"
	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/runner"
	"github.com/codematicbench/cmb/pkg/storage"
	"github.com/codematicbench/cmb/pkg/task"
)

var (
	// Global flags
	dbPath string

	// Run command flags
	agentsList      []string // Multiple agents supported
	taskPath        string
	configPath      string
	noSandbox       bool
	showDiff        bool
	preserveSandbox bool
	reviewWith      string
	runsCount       int // Number of runs per agent

	// Results command flags
	lastN       int
	taskName    string
	agentFilter string // For filtering results by agent
	showDiffRes bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "cmb",
	Short: "CodematicBench - Evaluate and compare AI coding agents",
	Long: `CodematicBench is an open-source framework for evaluating and comparing
AI coding agents on real-world, repository-scale tasks.`,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run one or more agents on a task",
	Long: `Execute one or more coding agents on a task and evaluate the results.

Examples:
  # Single agent, single run
  cmb run --agent opencode --task task.yaml

  # Single agent, multiple runs (test variance)
  cmb run --agent opencode --task task.yaml --runs 5

  # Multiple agents comparison
  cmb run --agent opencode --agent claude-code --task task.yaml

  # Full comparison with multiple runs
  cmb run --agent opencode --agent claude-code --task task.yaml --runs 3`,
	RunE: runTask,
}

var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "View past task results",
	Long:  `Query and display results from previous task executions.`,
	RunE:  showResults,
}

func init() {
	// Global flags
	// Default database location: $HOME/.cmb/results.db
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "." // Fallback to current directory if home not found
	}
	defaultDBPath := filepath.Join(homeDir, ".cmb", "results.db")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath, "Database path for results")

	// Run command flags
	runCmd.Flags().StringSliceVar(&agentsList, "agent", []string{}, "Agent(s) to use (can specify multiple times, required)")
	runCmd.Flags().StringVar(&taskPath, "task", "", "Path to task YAML file (required)")
	runCmd.Flags().StringVar(&configPath, "config", "", "Path to agent config file (optional)")
	runCmd.Flags().IntVar(&runsCount, "runs", 1, "Number of runs per agent")
	runCmd.Flags().BoolVar(&noSandbox, "no-sandbox", false, "Disable git worktree sandboxing")
	runCmd.Flags().BoolVar(&showDiff, "show-diff", false, "Display the code diff after execution")
	runCmd.Flags().BoolVar(&preserveSandbox, "preserve-sandbox", false, "Keep sandbox after execution for manual inspection")
	runCmd.Flags().StringVar(&reviewWith, "review-with", "", "Have another agent review the changes (e.g., --review-with claude-code)")
	runCmd.MarkFlagRequired("agent")
	runCmd.MarkFlagRequired("task")

	// Results command flags
	resultsCmd.Flags().IntVar(&lastN, "last", 10, "Number of recent results to show")
	resultsCmd.Flags().StringVar(&taskName, "task", "", "Filter by task name")
	resultsCmd.Flags().StringVar(&agentFilter, "agent", "", "Filter by agent name")
	resultsCmd.Flags().BoolVar(&showDiffRes, "show-diff", false, "Display code diffs for results")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(resultsCmd)
}

func runTask(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("=== CodematicBench Task Execution ===")
	fmt.Println()

	// Validate agents list
	if len(agentsList) == 0 {
		return fmt.Errorf("at least one agent must be specified with --agent")
	}

	// Load task
	fmt.Printf("[TASK] Loading task from: %s\n", taskPath)
	t, err := task.LoadFromFile(taskPath)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}
	fmt.Printf("   Task: %s\n", t.Name)
	fmt.Printf("   Language: %s\n", t.Language)
	fmt.Printf("   Repository: %s\n", t.Repo)
	if t.Timeout > 0 {
		fmt.Printf("   Timeout: %s\n", t.Timeout)
	}
	fmt.Println()

	// Determine mode: simple run vs comparison
	singleAgentSingleRun := len(agentsList) == 1 && runsCount == 1

	if singleAgentSingleRun {
		// Simple single run mode (original behavior)
		return runSingleAgent(ctx, agentsList[0], t)
	}

	// Multiple agents or multiple runs - use comparison mode
	return runComparison(ctx, agentsList, t)
}

func runSingleAgent(ctx context.Context, agentName string, t *task.Task) error {
	// Load or create config
	if configPath != "" {
		fmt.Printf("[CONFIG] Loading config from: %s\n", configPath)
	} else {
		fmt.Printf("[CONFIG] Using default config for: %s\n", agentName)
	}
	cfg, err := config.LoadOrDefault(configPath, agentName)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	fmt.Printf("   Model: %s (%s)\n", cfg.Model.Name, cfg.Model.Provider)
	if cfg.Model.Temperature > 0 {
		fmt.Printf("   Temperature: %.1f\n", cfg.Model.Temperature)
	}
	fmt.Println()

	// Create agent
	fmt.Printf("[AGENT] Initializing agent: %s\n", agentName)
	ag, err := agent.New(agentName, cfg)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	fmt.Println()

	// Create runner
	r := runner.NewRunner(!noSandbox)
	r.SetPreserveSandbox(preserveSandbox)

	if !noSandbox {
		fmt.Println("[SANDBOX] Sandbox mode: ENABLED (using git worktrees)")
		if preserveSandbox {
			fmt.Println("   [PRESERVE] Sandbox will be preserved after execution")
		}
	} else {
		fmt.Println("[WARNING] Sandbox mode: DISABLED (working directly in repo)")
	}
	fmt.Println()

	// Execute task
	fmt.Println("[RUN] Starting task execution...")
	fmt.Println()

	result, err := r.RunTask(ctx, t, ag, cfg)
	if err != nil {
		fmt.Printf("\n[ERROR] Execution failed: %v\n", err)
		if result != nil {
			displayResult(result)
		}
		return err
	}

	fmt.Println()

	// Display result
	displayResult(result)

	// Review with another agent if requested
	if reviewWith != "" && result.Success {
		fmt.Println()
		fmt.Printf("[REVIEW] Running code review with: %s\n", reviewWith)
		fmt.Println()

		// Temporarily preserve sandbox for review
		r.SetPreserveSandbox(true)

		// Create review task with diff in the instructions
		reviewTask := *t
		reviewTask.Name = t.Name + "-review"

		diff := ""
		if result.Evaluation != nil && result.Evaluation.Diff != "" {
			diff = result.Evaluation.Diff
		}

		// Load reviewer agent config
		reviewCfg, err := config.LoadOrDefault("", reviewWith)
		if err != nil {
			fmt.Printf("[WARNING] Failed to load reviewer config: %v\n", err)
		} else {
			// Get review template from config or use default
			reviewTemplate := reviewCfg.Prompts.Review
			if reviewTemplate == "" {
				reviewTemplate = getDefaultReviewTemplate()
			}

			// Replace placeholders in template
			reviewInstructions := strings.ReplaceAll(reviewTemplate, "{AGENT}", agentName)
			reviewInstructions = strings.ReplaceAll(reviewInstructions, "{TASK}", t.Name)
			reviewInstructions = strings.ReplaceAll(reviewInstructions, "{DIFF}", diff)
			reviewTask.Instructions = reviewInstructions

			// Create reviewer agent
			reviewAgent, err := agent.New(reviewWith, reviewCfg)
			if err != nil {
				fmt.Printf("[WARNING] Failed to create reviewer agent: %v\n", err)
			} else {
				// Run review
				reviewResult, err := r.RunTask(ctx, &reviewTask, reviewAgent, reviewCfg)
				if err != nil {
					fmt.Printf("[WARNING] Review failed: %v\n", err)
				} else {
					fmt.Println()
					fmt.Println("=== Code Review Result ===")
					fmt.Println()
					if reviewResult.Output != "" {
						fmt.Println(reviewResult.Output)
					}
					fmt.Println()

					// Save review result
					reviewResult.Task = t.Name + "-review-by-" + reviewWith
					if err := saveResult(reviewResult); err != nil {
						fmt.Printf("[WARNING] Failed to save review result: %v\n", err)
					}
				}
			}
		}

		// Restore preserve setting
		r.SetPreserveSandbox(preserveSandbox)
	}

	// Save to database
	fmt.Printf("[SAVE] Saving results to database: %s\n", dbPath)
	if err := saveResult(result); err != nil {
		fmt.Printf("[WARNING] Failed to save result: %v\n", err)
	} else {
		fmt.Println("   [OK] Results saved successfully")
	}
	fmt.Println()

	if !result.Success {
		return fmt.Errorf("task execution was not successful")
	}

	return nil
}

func runComparison(ctx context.Context, agents []string, t *task.Task) error {
	// Load configs for all agents
	configs := make(map[string]*config.AgentConfig)
	for _, agentName := range agents {
		configs[agentName] = config.DefaultConfig(agentName)
	}

	// Create comparison
	cmp := compare.NewComparison(t.Name, agents, runsCount, preserveSandbox)

	// Run comparison
	if len(agents) == 1 {
		fmt.Printf("Running agent '%s' %d times on task: %s\n", agents[0], runsCount, t.Name)
	} else {
		fmt.Printf("Comparing %d agents on task: %s\n", len(agents), t.Name)
		fmt.Printf("Runs per agent: %d\n", runsCount)
	}

	if preserveSandbox {
		fmt.Println("[PRESERVE] Sandboxes will be preserved after execution")
	}
	fmt.Println()

	if err := cmp.Run(ctx, t, configs); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Display results
	fmt.Println(cmp.Display())

	// Save all results
	for _, results := range cmp.Results {
		for _, result := range results {
			if err := saveResult(result); err != nil {
				fmt.Printf("Warning: Failed to save result: %v\n", err)
			}
		}
	}

	return nil
}
func showResults(cmd *cobra.Command, args []string) error {
	// Open database
	store, err := openStore()
	if err != nil {
		return err
	}
	defer store.Close()

	// Build filters
	filters := make(map[string]interface{})
	if taskName != "" {
		filters["task"] = taskName
	}
	if agentFilter != "" {
		filters["agent"] = agentFilter
	}

	// Query results
	results, err := store.GetResults(filters, lastN)
	if err != nil {
		return fmt.Errorf("failed to query results: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	// Display results
	fmt.Printf("Found %d results:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("%d. Agent: %s | Task: %s | Success: %v | Duration: %s\n",
			i+1, r.Agent, r.Task, r.Success, r.Duration)
		if r.Evaluation != nil {
			fmt.Printf("   Tests: %v | Files: %d | Lines: +%d/-%d\n",
				r.Evaluation.TestsPassed,
				r.Evaluation.FilesModified,
				r.Evaluation.LinesAdded,
				r.Evaluation.LinesDeleted)
		}
		if r.Error != "" {
			fmt.Printf("   Error: %s\n", r.Error)
		}

		// Show diff if requested
		if showDiffRes && r.Evaluation != nil && r.Evaluation.Diff != "" {
			fmt.Println("   Diff:")
			fmt.Println("   ```diff")
			for _, line := range strings.Split(r.Evaluation.Diff, "\n") {
				fmt.Printf("   %s\n", line)
			}
			fmt.Println("   ```")
		}
		fmt.Println()
	}

	return nil
}

func displayResult(r *agent.Result) {
	fmt.Println("=== Task Execution Result ===")
	fmt.Println()

	// Success indicator
	if r.Success {
		fmt.Println("[SUCCESS] Status: SUCCESS")
	} else {
		fmt.Println("[FAILED] Status: FAILED")
	}

	fmt.Printf("   Agent: %s\n", r.Agent)
	fmt.Printf("   Task: %s\n", r.Task)
	fmt.Printf("   Duration: %s\n", r.Duration)
	fmt.Println()

	if r.Evaluation != nil {
		fmt.Println("[CHANGES] Code Changes:")
		fmt.Printf("   Files modified: %d\n", r.Evaluation.FilesModified)
		fmt.Printf("   Lines added: +%d\n", r.Evaluation.LinesAdded)
		fmt.Printf("   Lines deleted: -%d\n", r.Evaluation.LinesDeleted)
		fmt.Printf("   Net change: %+d lines\n", r.Evaluation.LinesAdded-r.Evaluation.LinesDeleted)
		fmt.Println()

		fmt.Println("[TESTS] Test Results:")
		if r.Evaluation.TestsPassed {
			fmt.Println("   [PASS] All tests passed")
		} else {
			fmt.Println("   [FAIL] Tests failed")
		}
		if r.Evaluation.TestOutput != "" {
			fmt.Println("   Test output available in database")
		}
		fmt.Println()
	}

	if r.Error != "" {
		fmt.Println("[ERROR] Error Details:")
		fmt.Printf("   %s\n", r.Error)
		fmt.Println()
	}

	// Show diff if requested and available
	if showDiff && r.Evaluation != nil && r.Evaluation.Diff != "" {
		fmt.Println("[DIFF] Code Changes (Diff):")
		fmt.Println("```diff")
		fmt.Println(r.Evaluation.Diff)
		fmt.Println("```")
		fmt.Println()
	}
}

func saveResult(r *agent.Result) error {
	store, err := openStore()
	if err != nil {
		return err
	}
	defer store.Close()

	return store.SaveResult(r)
}

func openStore() (*storage.Store, error) {
	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	return storage.NewStore(dbPath)
}

func getDefaultReviewTemplate() string {
	return `You are reviewing code changes made by another agent ({AGENT}).

Original Task: {TASK}

Code Changes:
{DIFF}

Please provide a detailed code review with the following structure:

SCORES (1-5, where 5 is excellent):
- Code Quality: [score]/5
- Correctness: [score]/5
- Maintainability: [score]/5

DETAILED REVIEW:
1. Code Quality & Best Practices
   - What's good
   - What needs improvement

2. Correctness & Bugs
   - Potential issues
   - Edge cases

3. Security & Performance
   - Security concerns
   - Performance implications

4. Suggestions for Improvement
   - Specific recommendations
   - Code examples if applicable

OVERALL ASSESSMENT: [APPROVE / REQUEST CHANGES / REJECT]

Be constructive and specific in your feedback.`
}
