# tool-builder: Vision and Design

## What is tool-builder?

`tool-builder` is a command-line tool that builds new AI-powered command-line tools. It reads
a config file describing a tool's behavior, prompts, and characteristics, then runs a fully
functional agentic CLI backed by Claude.

The end goal: a developer writes a YAML config, points `tool-builder` at it, and gets a
working AI agent CLI — no boilerplate, no wiring up API clients, no CLI scaffolding.

## Problem being solved

Building AI-powered CLI tools today requires:
- Wiring up LLM API clients (streaming, retries, token management)
- Managing prompt composition and context injection
- Building the CLI scaffolding (flags, help text, subcommands)
- Implementing file read/write and shell tool-use plumbing
- Building an interactive accept/reject/refine loop for proposed changes

`tool-builder` absorbs all of that so tool authors only need to declare *what* their tool does.

## Prior art / what NOT to repeat from prpolish

`prpolish` (at `../../moovfinancial/prpolish`) was a hardcoded AI-assisted PR workflow tool.
Its agent pattern (`internal/agent/<name>/agent.go` + `prompt.md` sidecar) is worth borrowing,
but the overall structure was too tightly coupled to a single use case. `tool-builder` inverts
this: the agents and their prompts come from the config file, not from compiled-in packages.

## What a config-defined tool actually is

A config-defined tool is an **agentic CLI** — it embeds the Anthropic SDK and runs Claude as an
interactive, conversational agent. It is not a one-shot prompt runner. Think of it as a
specialized, opinionated Claude Code: it has domain knowledge baked into its system prompts,
knows which files to read and write, and knows which shell tools it's allowed to call.

Example: a `gotest` tool built with tool-builder would already know your company's Go testing
standards (from system prompts in the config), accept a file or package as its target, and
interactively generate or fix tests — asking clarifying questions, proposing changes, and writing
files when the user confirms.

The primary use case is **file editing and code generation**:
- Inputs: CLI flags, positional args, files read as context
- Outputs: files written (the agent manages edits, not stdout)
- Interaction: the user can guide the agent conversationally during a session

## Core concepts

### Config file
YAML. The complete description of a tool. Defines:
- Tool metadata: name, version, description
- Model selection and parameters
- System prompts (inline or file references) — baked-in domain knowledge
- Commands / subcommands: flags, args, per-command prompts
- Tool-use: which shell commands the agent may invoke (opt-in, explicit allowlist)

### Prompts
Prompts are first-class. They can be:
- Inline strings in the config
- References to `.md` or `.txt` files (paths relative to the config file)
- Composed: system prompt(s) encode domain knowledge; per-command prompts focus the task;
  the user provides the specific runtime request interactively

### Execution model: generate standalone binaries

`tool-builder build --config mytool.yaml [--output ./bin/mytool]`

Produces a standalone binary. Tool users only need the binary and an API key.
Tool authors need tool-builder and Go installed to build; end users need neither.

This is the primary model. The `run` command (interpreted) remains available
as a fast development loop for iterating on configs without rebuilding.

**How build works:**
1. Load and validate the config
2. Collect all referenced prompt files; fetch any `url:` sources and embed them
3. Write a temp directory containing:
   - `tool.yaml` (the config)
   - All prompt files at their relative paths
   - A generated `main.go` with `//go:embed` directives for every file
   - A `go.mod` pinned to the current tool-builder release
4. Run `go build -o <output>` inside the temp directory
5. Move the binary to the output path; clean up temp files

**What the generated binary contains:**
- The config YAML (embedded)
- All prompt file contents (embedded at build time, including URL sources)
- The tool-builder runtime as a compiled-in dependency
- Cobra CLI wired to the config's commands, flags, and args
- No reference to the config file path at runtime

**Why not ship the config file alongside the binary?**
Embedding everything means the binary is truly self-contained — one file to
distribute, no config drift, no missing prompt files.

### tool-builder as a library

Generated binaries import exactly ONE package from tool-builder:
`github.com/rayjohnson/tool-builder/pkg/runtime`

That package provides `runtime.Run(embeds Embeds, args []string) error`.
Everything else — config parsing, the agent loop, providers, prompt loading —
stays in `internal/` and is never directly imported by generated binaries.
Go's `internal/` restriction only blocks direct imports from outside the module;
`pkg/runtime` can freely import `internal/runner` etc. as transitive deps.

Package layout:
- `pkg/runtime` — **only public package**: `Run(embeds, args)` entry point
- `internal/config` — config schema and loader (unchanged)
- `internal/runner` — agent loop and file/shell tools (unchanged)
- `internal/provider` — LLM provider interface and adapters (unchanged)
- `internal/systemprompt` — prompt assembly (unchanged)
- `internal/codegen` — **new**: generates `main.go` for built binaries
- `cmd/` — tool-builder CLI commands (unchanged)

### GitHub and versioning

The module (`github.com/rayjohnson/tool-builder`) must be published and tagged
for generated binaries to resolve their dependency. Generated `go.mod` pins to
the tool-builder version that built them.

Dev builds (`version = "dev"`) emit a `replace` directive pointing to the local
checkout so `go build` works before any tag exists:
```
require github.com/rayjohnson/tool-builder v0.0.0
replace github.com/rayjohnson/tool-builder => /path/to/local/checkout
```

Released builds (`version = "v0.1.0"`) emit a normal require with no replace:
```
require github.com/rayjohnson/tool-builder v0.1.0
```

This means `tool-builder build` works in both development and production without
the user needing to think about module resolution.

### Tool-use
Whether the agent can invoke shell commands is a **config-level opt-in** (not always-on).
A tool that only reads and writes files should not have shell access. A tool that needs to
run `go test` to verify its output declares that explicitly. This scopes the blast radius and
makes capabilities auditable from the config alone.

### Distribution
- `tool-builder` itself: single binary via goreleaser (tool authors install this)
- Generated tools: single binary, distributed however the tool author chooses
  (goreleaser, homebrew, direct download, etc.)

## Key open questions

1. **Ambient context**: should a generated tool automatically read a `CLAUDE.md`
   from the working directory and inject it as additional system context?
2. **URL prompt caching at build time**: fetch once and embed, or keep as runtime
   fetch? Embedding gives a truly offline binary; runtime fetch gives fresher content.
   Decided: embed at build time for standalone guarantees; document the trade-off.

## Technology decisions

- **Language**: Go — ships as single binary, strong CLI ecosystem
- **CLI framework**: Cobra (in tool-builder CLI and in generated binaries)
- **LLM**: Anthropic Claude via the Anthropic Go SDK
- **Config parsing**: `go-yaml`
- **Release**: goreleaser
- **Code generation**: text/template for `main.go`; `go build` via `os/exec`

## Status

- 2026-05-06: Project started. Decided YAML config, file-edit output model,
  tool-use as explicit config opt-in.
- 2026-05-06: Implemented interpreted runner (`run` command) and validated
  end-to-end with commit-msg sample app.
- 2026-05-06: Switched primary model to generated binaries (`build` command).
  `run` retained for development iteration. Requires `internal/` → `pkg/` rename.
