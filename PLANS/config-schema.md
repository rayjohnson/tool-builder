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
# Format: "provider/model-id"
# Supported providers (phase 1): anthropic
# Planned: openai, google, xai
model: anthropic/claude-opus-4-7
model_params:
  max_tokens: 8096
  temperature: 0.2              # lower = more deterministic code output

# ── Environment variables ─────────────────────────────────────────────────────
# Declares env vars the tool needs at runtime.
# Required vars cause a hard error with a helpful message if missing.
# The provider's API key (e.g., ANTHROPIC_API_KEY) is always implicitly required
# and does NOT need to be listed here — tool-builder handles it automatically.
env:
  - name: GITHUB_TOKEN
    description: GitHub token for reading private repos
    required: true
  - name: GOTEST_OUTPUT_DIR
    description: Override directory for generated test files
    required: false
    default: "."

# ── System prompts ───────────────────────────────────────────────────────────
# Concatenated in order and sent as the model's system prompt.
# They encode the tool's domain knowledge — standards, idioms, conventions.
# Paths are relative to the config file (not the working directory).
# Sources: inline text, local file (relative to config), or remote URL.
system_prompts:
  - file: prompts/go_testing_standards.md
  - file: prompts/company_conventions.md
  - url: https://raw.githubusercontent.com/myorg/standards/main/go-testing.md
  - text: |
      Always use table-driven tests with a 'tests' slice of structs.
      Test function names must follow TestXxx_Scenario convention.

# ── File access ───────────────────────────────────────────────────────────────
# Declares what files in the user's working directory the agent is ALLOWED to
# read and write during the conversation. This is a capability scope, not a
# pre-load list — the agent decides what to actually read based on the task.
#
# The agent receives read_file and write_file tools scoped to these patterns.
# Absent = no file access (agent can only use information from system prompts
# and what the user types).
#
# Patterns are relative to the working directory where the tool is invoked,
# NOT relative to the config file.
file_access:
  read:
    - glob: "**/*.go"           # any Go file anywhere in the working tree
    - dir: docs/                # everything under docs/
  write:
    - glob: "**/*.go"           # agent may only write Go files

# ── Commands ─────────────────────────────────────────────────────────────────
# A tool may define one or more subcommands. If only one command is defined
# and it is named "default", tool-builder skips the subcommand and runs it
# directly (e.g., `gotest ./foo.go` instead of `gotest generate ./foo.go`).
commands:
  - name: generate
    description: Generate tests for a Go file or package
    # Per-command prompt: appended after system prompts, before user input.
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

  - name: fix
    description: Fix failing tests in a Go file
    prompt: |
      The user will provide a Go test file with failures. Diagnose and fix them.
    args:
      - name: target
        description: Go test file to fix
        required: true

# ── Tool use ─────────────────────────────────────────────────────────────────
# Opt-in. Absent = agent cannot execute any shell commands.
# Only listed commands are allowed; tool-builder validates before each execution.
tool_use:
  enabled: true
  shell:
    - command: go
      args: [test, build, vet]   # only these subcommands are permitted
    - command: golangci-lint
      args: [run]

# ── Interaction ───────────────────────────────────────────────────────────────
# How the agent presents file changes to the user.
# Options: "confirm"     — show diff, ask yes/no before writing
#          "interactive" — accept/reject/refine loop (like prpolish)
#          "direct"      — write immediately (use with care)
output_mode: confirm
```

## How file context actually works

There are three distinct sources of file content for the agent. It is important
to understand the difference:

**1. System prompts** — static domain knowledge baked into the tool itself.
These files live alongside the config (relative to the config file) and are
always loaded at startup. Examples: `prompts/go_testing_standards.md`, inline
convention text. These never come from the user's working directory.

**2. File access scope** (`file_access`) — a capability declaration. The agent
gets `read_file` and `write_file` tools scoped to these patterns. The agent
reads files *on demand* as the conversation progresses — only the files
relevant to the current task. Patterns are relative to the working directory.
Declaring `dir: .` does not read every file upfront; it means the agent *may*
read any file if it needs to.

**3. Positional args** — files the user explicitly names on the command line
(e.g., `gotest ./pkg/foo.go`). These are read immediately and injected into
the first message as the primary subject of the task. They are the agent's
starting point.

## Top-level fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Tool name (used in help text and binary name when generated) |
| `version` | string | yes | Semver version string |
| `description` | string | yes | One-line description for help text |
| `model` | string | yes | `provider/model-id` — see Model Providers below |
| `model_params` | object | no | `max_tokens`, `temperature`, etc. |
| `env` | list | no | Environment variable declarations (see below) |
| `system_prompts` | list | yes | One or more system prompt entries (see below) |
| `file_access` | object | no | `read` and `write` scope lists (see below) |
| `commands` | list | yes | One or more command definitions (see below) |
| `tool_use` | object | no | Shell tool allowlist; omit to disable all tool use |
| `output_mode` | string | no | How file changes are presented; default `confirm` |

## Model providers

The `model` field uses `provider/model-id` format. This keeps the config
self-documenting and makes adding new providers a matter of adding a new
provider adapter in tool-builder's internals without changing the schema.

| Provider prefix | Status | API key env var |
|---|---|---|
| `anthropic` | Phase 1 | `ANTHROPIC_API_KEY` |
| `google` | Planned | `GOOGLE_API_KEY` |
| `xai` | Planned | `XAI_API_KEY` |
| `openai` | Planned | `OPENAI_API_KEY` |

The provider's API key is always implicitly required and does not need to
appear in the `env` section — tool-builder checks for it automatically and
emits a clear error if missing.

## Environment variable fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Environment variable name |
| `description` | string | yes | Shown in error messages and `--help` output |
| `required` | bool | no | Hard error at startup if missing; default false |
| `default` | string | no | Value used when absent and `required` is false |

## System prompt entries

Each entry in `system_prompts` is one of:

```yaml
- text: "inline prompt text"
- file: path/to/prompt.md        # relative to the config file
- url: https://example.com/p.md  # fetched at startup; cached for the session
```

URL prompts are fetched once at tool startup and cached for the session.
Not cached across runs — every invocation fetches fresh content.

## file_access patterns

```yaml
file_access:
  read:
    - glob: "**/*.go"       # glob, relative to working directory
    - dir: src/             # entire subtree (shorthand for glob: "src/**/*")
    - dir: .                # entire working directory
  write:
    - glob: "**/*.go"
    - dir: src/
```

`read` and `write` are independent. A tool that should never write files
omits `write` entirely.

## Command fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Subcommand name (`default` to skip subcommand layer) |
| `description` | string | yes | Help text for this command |
| `prompt` | string | no | Inline prompt appended after system prompts |
| `args` | list | no | Positional arguments (files named here are read and injected upfront) |
| `flags` | list | no | Named flags |

## Flag types

Supported: `string`, `bool`, `int`, `string_slice`

## Open schema questions

- For `url` system prompts: should there be a `cache: session | always | never` option?
- Should `file_access` patterns support an `exclude` list (e.g., ignore `vendor/`, `node_modules/`)?
- Should `model_params` be provider-specific (some params don't map across providers)?
