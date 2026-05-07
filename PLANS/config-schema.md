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
# tool-builder reads these from the environment and injects them as needed.
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
# Sources: inline text, local file (relative to config), or remote URL.
system_prompts:
  - file: prompts/go_testing_standards.md
  - file: prompts/company_conventions.md
  - url: https://raw.githubusercontent.com/myorg/standards/main/go-testing.md
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

    # Files/directories to read into context before the conversation starts.
    # Positional args are always read automatically (no need to list from_args
    # separately unless you want to be explicit).
    context_files:
      - from_args: true           # reads the file(s) passed as positional args
      - path: docs/testing.md     # static file always included
      - glob: "internal/**/*.go"  # glob pattern, relative to working directory
      - dir: .                    # entire working directory, recursively

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
| `required` | bool | no | If true, hard error at startup if missing; default false |
| `default` | string | no | Value used when the var is absent and `required` is false |

## System prompt entries

Each entry in `system_prompts` is one of:

```yaml
- text: "inline prompt text"
- file: path/to/prompt.md        # relative to the config file
- url: https://example.com/p.md  # fetched at runtime; cached for the session
```

URL prompts are fetched once at tool startup and cached for the session.
They are not cached across runs — every invocation fetches fresh content.
(Caching across runs is a future concern.)

## Command fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Subcommand name (`default` to skip subcommand layer) |
| `description` | string | yes | Help text for this command |
| `prompt` | string | no | Inline prompt appended after system prompts |
| `args` | list | no | Positional arguments |
| `flags` | list | no | Named flags |
| `context_files` | list | no | Files/dirs to read into context before conversation |

## context_files entries

```yaml
- from_args: true               # reads files passed as positional args
- path: relative/file.md        # single file, relative to working directory
- glob: "src/**/*.go"           # glob pattern, relative to working directory
- dir: .                        # all files in this directory, recursively
- dir: src/internal             # subtree rooted at a specific directory
```

`dir` is equivalent to `glob: "<dir>/**/*"` but more readable and the
intended way to say "all files here and below."

Large context loads (glob/dir on a big repo) will be flagged with a warning
and a file-count summary before the conversation starts.

## Flag types

Supported: `string`, `bool`, `int`, `string_slice`

## Open schema questions

- For `url` system prompts: should there be a `cache: session | always | never` option?
- Should `context_files` support filtering within a `dir` (e.g., ignore `vendor/`, `node_modules/`)?
- Should `model_params` be provider-specific (some params don't map across providers)?
