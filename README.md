# tool-builder

Build AI-powered command-line tools from a YAML config file.

Define your tool's domain knowledge in system prompts, declare what files it
can read and write, and optionally give it shell commands to run. `tool-builder`
handles the LLM wiring, streaming output, file access scoping, and the
confirm-before-write loop.

## Quick start

### Install

**Option 1 — go install (requires Go 1.26+):**

```sh
go install github.com/rayjohnson/tool-builder@latest
```

**Option 2 — download a pre-built binary:**

Grab the tarball for your platform from the
[latest release](https://github.com/rayjohnson/tool-builder/releases/latest)
(`tool-builder_<version>_<os>_<arch>.tar.gz`), then extract and install:

```sh
# Example for macOS Apple Silicon — replace version and arch as needed:
tar -xz tool-builder < tool-builder_0.1.2_darwin_arm64.tar.gz
mv tool-builder /usr/local/bin/
```

**Option 3 — build from source:**

```sh
git clone https://github.com/rayjohnson/tool-builder
cd tool-builder
make install                        # builds and installs to /usr/local/bin
# or: INSTALL_DIR=~/bin make install
```

### Run a sample tool

```sh
export ANTHROPIC_API_KEY=sk-ant-...
tool-builder run --config sample-apps/commit-msg/tool.yaml
```

## How it works

Each tool is defined by a YAML config file:

```yaml
name: commit-msg
version: 0.1.0
description: Generate a Git commit message for staged changes
model: anthropic/claude-sonnet-4-6

system_prompts:
  - file: prompts/system.md       # domain knowledge baked into the tool

tool_use:
  enabled: true
  shell:
    - command: git
      args: [diff, log, status]   # only these subcommands are allowed

commands:
  - name: default
    description: Generate a commit message for staged changes
```

Run it with `tool-builder run`, or build a self-contained binary to distribute:

```sh
# Run (requires tool-builder in PATH)
tool-builder run --config commit-msg/tool.yaml

# Build a standalone binary (embeds config + prompts; no tool-builder needed at runtime)
tool-builder build --config commit-msg/tool.yaml -o ./bin/commit-msg
./bin/commit-msg
```

## Sample tools

| Tool | Description |
|---|---|
| [commit-msg](sample-apps/commit-msg/) | Generate a Git commit message from staged changes |
| [test-builder](sample-apps/test-builder/) | Generate or fix Go tests |
| [lint-fixer](sample-apps/lint-fixer/) | Run golangci-lint and fix issues |

Each sample includes a `Makefile` — `make build` produces a standalone binary ready to distribute.

## Config reference

See [`PLANS/config-schema.md`](PLANS/config-schema.md) for the full schema
with annotated examples.

Key sections:

| Field | What it does |
|---|---|
| `model` | `provider/model-id` — e.g. `anthropic/claude-opus-4-7` |
| `system_prompts` | Inline text, local files, or URLs — the tool's domain knowledge |
| `file_access` | Scopes what files the agent may read/write in the working directory |
| `tool_use` | Shell commands the agent may invoke, with a per-command allowlist |
| `output_mode` | `confirm` (diff + y/n), `interactive` (refine loop), or `direct` |
| `commands` | Subcommands with their own prompts, flags, and args |

## Requirements

- `ANTHROPIC_API_KEY` environment variable
- Any shell tools declared in your config's `tool_use` section
- Go 1.26+ (only needed for `go install`, building from source, or `tool-builder build`)

## Writing your own tool

1. Create a directory for your tool
2. Write a `tool.yaml` config
3. Write one or more prompt files (markdown works well)
4. Run with `tool-builder run --config your-tool/tool.yaml`
5. When ready to share: `tool-builder build --config your-tool/tool.yaml -o ./bin/your-tool`

See the [sample apps](sample-apps/) for complete working examples to copy from.
