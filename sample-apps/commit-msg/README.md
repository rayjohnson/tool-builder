# commit-msg

Generates a Git commit message for your staged changes using Claude.
Reads the diff, studies the project's commit history to match its style,
and writes a message following best practices. Offers a file picker for
unstaged changes and will stage files and commit on your behalf.

## Usage

```sh
cd sample-apps/commit-msg
make build          # produces ./bin/commit-msg
make install        # copies to ~/bin/commit-msg

# Then run it from any git repo:
commit-msg
```

## Requirements

- `ANTHROPIC_API_KEY` set in environment
- `git` in PATH

## What it does

1. Runs `git status` to see what's staged and unstaged
2. If there are unstaged files, shows a picker so you can include them
3. Reads diffs for staged changes and any files you selected
4. Runs `git log --oneline -10` to learn the project's commit style
5. Reads `CLAUDE.md` or `.github/CONTRIBUTING.md` if present
6. Writes a commit message and prints it
7. Asks if you want to stage any selected unstaged files
8. Asks if you want to run `git commit`

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
