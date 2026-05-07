# Config reference

A tool-builder config is a YAML file that fully describes a CLI tool. Every behavioral
aspect of the tool — what it knows, what it can touch, what commands it exposes — comes
from this file.

## Top-level fields

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | yes | — | Tool name; used in help text |
| `version` | string | no | — | Semver version string |
| `description` | string | no | — | One-line description for `--help` |
| `model` | string | yes | — | `provider/model-id` (see [Model](#model)) |
| `model_params` | object | no | — | Token and temperature overrides |
| `env` | list | no | — | Environment variable declarations |
| `system_prompts` | list | yes | — | The tool's domain knowledge |
| `file_access` | object | no | — | Read/write scope in the working directory |
| `tool_use` | object | no | — | Shell and TUI tool opt-ins |
| `output_mode` | string | no | `confirm` | How file writes are presented |
| `commands` | list | yes | — | Subcommands with prompts, args, and flags |

---

## model

Format: `provider/model-id`

```yaml
model: anthropic/claude-sonnet-4-6
```

The provider prefix determines which API client is used. Only `anthropic` is currently
supported. The API key for the provider must be set in the environment — tool-builder
checks for it at startup and emits a clear error if it is missing.

| Provider | Status | API key env var |
|---|---|---|
| `anthropic` | supported | `ANTHROPIC_API_KEY` |

Model IDs — current Anthropic models:
- `claude-opus-4-7` — most capable, higher cost
- `claude-sonnet-4-6` — balanced capability and speed
- `claude-haiku-4-5-20251001` — fastest, lowest cost

---

## model_params

Optional tuning for the model call.

```yaml
model_params:
  max_tokens: 8096
  temperature: 0.2
```

| Field | Type | Description |
|---|---|---|
| `max_tokens` | int | Maximum tokens in the response |
| `temperature` | float | Sampling temperature (0.0–1.0); lower = more deterministic |

---

## env

Declares environment variables the tool needs. The provider's API key does not need to
be listed here — tool-builder handles it automatically.

```yaml
env:
  - name: GITHUB_TOKEN
    description: GitHub token for reading private repos
    required: true
  - name: OUTPUT_DIR
    description: Directory for generated files
    required: false
    default: "."
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Environment variable name |
| `description` | string | yes | Shown in startup errors and `--help` |
| `required` | bool | no | Hard error at startup if missing; default false |
| `default` | string | no | Value used when unset and not required |

---

## system_prompts

A list of prompt sources assembled in order into the model's system prompt. This is the
tool's domain knowledge layer — conventions, rules, idioms the agent should follow.

```yaml
system_prompts:
  - file: prompts/system.md
  - text: |
      Always use table-driven tests with a 'tests' slice of structs.
  - url: https://raw.githubusercontent.com/myorg/standards/main/go-testing.md
```

Each entry uses exactly one key:

| Key | Description |
|---|---|
| `text` | Inline prompt text |
| `file` | Path to a prompt file, relative to the config file |
| `url` | URL fetched at build time and embedded in the binary |

Prompt sources are concatenated in the order listed. URL sources are fetched once at
`tool-builder build` time and embedded — the binary is self-contained and works offline.

At least one entry is required.

---

## file_access

Declares what files the agent may read and write in the user's working directory. Without
this section, the agent has no file access.

```yaml
file_access:
  read:
    - glob: "**/*.go"
    - dir: docs/
  write:
    - glob: "**/*.go"
```

`read` and `write` are independent lists of patterns. A tool that should never write files
omits `write` entirely.

### Pattern types

Each pattern entry uses exactly one key:

| Key | Description | Example |
|---|---|---|
| `glob` | Doublestar glob relative to the working directory | `"**/*.go"`, `"src/**/*.ts"` |
| `dir` | All files under a directory (relative to working directory) | `docs/`, `.` |

`dir: .` means the entire working directory tree. Patterns are evaluated against the
working directory where the tool is invoked, not the config file location.

---

## tool_use

Opts the agent into calling external tools. Omitting this section means the agent cannot
execute shell commands or display TUI prompts.

```yaml
tool_use:
  enabled: true
  shell:
    - command: git
      args: [diff, log, status, add, commit]
    - command: go
      args: [test, build, vet]
  tui:
    - list_select
    - confirm
```

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Must be `true` for any tool use to work |
| `shell` | list | Shell commands the agent may invoke |
| `tui` | list | Interactive TUI tools the agent may call |

### shell entries

Each shell entry declares one command and an allowlist of permitted first arguments
(subcommands or flags):

```yaml
shell:
  - command: git
    args: [diff, log, status]   # only these subcommands are permitted
```

The agent may call `git diff` and `git log` but not `git push`. Arguments after the first
are not restricted — `git diff --staged HEAD~1` is permitted if `diff` is in the args list.

| Field | Type | Required | Description |
|---|---|---|---|
| `command` | string | yes | The executable to allow |
| `args` | list of strings | yes | Permitted first arguments (subcommands) |

### tui entries

A list of TUI tool names to enable. Each name must be one of the four built-in tools:

| Name | Description |
|---|---|
| `list_select` | Interactive selection list; returns chosen items |
| `confirm` | Single-keypress yes/no prompt; returns `"yes"` or `"no"` |
| `text_input` | Full-cursor text field; returns typed text |
| `text_editor` | Opens `$EDITOR` with a draft; returns edited content |

See [tui-tools.md](tui-tools.md) for full input schemas and usage guidance.

---

## output_mode

Controls how the agent presents proposed file writes to the user.

| Value | Behavior |
|---|---|
| `confirm` | Show a unified diff and ask yes/no before writing (default) |
| `interactive` | Show a diff; user can accept, reject, or type feedback to refine |
| `direct` | Write immediately without prompting |

Default: `confirm`.

`direct` is appropriate for fully automated pipelines. `interactive` is appropriate when
the user may want to guide or correct the agent's proposed changes. `confirm` is the safe
default for most tools.

---

## commands

Defines the CLI surface of the tool. At least one command is required.

```yaml
commands:
  - name: generate
    description: Generate tests for a Go source file
    prompt: |
      The user wants you to generate tests for the provided file.
    args:
      - name: target
        description: Go source file to generate tests for
        required: true
    flags:
      - name: output
        short: o
        description: Output file path
        type: string

  - name: fix
    description: Fix failing tests in an existing test file
    args:
      - name: target
        description: Test file to fix
        required: true
```

A command named `default` means no subcommand keyword is needed — the tool runs as
`my-tool [args]` rather than `my-tool subcommand [args]`. A config with multiple commands
exposes them as subcommands.

### Command fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Subcommand name, or `default` for no subcommand |
| `description` | string | no | Help text for this command |
| `prompt` | string | no | Inline text appended to system prompts for this command |
| `args` | list | no | Positional arguments |
| `flags` | list | no | Named flags |

### args entries

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Argument name (used in help and prompt injection) |
| `description` | string | no | Help text |
| `required` | bool | no | Error if missing; default false |

Positional args that are file paths are read and injected into the first user message.
They are the agent's starting context for the task.

### flags entries

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Flag name (used as `--name` on the CLI) |
| `short` | string | no | Single-character alias (used as `-x`) |
| `description` | string | no | Help text |
| `type` | string | no | `string`, `bool`, `int`, or `string_slice`; default `string` |
| `default` | any | no | Default value when the flag is not provided |

---

## Validation rules

tool-builder validates the config at startup and at build time. A config that fails
validation produces a clear error and the tool exits immediately.

Required fields: `name`, `model`, at least one `system_prompts` entry, at least one command.

- `model` must be in `provider/model-id` format
- `output_mode` must be `confirm`, `interactive`, or `direct`
- Each command must have a `name`
- Each `tool_use.tui` entry must be one of the four known tool names
- TUI tool `enabled: true` must be set for any shell or TUI tools to be registered
