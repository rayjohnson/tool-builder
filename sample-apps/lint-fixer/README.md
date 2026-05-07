# lint-fixer

Runs `golangci-lint` on your Go project and fixes every issue with minimal,
targeted changes. Reviews each proposed file edit before writing.

## Usage

```sh
# Fix all lint issues in the current directory
tool-builder run --config tool.yaml

# Lint a specific package subtree
tool-builder run --config tool.yaml --path ./internal/...

# Fix only issues from one linter
tool-builder run --config tool.yaml --only misspell
```

Run from the root of the Go project. A `.golangci.yml` in that directory will
be read and respected — the tool reads it to understand which linters are active.

## Requirements

- `ANTHROPIC_API_KEY` set in environment
- `golangci-lint` in PATH
- `go` in PATH

## What it does

1. Runs `golangci-lint run ./...` (or the specified path)
2. Groups issues by file
3. For each file, proposes the minimal fix and shows a diff
4. You accept, reject, or give feedback on each change (interactive mode)
5. After all changes, runs lint again to confirm the project is clean

## Notes on interactive mode

`lint-fixer` uses `output_mode: interactive` so you can review every fix
and push back if the proposed change is wrong or not what you want. Type
feedback directly to refine a proposal before accepting it.

## Customizing

The system prompt in `prompts/system.md` documents how each linter's issues
are handled. If your project has specific patterns (e.g., particular gosec
rules that are always false positives for you), edit the system prompt to
document that so the tool handles them consistently.
