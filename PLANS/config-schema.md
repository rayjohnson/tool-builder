# tool-builder Config Schema (Draft)

This is the YAML config format that defines a tool built with `tool-builder`.
Everything about the tool's behavior comes from this file.

## Annotated example: `gotest.yaml`

```yaml
# ── Tool metadata ────────────────────────────────────────────────────────────
name: gotest
version: 1.0.0
description: Generate or fix Go tests following company standards

# ── Model ────────────────────────────────────────────────────────────────────
model: claude-opus-4-7          # any Anthropic model ID
model_params:
  max_tokens: 8096
  temperature: 0.2              # lower = more deterministic code output

# ── System prompts ───────────────────────────────────────────────────────────
# These are concatenated (in order) and sent as the Claude system prompt.
# They encode the tool's domain knowledge — company standards, idioms, etc.
# Paths are relative to this config file.
system_prompts:
  - file: prompts/go_testing_standards.md
  - file: prompts/company_conventions.md
  - text: |
      Always use table-driven tests with a 'tests' slice of structs.
      Test function names must follow TestXxx_Scenario convention.

# ── Commands ─────────────────────────────────────────────────────────────────
# A tool may define one or more subcommands. If only one command is defined
# and it is named "default", tool-builder skips the subcommand and runs it
# directly (e.g., `gotest ./foo.go` instead of `gotest generate ./foo.go`).
commands:
  - name: generate
    description: Generate tests for a Go file or package
    # Per-command prompt: appended after the system prompts, before user input.
    prompt: |
      The user will provide a Go source file. Analyze it and generate
      comprehensive tests. Ask clarifying questions if the intent is ambiguous.

    args:
      - name: target
        description: Go file or package path to generate tests for
        required: true

    flags:
      - name: fix
        short: f
        description: Fix existing tests rather than generate new ones
        type: bool
        default: false
      - name: output
        short: o
        description: Output file path (default: <target>_test.go)
        type: string

    # Files to read into context before the conversation starts.
    # In addition to explicitly listed files, files passed as positional args
    # are always read automatically.
    context_files:
      - from_args: true          # reads the file(s) passed as positional args
      # - path: docs/testing.md  # static file always included

  - name: fix
    description: Fix failing tests in a Go file
    prompt: |
      The user will provide a Go test file with failures. Diagnose and fix them.
    args:
      - name: target
        description: Go test file to fix
        required: true
    context_files:
      - from_args: true

# ── Tool use ─────────────────────────────────────────────────────────────────
# Opt-in. If this section is absent, the agent cannot execute any shell commands.
# Only listed commands are allowed; tool-builder validates before execution.
tool_use:
  enabled: true
  shell:
    - command: go
      args: [test, build, vet]   # only these subcommands are permitted
    - command: golangci-lint
      args: [run]

# ── Interaction ───────────────────────────────────────────────────────────────
# How the agent presents file changes to the user.
# Options: "confirm"  — show diff, ask yes/no before writing
#          "interactive" — accept/reject/refine loop (like prpolish)
#          "direct"   — write immediately (use with care)
output_mode: confirm
```

## Top-level fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Tool name (used in help text and binary name when generated) |
| `version` | string | yes | Semver version string |
| `description` | string | yes | One-line description for help text |
| `model` | string | yes | Anthropic model ID |
| `model_params` | object | no | `max_tokens`, `temperature`, etc. |
| `system_prompts` | list | yes | One or more system prompt entries (see below) |
| `commands` | list | yes | One or more command definitions (see below) |
| `tool_use` | object | no | Shell tool allowlist; omit to disable all tool use |
| `output_mode` | string | no | How file changes are presented; default `confirm` |

## System prompt entries

Each entry in `system_prompts` is one of:

```yaml
- text: "inline prompt text"
- file: path/to/prompt.md      # relative to the config file
```

## Command fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Subcommand name (use `default` to skip subcommand) |
| `description` | string | yes | Help text for this command |
| `prompt` | string | no | Inline prompt appended after system prompts |
| `args` | list | no | Positional arguments |
| `flags` | list | no | Named flags |
| `context_files` | list | no | Files to read into context before conversation |

## Flag types

Supported flag types: `string`, `bool`, `int`, `string_slice`

## Open schema questions

- Should `context_files` support glob patterns (e.g., `*.go` in the working dir)?
- Should there be a `max_conversation_turns` limit to prevent runaway sessions?
- Should `system_prompts` support a `url:` source for fetching from a remote repo?
- How are secrets/env vars exposed to the tool if it needs them (e.g., API keys for tool use)?
