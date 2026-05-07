# commit-msg

Generates a Git commit message for your staged changes using Claude.
Reads the diff, studies the project's commit history to match its style,
and writes a message following best practices.

## Usage

### Run directly (no build step)

```sh
# Stage your changes first, then:
tool-builder run --config tool.yaml

# With a hint about what you're doing:
tool-builder run --config tool.yaml --hint "fixing the auth regression from last PR"
```

Run from the root of the Git repository you want to commit in.

### Build a standalone binary

```sh
make build
# produces ./bin/commit-msg

# Then run it from any repo — no tool-builder required at runtime:
./bin/commit-msg
./bin/commit-msg --hint "fixing the auth regression from last PR"
```

The binary embeds the config and all prompt files. Distribute it to your team
and they only need an `ANTHROPIC_API_KEY` — no Go toolchain or tool-builder install.

## Requirements

- `ANTHROPIC_API_KEY` set in environment
- `git` in PATH
- Changes staged with `git add`
- `tool-builder` in PATH (only needed for `run` or `make build`, not the built binary)

## What it does

1. Runs `git diff --staged` to read the exact staged changes
2. Runs `git log --oneline -10` to learn the project's commit style
3. Reads `CLAUDE.md` or `.github/CONTRIBUTING.md` if present (project context)
4. Streams a commit message to your terminal

## Why sonnet, not opus?

Commit message generation is a well-scoped, low-complexity task.
`claude-sonnet-4-6` is fast and cheap enough that you can run this on
every commit without thinking about it. Change `model` in `tool.yaml`
if you want a different model.

## Customizing

Edit `prompts/system.md` to add project-specific conventions — preferred
prefixes, ticket number format, length limits, anything that matters to
your team. The system prompt already tells the agent to infer style from
`git log`, but explicit rules in the prompt take precedence.
