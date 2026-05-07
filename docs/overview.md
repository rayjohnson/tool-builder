# tool-builder overview

`tool-builder` compiles AI-powered CLI tools from a YAML config file into self-contained
binaries. You write a config that declares what the tool knows, what files it can touch,
and what shell commands it can run. `tool-builder build` packages that into a single binary
anyone can run — no Go, no tool-builder, no config files needed at runtime.

## The build model

```
tool.yaml + prompts/ → tool-builder build → ./bin/my-tool (standalone binary)
```

The binary embeds the config YAML and all prompt files at build time. At runtime it reads
an API key from the environment, connects to Claude, and runs the agent loop. Users of your
tool get a single file to download; they never interact with tool-builder directly.

```sh
# Tool author workflow:
tool-builder build --config my-tool/tool.yaml -o ./bin/my-tool
./bin/my-tool --help

# End user experience:
export ANTHROPIC_API_KEY=sk-ant-...
./bin/my-tool                      # just works
```

## What the agent actually does

A built tool is not a one-shot prompt runner. It is a full agentic loop backed by Claude:

1. Assembles the system prompt from all entries in `system_prompts`
2. Accepts the user's command-line invocation (subcommand, flags, args)
3. Reads any positional-arg files and injects them as the first user message
4. Enters a tool-use loop: Claude may read files, write files, run shell commands, and
   call TUI tools to pause for user input — in any order, as many times as needed
5. The conversation continues until Claude issues a final response with no tool calls

The agent decides what to read and what to do based on the system prompt and the task. It is
not driven by a hardcoded script.

## Core concepts

### System prompts — the tool's domain knowledge

System prompts are the "what this tool knows" layer. They are assembled from one or more
sources (inline text, local files, or URLs) and sent as Claude's system prompt on every
invocation. This is where you encode conventions, standards, idioms, and rules specific to
your tool's domain.

System prompts are baked into the binary at build time. URL sources are fetched once at
build time and embedded — the binary is fully self-contained and works offline.

### File access scope

`file_access` declares which files in the user's working directory the agent is allowed
to read and write. It is a capability declaration, not a pre-load list — the agent reads
files on demand as the conversation progresses.

- `read` patterns: files the agent may read (via the `read_file` tool)
- `write` patterns: files the agent may write (via the `write_file` tool)

A tool that should never write files simply omits `write`. Without `file_access`, the agent
has no file access at all.

### Commands and subcommands

Each tool defines one or more commands. A command named `default` means no subcommand is
needed on the CLI — the tool runs directly. Multiple commands become subcommands:

```sh
my-tool                       # default command
my-tool generate ./foo.go     # named command
my-tool fix ./foo_test.go     # named command
```

Each command can have its own prompt (appended after system prompts), positional args, and
flags. Positional args that name files are read and injected into the first message.

### Tool use — shell and TUI

`tool_use` opts the agent into calling external tools:

- **Shell tools** — specific shell commands with an allowlist of permitted subcommands.
  The agent may call `git diff` but not `git push` unless `push` is explicitly listed.
- **TUI tools** — interactive terminal UI components the agent calls to pause for user
  input: a selection list, a yes/no prompt, a text field, or a full editor session.

Both are off by default. Omitting `tool_use` gives the agent no shell or TUI access.

### Output mode — how file writes work

When the agent proposes writing a file, `output_mode` controls what happens:

| Mode | Behavior |
|---|---|
| `confirm` | Show a diff and ask yes/no before writing (default) |
| `interactive` | Show a diff; user can accept, reject, or give feedback to refine |
| `direct` | Write immediately without asking |

The default is `confirm`. Use `direct` only for fully automated pipelines where you trust
the agent unconditionally.

## A note on agent interaction

The agent's text output ends the conversation turn immediately. If you want the agent to
pause for a yes/no decision or collect input from the user mid-task, it must call a TUI
tool — it cannot ask a question in text and wait for a reply. System prompts for
interactive tools must instruct the agent to use `confirm` or other TUI tools instead of
asking questions in text.

See [tui-tools.md](tui-tools.md) for the full TUI tool reference.
