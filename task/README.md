# Task Definitions

This directory contains task definitions for CodematicBench. Tasks are repository-scale coding challenges that agents must complete.

## Task Format

Tasks are defined in YAML with the following structure:

```yaml
name: "task-name"
description: "Brief description of the task"
language: "python|go|javascript|..."
repo: "./path/to/repository"
instructions: |
  Detailed instructions for the agent.
  Multiple lines are supported.

success_criteria:
  - Criterion 1
  - Criterion 2
  - Criterion 3

evaluation:
  run_tests: "command to run tests"
  check_diff: true
  custom_cmd: "optional custom evaluation"

timeout: 600s  # Optional timeout
initial_state: "main"  # Optional git branch/commit
```

## Available Tasks

### refactor-auth.yaml
**Language:** Python
**Difficulty:** Medium
**Description:** Refactor authentication system from sessions to JWT tokens

**Skills tested:**
- Refactoring existing code
- Security best practices
- Test-driven development
- API design

### add-pagination.yaml
**Language:** Go
**Difficulty:** Medium
**Description:** Add cursor-based pagination to API endpoints

**Skills tested:**
- API design
- Database query optimization
- Edge case handling
- Documentation

### fix-race-condition.yaml
**Language:** Go
**Difficulty:** Hard
**Description:** Identify and fix race condition in concurrent code

**Skills tested:**
- Debugging
- Concurrency control
- Performance optimization
- Testing concurrent code

## Creating Custom Tasks

1. Create a new YAML file in this directory
2. Define the task using the format above
3. Ensure the repository exists and is a valid Git repository
4. Add appropriate success criteria
5. Specify how to evaluate success (tests, diff, custom command)

## Task Guidelines

### Good Tasks
- YES: Have clear, specific instructions
- YES: Include measurable success criteria
- YES: Test real-world coding skills
- YES: Can be completed in 5-15 minutes
- YES: Have automated evaluation

### Bad Tasks
- NO: Vague or ambiguous requirements
- NO: Too simple (one-line changes)
- NO: Too complex (multi-day work)
- NO: No way to verify success
- NO: Require external dependencies that are hard to set up

## Running Tasks

```bash
# Run a single agent on a task
cmb run --agent opencode --task task/refactor-auth.yaml

# Compare multiple agents
cmb run --agent opencode --agent claude-code --task task/add-pagination.yaml

# Use custom configuration
cmb run --agent opencode --task task/fix-race-condition.yaml --config config/my-config.yaml
```

## Contributing Tasks

We welcome contributions of new tasks! To contribute:

1. Create a task YAML file following the format above
2. Test it with at least one agent
3. Add it to this README
4. Submit a pull request

Good task ideas:
- Real bugs from open source projects
- Common refactoring scenarios
- Feature implementations
- Performance optimizations
- Security fixes
- Documentation improvements

## Task Repository Requirements

Each task needs a repository to work with. Requirements:

- Must be a Git repository
- Should have existing tests
- Should be in a working state initially
- Should not require external services (databases, APIs) if possible
- Should include a README explaining the codebase

## Evaluation Methods

### run_tests
Execute a test command and check for success (exit code 0).

```yaml
evaluation:
  run_tests: "pytest tests/"
```

### check_diff
Verify that code changes were made.

```yaml
evaluation:
  check_diff: true
```

### custom_cmd
Run a custom evaluation script.

```yaml
evaluation:
  custom_cmd: "./script/validate.sh"
```

## Metrics

For each task execution, CodematicBench tracks:

- **Success:** Did the agent complete the task?
- **Duration:** How long did it take?
- **Code Changes:** Files modified, lines added/deleted
- **Test Results:** Pass/fail status and output
- **Cost:** Estimated API cost

## Examples

See the included task files for examples of well-defined tasks.
