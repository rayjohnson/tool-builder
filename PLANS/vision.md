# tool-builder: Vision and Design

## What is tool-builder?

`tool-builder` is a command-line tool that builds new AI-powered command-line tools. It reads
a config file describing a tool's behavior, prompts, and characteristics, then generates (or runs) a
fully functional CLI tool backed by an LLM (Claude).

The end goal: a developer downloads `tool-builder` as a single binary, points it at a config file,
and gets a new CLI tool — no boilerplate, no wiring up API clients, just a config.

## Problem being solved

Building AI-powered CLI tools today requires:
- Wiring up LLM API clients
- Managing prompts and context
- Handling streaming output, token limits, retries
- Building the CLI scaffolding (flags, help text, subcommands)

`tool-builder` absorbs all of that so tool authors only need to declare *what* their tool does.

## Prior art / what NOT to repeat from prpolish

`prpolish` (at `../../moovfinancial/prpolish`) was a hardcoded AI-assisted PR workflow tool.
Its agent pattern (`internal/agent/<name>/agent.go` + `prompt.md` sidecar) is worth borrowing,
but the overall structure was too tightly coupled to a single use case. `tool-builder` inverts
this: the "agents" and their prompts come from the config file, not from compiled-in packages.

## Core concepts

### Config file
The entry point for every tool built with `tool-builder`. Likely YAML or TOML. Defines:
- Tool metadata: name, version, description
- Commands / subcommands and their behavior
- Prompts (inline or file references) that drive LLM behavior
- Input/output shape (stdin, flags, file args, stdout, etc.)
- Model selection and parameters
- Optional: tool-use definitions (shell commands the LLM can invoke)

### Prompts
Prompts are first-class citizens. They can be:
- Inline strings in the config
- References to `.md` or `.txt` files (relative to the config)
- Composed from multiple pieces (system prompt + user prompt template)

### Generated vs. interpreted tool
Two possible execution models (TBD which to start with):
1. **Interpreted**: `tool-builder run --config mytool.yaml <args>` — `tool-builder` reads the config
   and runs the tool on the fly. Simple to iterate on.
2. **Generated**: `tool-builder build --config mytool.yaml` — emits a standalone binary or script.
   Harder but delivers on the "single download" promise for the built tool.

Starting with the interpreted model is lower risk and lets us prove the config schema before
committing to a code-gen approach.

### Distribution
`tool-builder` itself should ship as a single binary (Go, goreleaser). Tools built with it can
either require `tool-builder` to be installed, or (if we go the generated route) be fully
standalone.

## Key open questions

1. Config format: YAML vs TOML vs custom DSL?
2. Interpreted first or generated first?
3. How to handle multi-turn / conversational tools vs. one-shot tools?
4. How much tool-use (shell exec, file read, etc.) should be exposed to config-defined tools?
5. How does a config-defined tool get context from its environment (git repo, CI, etc.)?
6. Authentication: tools built with `tool-builder` will need an `ANTHROPIC_API_KEY` — how is this
   communicated/documented?

## Technology decisions (tentative)

- **Language**: Go (consistent with prpolish, ships as single binary, good CLI ecosystem)
- **CLI framework**: Cobra (standard in Go ecosystem, used by prpolish)
- **LLM**: Anthropic Claude via the Anthropic Go SDK
- **Config parsing**: `viper` or `go-yaml` (TBD)
- **Release**: goreleaser (can reuse `.goreleaser.yaml` pattern from prpolish)

## What to build first (phase 1 goals)

1. Repo scaffolding: go module, Cobra root command, Makefile, goreleaser config
2. Config schema: define and document the config file format
3. Interpreter: `tool-builder run --config <file> [args]` executes a config-defined tool
4. Simple example tool: a config that defines a one-shot prompt-based CLI (e.g., a `commit-msg`
   generator) to validate the whole pipeline end-to-end

## What prpolish patterns are worth borrowing

- `internal/claude/client.go` — streaming Claude client wrapper
- `internal/agent/agent.go` — the agent abstraction (input → prompt → LLM → structured output)
- `internal/agent/<name>/prompt.md` — prompt-as-sidecar-file pattern
- `.goreleaser.yaml` — release pipeline config

## Status

- 2026-05-06: Project started from scratch. Discussing design. No code yet.
