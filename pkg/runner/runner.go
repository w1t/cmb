package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/codematicbench/cmb/pkg/agent"
	"github.com/codematicbench/cmb/pkg/config"
	"github.com/codematicbench/cmb/pkg/task"
)

// Runner orchestrates task execution in sandboxed environments
type Runner struct {
	UseSandbox      bool
	PreserveSandbox bool
}

// NewRunner creates a new task runner
func NewRunner(useSandbox bool) *Runner {
	return &Runner{
		UseSandbox:      useSandbox,
		PreserveSandbox: false,
	}
}

// SetPreserveSandbox sets whether to preserve the sandbox after execution
func (r *Runner) SetPreserveSandbox(preserve bool) {
	r.PreserveSandbox = preserve
}

// RunTask executes a task with the specified agent
func (r *Runner) RunTask(ctx context.Context, t *task.Task, ag agent.Agent, cfg *config.AgentConfig) (*agent.Result, error) {
	var sandbox *Sandbox
	var err error

	// Create sandbox if enabled
	if r.UseSandbox {
		fmt.Println("[SANDBOX] Creating isolated sandbox environment...")
		sandbox, err = NewSandbox(t.Repo, t.InitialState)
		if err != nil {
			return nil, fmt.Errorf("failed to create sandbox: %w", err)
		}
		defer func() {
			if r.PreserveSandbox {
				fmt.Println("[PRESERVE] Sandbox preserved for inspection")
				fmt.Printf("   Location: %s\n", sandbox.WorktreePath)
				fmt.Println("   To inspect: cd " + sandbox.WorktreePath)
				fmt.Println("   To view diff: cd " + sandbox.WorktreePath + " && git diff HEAD")
				fmt.Println("   To cleanup later: cd " + sandbox.RepoPath + " && git worktree remove " + sandbox.WorktreePath)
			} else {
				fmt.Println("[CLEANUP] Cleaning up sandbox...")
				if cleanErr := sandbox.Cleanup(); cleanErr != nil {
					fmt.Printf("   [WARNING] Failed to cleanup sandbox: %v\n", cleanErr)
				} else {
					fmt.Println("   [OK] Sandbox cleaned up")
				}
			}
		}()

		fmt.Printf("   Worktree created at: %s\n", sandbox.WorktreePath)
		if t.InitialState != "" {
			fmt.Printf("   Checked out: %s\n", t.InitialState)
		} else {
			fmt.Printf("   Using branch: %s\n", sandbox.Branch)
		}
		fmt.Println()

		// Update task to use sandbox path
		taskCopy := *t
		taskCopy.Repo = sandbox.WorktreePath
		t = &taskCopy
	}

	// Set timeout if specified
	if t.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.Timeout)
		defer cancel()
	}

	// Execute agent
	fmt.Printf("[EXEC] Executing agent: %s\n", ag.Name())
	fmt.Println("   Working directory:", t.Repo)
	fmt.Println("   This may take a while...")
	fmt.Println()

	result, err := ag.Run(ctx, t, cfg)
	if err != nil {
		return result, err
	}

	fmt.Printf("[DONE] Agent execution completed in %s\n", result.Duration)
	fmt.Println()

	// Evaluate results if sandbox is enabled
	if r.UseSandbox && sandbox != nil {
		fmt.Println("[EVAL] Evaluating results...")
		evalResult, evalErr := r.evaluate(t, sandbox)
		if evalErr != nil {
			// Don't fail the whole run on evaluation errors
			fmt.Printf("   [WARNING] Evaluation error: %v\n", evalErr)
			result.Error = fmt.Sprintf("evaluation error: %v", evalErr)
		} else {
			result.Evaluation = evalResult
			fmt.Printf("   Files modified: %d\n", evalResult.FilesModified)
			fmt.Printf("   Lines added: %d\n", evalResult.LinesAdded)
			fmt.Printf("   Lines deleted: %d\n", evalResult.LinesDeleted)
			if t.Evaluation.RunTests != "" {
				if evalResult.TestsPassed {
					fmt.Println("   [PASS] Tests passed")
				} else {
					fmt.Println("   [FAIL] Tests failed")
				}
			}
			// Update success based on evaluation
			if t.Evaluation.RunTests != "" {
				result.Success = result.Success && evalResult.TestsPassed
			}
		}
		fmt.Println()
	}

	return result, nil
}

// evaluate runs the evaluation criteria on the sandbox
func (r *Runner) evaluate(t *task.Task, sandbox *Sandbox) (*agent.EvalResult, error) {
	result := &agent.EvalResult{}

	// Get diff statistics
	fmt.Println("   Analyzing code changes...")
	files, added, deleted, err := sandbox.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	result.FilesModified = files
	result.LinesAdded = added
	result.LinesDeleted = deleted

	// Capture the actual diff
	diff, err := sandbox.GetDiff()
	if err != nil {
		// Don't fail evaluation if we can't get diff, just log it
		fmt.Printf("   ⚠️  Warning: Could not capture diff: %v\n", err)
	} else {
		result.Diff = diff
	}

	// Run tests if specified
	if t.Evaluation.RunTests != "" {
		fmt.Printf("   Running tests: %s\n", t.Evaluation.RunTests)
		testsPassed, output, err := runTests(sandbox.WorktreePath, t.Evaluation.RunTests)
		result.TestsPassed = testsPassed
		result.TestOutput = output
		if err != nil {
			return result, fmt.Errorf("test execution failed: %w", err)
		}
	} else {
		// If no tests specified, assume passed
		result.TestsPassed = true
	}

	// Run custom evaluation command if specified
	if t.Evaluation.CustomCmd != "" {
		// TODO: Implement custom command evaluation
	}

	return result, nil
}

// RunMultiple executes the same task multiple times and returns all results
func (r *Runner) RunMultiple(ctx context.Context, t *task.Task, ag agent.Agent, cfg *config.AgentConfig, runs int) ([]*agent.Result, error) {
	results := make([]*agent.Result, 0, runs)

	for i := 0; i < runs; i++ {
		result, err := r.RunTask(ctx, t, ag, cfg)
		if err != nil {
			// Continue with other runs even if one fails
			if result == nil {
				result = &agent.Result{
					Agent:     ag.Name(),
					Task:      t.Name,
					Success:   false,
					Error:     err.Error(),
					StartTime: time.Now(),
					EndTime:   time.Now(),
				}
			}
		}
		results = append(results, result)
	}

	return results, nil
}
