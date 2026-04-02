# Agent Configuration Files

This directory contains agent configuration files for CodematicBench. Configurations allow you to customize agent behavior, model selection, and prompts.

## Configuration Format

Configurations are defined in YAML:

```yaml
name: "config-name"
agent: "agent-name"  # opencode, claude-code, codex, aider, kiro

model:
  provider: "anthropic|openai|aws-bedrock"
  name: "model-name"
  temperature: 0.0
  max_tokens: 4096

prompts:
  system: |
    Custom system prompt for the agent.
    Can be multiple lines.

  spec_template: |  # Optional, for spec-driven agents
    Specification template

settings:
  # Agent-specific settings
  mode: "build"
  auto_commit: true

context:
  max_files: 20
  include_test_files: true
  documentation_urls:
    - "./README.md"
```

## Available Configurations

### OpenCode Configurations

#### opencode-default.yaml
- **Provider:** Anthropic
- **Model:** Claude Sonnet 4
- **Focus:** Balanced performance and cost

#### opencode-gpt4.yaml
- **Provider:** OpenAI
- **Model:** GPT-4
- **Focus:** High capability, higher cost

### Claude Code Configurations

#### claude-code-default.yaml
- **Provider:** Anthropic
- **Model:** Claude Sonnet 4
- **Focus:** Standard Claude Code behavior

#### claude-code-opus.yaml
- **Provider:** Anthropic
- **Model:** Claude Opus 4
- **Focus:** Maximum capability for complex tasks

### Kiro Configurations

#### kiro-default.yaml
- **Provider:** AWS Bedrock
- **Model:** Claude Sonnet 4
- **Focus:** Spec-driven development

## Configuration Fields

### model
Defines which LLM to use.

- `provider`: API provider (anthropic, openai, google, aws-bedrock)
- `name`: Model identifier (claude-sonnet-4, gpt-4, etc.)
- `temperature`: Randomness (0.0 = deterministic, 1.0 = creative)
- `max_tokens`: Maximum response length

### prompts
Custom prompts for the agent.

- `system`: System-level instructions
- `spec_template`: Template for specifications (Kiro only)

### settings
Agent-specific settings.

- `mode`: Execution mode
- `auto_commit`: Auto-commit changes
- `spec_first`: Generate spec before coding (Kiro)
- `max_plan_depth`: Maximum task decomposition depth

### context
Context window configuration.

- `max_files`: Maximum files in context
- `include_test_files`: Include test files in context
- `persist_context`: Maintain context across runs
- `documentation_urls`: Additional documentation to include

## Using Configurations

### Single Agent
```bash
cmb run --agent opencode --task task/my-task.yaml --config config/opencode-gpt4.yaml
```

### Default Configuration
If no config is specified, the agent's default configuration is used:

```bash
# Uses opencode-default configuration
cmb run --agent opencode --task task/my-task.yaml
```

## Creating Custom Configurations

1. Copy an existing configuration file
2. Modify the fields as needed
3. Save with a descriptive name
4. Use with `--config` flag

Example custom configuration:

```yaml
name: "my-custom-config"
agent: "opencode"

model:
  provider: "anthropic"
  name: "claude-opus-4"
  temperature: 0.2  # Slightly more creative
  max_tokens: 8192

prompts:
  system: |
    You are a senior software engineer specializing in Python.
    Focus on performance optimization and Pythonic code.
    Always add comprehensive docstrings and type hints.

context:
  max_files: 30
  include_test_files: true
```

## Configuration Best Practices

### Temperature Settings
- **0.0**: Deterministic, reproducible (recommended for benchmarking)
- **0.2-0.4**: Slightly more creative, still consistent
- **0.7-1.0**: Very creative, less predictable

### System Prompts
- Be specific about coding style and priorities
- Include language-specific best practices
- Mention testing requirements
- Keep prompts focused (avoid conflicting instructions)

### Context Configuration
- Larger `max_files` = more context but higher cost
- Include test files for TDD workflows
- Add documentation for unfamiliar codebases

## Comparing Configurations

The current CLI accepts a single `--config` value per `cmb run` invocation. To compare configurations, run separate experiments against the same task and inspect the saved results with `cmb results`.

```bash
cmb run --agent opencode --task task/refactor-auth.yaml --config config/temp-0.yaml
cmb run --agent opencode --task task/refactor-auth.yaml --config config/temp-0.2.yaml
cmb results --task refactor-auth --agent opencode --last 10
```

## Environment Variables

Some configurations require environment variables:

### Anthropic (Claude)
```bash
export ANTHROPIC_API_KEY="your-key"
```

### OpenAI (GPT)
```bash
export OPENAI_API_KEY="your-key"
```

### AWS Bedrock (Kiro)
```bash
export AWS_REGION="us-east-1"
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
```

## Troubleshooting

### Invalid Provider
Make sure the provider name matches the agent's supported providers.

### Model Not Available
Verify the model name is correct and you have API access.

### API Key Issues
Check that environment variables are set correctly.

### Configuration Not Loading
Verify YAML syntax with a YAML validator.

## Examples

### Conservative Configuration
Focus on safety and correctness:

```yaml
model:
  temperature: 0.0

prompts:
  system: |
    Prioritize correctness over speed.
    Add extensive error handling.
    Write comprehensive tests.
```

### Performance-Focused Configuration
Optimize for speed:

```yaml
model:
  temperature: 0.0

prompts:
  system: |
    Focus on performance optimization.
    Use efficient algorithms and data structures.
    Profile before optimizing.
```

### Refactoring Configuration
For code cleanup tasks:

```yaml
prompts:
  system: |
    Focus on code quality and maintainability.
    Apply design patterns where appropriate.
    Improve naming and structure.
    Don't change functionality.
```
