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
sudo make install                   # builds and installs to /usr/local/bin
# or without sudo: INSTALL_DIR=~/bin make install
```

### Build and run a sample tool

```sh
export ANTHROPIC_API_KEY=sk-ant-...
cd sample-apps/commit-msg
make install        # builds and copies to ~/bin/commit-msg
commit-msg          # run from any git repo
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

Build it into a self-contained binary to distribute:

```sh
# Build a standalone binary (embeds config + prompts; no tool-builder needed at runtime)
tool-builder build --config commit-msg/tool.yaml -o ./bin/commit-msg
./bin/commit-msg
```

## Requirements

- `ANTHROPIC_API_KEY` environment variable
- Any shell tools declared in your config's `tool_use` section
- Go 1.26+ (only needed for `go install`, building from source, or `tool-builder build`)

## Documentation

See the **[docs/](docs/)** folder for the full manual — config reference, TUI tools,
and complete annotated examples.
