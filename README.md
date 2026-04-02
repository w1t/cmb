# CodematicBench

CodematicBench (`cmb`) is a local-first framework for evaluating AI coding agents on repository-scale tasks. It runs agents against real repos, captures outcomes, measures code changes and test results, and stores runs for later comparison.

The project is implemented in Go and currently supports `opencode`, `claude-code`, `codex`, `aider`, and `kiro` through their respective CLIs.

## What It Does

- Runs one agent against a task definition
- Runs the same task multiple times to measure variance
- Runs multiple agents against the same task for side-by-side comparison
- Executes work in isolated git worktrees by default
- Stores results in SQLite for later inspection

## Current Status

This repo is usable as an alpha-stage benchmarking harness. The core runner, task/config formats, sandboxing, metrics, and result storage are implemented. Some surrounding docs still describe older command shapes; the sections below reflect the current CLI.

## Requirements

- Go 1.24+
- Git
- At least one supported agent CLI installed and available in `PATH`

Depending on which agent you use, you may also need provider credentials such as `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, or AWS Bedrock credentials.

## Quick Start

Build the CLI:

```bash
go build -o cmb ./cmd/cmb
```

Run a single task:

```bash
./cmb run --agent codex --task task/test-simple.yaml
```

Run one agent multiple times:

```bash
./cmb run --agent codex --task task/test-simple.yaml --runs 3
```

Compare multiple agents on the same task:

```bash
./cmb run \
  --agent codex \
  --agent claude-code \
  --task task/test-simple.yaml \
  --runs 3
```

View saved results:

```bash
./cmb results --last 10
```

## CLI Model

The current CLI has two primary commands:

- `cmb run`: execute one or more agents on a task
- `cmb results`: query previously saved results

Comparison is handled through `cmb run` by passing multiple `--agent` flags and/or a `--runs` count.

## Repository Layout

```text
cmd/cmb/        CLI entry point
pkg/agent/      Agent integrations
pkg/runner/     Task execution and sandboxing
pkg/task/       Task loading and validation
pkg/config/     Agent configuration loading and defaults
pkg/metrics/    Aggregation and reporting
pkg/storage/    SQLite persistence
task/          Example benchmark tasks
config/        Example agent configurations
```

## Task Format

Tasks are YAML files that point at a repo, provide instructions, and define evaluation commands.

```yaml
name: "add-pagination"
language: "go"
repo: "./test-repos/chi"

instructions: |
  Add cursor-based pagination to the relevant API endpoints.

evaluation:
  run_tests: "go test ./..."
  check_diff: true

timeout: 600s
```

See [task/README.md](task/README.md) for the full task format.

## Configurations

Agent configs live in `config/` and let you tune model choice, prompts, and agent-specific settings.

```bash
./cmb run --agent codex --task task/test-simple.yaml --config config/codex-default.yaml
```

See [config/README.md](config/README.md) for details.

## Agent Execution Contract

An Agent Execution Contract (AEC) is the practical set of rules that governs how an agent operates in a given environment:

- Permitted capabilities
- Observable state
- Required preconditions
- State transitions
- Invariants

CodematicBench is intended to help you test those constraints empirically by changing the setup, rerunning the task, and observing what changed.
