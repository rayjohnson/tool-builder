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
| `context` | list | no | — | Files/URLs loaded at runtime and injected into the system prompt |
| `file_access` | object | no | — | Read/write scope in the working directory |
| `tool_use` | object | no | — | Shell, TUI, and web tool opt-ins |
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

## context

A list of files or URLs loaded at **runtime** from the user's working directory and
appended to the system prompt. Use this to inject project-specific context that changes
between invocations — a `AGENTS.md` file, a local schema, a changelog, etc.

```yaml
context:
  - path: AGENTS.md          # relative to the working directory; skipped if missing
  - path: docs/schema.json
  - url: https://example.com/api-spec.md   # fetched at runtime; skipped on error
```

Each entry uses exactly one key:

| Key | Description |
|---|---|
| `path` | Path relative to the working directory. Silently skipped if the file does not exist. |
| `url` | URL fetched via HTTP GET at runtime. Silently skipped on network or HTTP errors. |

Content is formatted under a `## Project context` heading with XML-style tags naming
each source. If no sources produce content (all missing or errored), nothing is appended.

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

### Built-in file tools

When `file_access` is configured, tool-builder automatically registers these built-in
file tools for the agent:

| Tool | Registered when | Description |
|---|---|---|
| `read_file` | Any `read` pattern is configured | Read a file within the read scope |
| `write_file` | Any `write` pattern is configured | Write full file content within the write scope |
| `edit_file` | Any `write` pattern is configured | Make a targeted replacement in an existing file |

#### edit_file

`edit_file` is the preferred tool for modifying existing files. It takes an exact string
to find (`old_string`) and a replacement (`new_string`), and applies a single targeted
edit. The `old_string` must appear **exactly once** in the file — if it matches zero or
more than one time the tool returns an error instructing the agent to add more surrounding
context.

`edit_file` is always registered alongside `write_file` whenever write access is
configured. No additional config is needed.

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

A list of TUI tool names to enable. Each name must be one of the five built-in tools:

| Name | Description |
|---|---|
| `list_select` | Interactive selection list; returns chosen items |
| `confirm` | Single-keypress yes/no prompt; returns `"yes"` or `"no"` |
| `text_input` | Full-cursor text field; returns typed text |
| `text_editor` | Opens `$EDITOR` with a draft; returns edited content |
| `show_diff` | Scrollable colored unified diff viewer; returns `"accept"`, `"accept_all"`, `"reject"`, or feedback text |

See [tui-tools.md](tui-tools.md) for full input schemas and usage guidance.

### mcp entries

A list of MCP (Model Context Protocol) servers to connect to at runtime. Each server's
tools are registered alongside the built-in tools and made available to the agent.

```yaml
tool_use:
  enabled: true
  mcp:
    - name: filesystem
      command: npx
      args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
    - name: context7
      command: npx
      args: ["-y", "@upstash/context7-mcp@latest"]
      tools:
        - resolve-library-id
        - get-library-docs
    - name: my-server
      url: http://localhost:8080/mcp
```

Each entry requires either `command` (stdio subprocess) or `url` (HTTP streamable transport),
not both.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Short identifier used to prefix tool names (e.g. `filesystem` → `filesystem__read_file`) |
| `command` | string | one of | Executable to launch as a stdio MCP subprocess |
| `args` | list | no | Arguments passed to the subprocess |
| `env` | list | no | Additional `KEY=VALUE` env vars for the subprocess |
| `url` | string | one of | Base URL of an HTTP streamable MCP server |
| `tools` | list | no | Allowlist of MCP tool names to register (default: all tools the server exposes) |

**Tool naming:** every MCP tool is registered with a `servername__toolname` prefix so the agent
always knows which server a tool belongs to and collisions between servers are impossible.
For example, a server named `filesystem` that exposes `read_file` is registered as
`filesystem__read_file`.

**Subprocess servers:** the command must be available in `PATH` when the tool is run. The
subprocess is started when the agent loop begins and stopped when it exits. npx-based servers
(e.g. `npx -y @upstash/context7-mcp@latest`) work as long as Node.js is installed.

**HTTP servers:** the server must be running and reachable before the tool is invoked. Uses
the MCP Streamable HTTP transport.

See [tui-tools.md](tui-tools.md) for full input schemas and usage guidance.

---

### web entries

A list of web tool names to enable. Currently one built-in web tool is supported:

| Name | Description | Requirement |
|---|---|---|
| `fetch` | Fetches a URL and returns its content as plain text | None |

```yaml
tool_use:
  enabled: true
  web:
    - fetch
```

---

## output_mode

Controls how the agent presents proposed file writes to the user.

| Value | Behavior |
|---|---|
| `confirm` | Show a unified diff and ask yes/no before writing (default) |
| `interactive` | Show a diff; user can accept, reject, or type feedback to refine |
| `direct` | Write immediately without prompting |
| `terse` | Suppress agent text between tool calls; only the final summary line is printed |

Default: `confirm`.

`direct` is appropriate for fully automated pipelines. `interactive` is appropriate when
the user may want to guide or correct the agent's proposed changes. `confirm` is the safe
default for most tools. `terse` is appropriate for tools that manage their own progress
output through TUI tools and don't want the agent's intermediate narration to appear.

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
| `session` | bool | no | If true, after each agent turn prompt the user for a follow-up message, continuing the conversation until the user types `exit`, `quit`, or sends EOF |

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
- `output_mode` must be `confirm`, `interactive`, `direct`, or `terse`
- Each command must have a `name`
- Each `tool_use.tui` entry must be one of the five known tool names (`list_select`, `confirm`, `text_input`, `text_editor`, `show_diff`)
- Each `tool_use.web` entry must be `fetch`
- Each `tool_use.mcp` entry must have a `name` (alphanumeric/underscore/hyphen, unique), and exactly one of `command` or `url`
- TUI tool `enabled: true` must be set for any shell or TUI tools to be registered
