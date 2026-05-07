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

### Execution model: interpreted first, generated later

**Phase 1 — Interpreted**: `tool-builder run --config mytool.yaml [args]`
tool-builder stays in the loop at runtime. Users must have tool-builder installed.
Lets us prove the config schema end-to-end before committing to codegen.

**Phase 2 — Generated**: `tool-builder build --config mytool.yaml`
Emits a standalone binary with flags, help text, and the runtime agent embedded.
No tool-builder needed to run the generated tool. Deferred until the config schema is proven.

### Tool-use
Whether the agent can invoke shell commands is a **config-level opt-in** (not always-on).
A tool that only reads and writes files should not have shell access. A tool that needs to
run `go test` to verify its output declares that explicitly. This scopes the blast radius and
makes capabilities auditable from the config alone.

### Distribution
`tool-builder` itself ships as a single binary (Go + goreleaser).
Phase 1 tools require tool-builder to be installed at runtime.
Phase 2 (generated) tools are self-contained.

## Key open questions

1. **Interaction model for file changes**: how does the agent present proposed edits before
   writing? Options:
   - Accept/reject/refine loop (like prpolish's patch viewer)
   - Show diff, prompt for confirmation, then write
   - Write directly with undo support
2. **Ambient context**: should a running tool automatically pick up a `CLAUDE.md` from the
   working directory to inject additional project context?
3. **Multi-command tools**: does phase 1 need subcommands, or is a single-command tool enough
   to validate the schema?

## Technology decisions

- **Language**: Go — ships as single binary, strong CLI ecosystem
- **CLI framework**: Cobra
- **LLM**: Anthropic Claude via the Anthropic Go SDK
- **Config parsing**: `go-yaml` (viper is overkill for a single config file)
- **Release**: goreleaser

## Phase 1 goals

1. Repo scaffolding: go module, Cobra root command, Makefile, goreleaser config
2. Config schema: define and document the YAML format (see `config-schema.md`)
3. Interpreter: `tool-builder run --config <file> [args]` runs a config-defined agent
4. Working example: a `gotest` config that generates Go tests for a given file

## What to borrow from prpolish

- `internal/claude/client.go` — streaming Claude client wrapper
- `internal/agent/agent.go` — agent abstraction (input → prompt → LLM → output)
- `internal/agent/<name>/prompt.md` — prompt-as-sidecar-file pattern
- `.goreleaser.yaml` — release pipeline

## Status

- 2026-05-06: Project started from scratch. Design phase. No code yet.
  Decided: YAML config, interpreted execution model first, file-edit output model,
  tool-use as explicit opt-in in config.
