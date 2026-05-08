# Examples

Complete annotated tool configs from the sample apps. These are working tools you can
copy and customize.

---

## commit-msg

Generates a Git commit message for staged changes. Uses `list_select` so the user can
include unstaged files, and `confirm` before running `git commit`.

**Key design choices:**
- Uses `claude-sonnet-4-6` (fast and cheap — run on every commit)
- Single `default` command (no subcommand needed)
- File access is read-only and scoped to project convention files only
- Both `list_select` and `confirm` TUI tools enabled
- `output_mode: confirm` — but in practice commit-msg doesn't write files; output_mode
  mainly matters for tools that write source files

```yaml
name: commit-msg
version: 0.1.0
description: Generate a Git commit message for your staged changes

model: anthropic/claude-sonnet-4-6
model_params:
  max_tokens: 1024
  temperature: 0.2           # low temperature = consistent message style

system_prompts:
  - file: prompts/system.md  # commit message rules, style adaptation guidance

# Read project convention files if they exist — gives the agent context
# without requiring the user to pass them manually.
file_access:
  read:
    - glob: "CLAUDE.md"
    - glob: ".github/CONTRIBUTING.md"
    - glob: ".github/commit-conventions.md"

tool_use:
  enabled: true
  tui:
    - list_select             # used to pick unstaged files to include
    - confirm                 # used before git add and git commit
  shell:
    - command: git
      args: [diff, log, status, show, add, commit]

output_mode: confirm

commands:
  - name: default
    description: Generate a commit message for currently staged changes
    prompt: |
      Look at the staged changes and write a commit message for them.
      Follow the style conventions from the recent git log.
    flags:
      - name: hint
        short: m
        description: Optional hint about what the commit is doing or why
        type: string
```

**System prompt structure (prompts/system.md):**

The system prompt tells the agent exactly how to proceed:
1. Run `git status` to see what's staged
2. If there are unstaged files, call `list_select` to let the user pick
3. Run `git diff --staged` and `git diff <file>` for selected unstaged files
4. Run `git log --oneline -10` to learn the project's commit style
5. Write the commit message and print it
6. Call `confirm` to ask about staging selected files
7. Call `confirm` to ask about running `git commit`

The system prompt explicitly instructs: *never ask yes/no questions in plain text —
the conversation ends immediately after you stop generating text. Use the `confirm` tool
any time you need the user to make a decision.*

---

## test-builder

Generates or fixes Go tests following idiomatic best practices: table-driven tests,
testify assertions, subtests.

**Key design choices:**
- Uses `claude-opus-4-7` (complex code generation benefits from the most capable model)
- Two named commands: `generate` and `fix`
- File access: read all `.go` files, write only `*_test.go` files
- No TUI tools — the agent works autonomously and proposes changes via `confirm`
- Runs `go test` after writing to verify the tests pass

```yaml
name: test-builder
version: 0.1.0
description: Generate or fix Go tests following best practices

model: anthropic/claude-opus-4-7
model_params:
  max_tokens: 16000          # tests can be long; give it room
  temperature: 0.1           # very low — code output should be deterministic

system_prompts:
  - file: prompts/system.md  # Go testing conventions, testify usage, table-driven style

file_access:
  read:
    - glob: "**/*.go"        # read any Go file for context
  write:
    - glob: "**/*_test.go"   # only allowed to write test files

tool_use:
  enabled: true
  shell:
    - command: go
      args: [test, build, vet]   # verify generated tests compile and pass

output_mode: confirm

commands:
  - name: generate
    description: Generate tests for a Go source file
    prompt: |
      The user wants you to generate tests for the Go source file provided.
      Read the file, identify exported functions and methods, write comprehensive
      table-driven tests using testify. Run go test after writing to confirm they pass.
    args:
      - name: target
        description: Go source file to generate tests for (e.g. ./internal/foo/bar.go)
        required: true
    flags:
      - name: output
        short: o
        description: "Output file path (default: replaces .go with _test.go)"
        type: string
      - name: run
        short: r
        description: Run go test after generating and show results
        type: bool
        default: true

  - name: fix
    description: Fix failing or broken tests in an existing test file
    prompt: |
      The user wants you to fix the test file provided.
      Read both the test file and the source file it tests. Understand what is failing
      before proposing changes. Run go test before and after to show the improvement.
    args:
      - name: target
        description: Go test file to fix (e.g. ./internal/foo/bar_test.go)
        required: true
    flags:
      - name: run
        short: r
        description: Run go test before and after to show the improvement
        type: bool
        default: true
```

**Usage:**

```sh
test-builder generate ./internal/mypackage/foo.go
test-builder generate ./internal/mypackage/foo.go -o ./internal/mypackage/foo_test.go
test-builder fix ./internal/mypackage/foo_test.go
```

---

## lint-fixer

Runs `golangci-lint` and fixes every issue with minimal targeted changes. Shows each
proposed fix as a scrollable diff before writing.

**Key design choices:**
- Uses `claude-opus-4-7` (lint fixes require understanding subtle code semantics)
- Single `default` command
- Read all `.go` files plus lint config files; write any `.go` file
- `output_mode: terse` — suppresses agent narration between tool calls; only the final
  summary line is printed (all user interaction goes through TUI tools)
- `show_diff` + `confirm` TUI tools — user reviews each proposed fix via a scrollable
  colored diff, with an "accept all" escape hatch to approve remaining diffs at once
- Runs `golangci-lint --fix` first to auto-apply safe fixes, then handles the rest manually

```yaml
name: lint-fixer
version: 0.1.0
description: Run golangci-lint and fix issues with minimal targeted changes

model: anthropic/claude-opus-4-7
model_params:
  max_tokens: 16000
  temperature: 0.1

system_prompts:
  - file: prompts/system.md  # linter-specific fix strategies, what to avoid

file_access:
  read:
    - glob: "**/*.go"
    - path: .golangci.yml        # read lint config to understand active linters
    - path: .golangci.yaml
    - path: .golangci.toml
  write:
    - glob: "**/*.go"

tool_use:
  enabled: true
  tui:
    - show_diff                  # scrollable colored diff viewer with accept/reject/feedback
    - confirm                    # yes/no prompt before starting fixes
  shell:
    - command: golangci-lint
      args: [run]
    - command: go
      args: [build, vet]

output_mode: terse              # suppress agent narration; all output is through TUI tools

commands:
  - name: default
    description: Run golangci-lint on the current directory and fix all issues
    prompt: |
      Begin. Call run_golangci_lint now.
    flags:
      - name: path
        short: p
        description: Package pattern to lint (default ./...)
        type: string
        default: "./..."
      - name: config
        short: c
        description: Path to golangci-lint config file (default auto-detected)
        type: string
      - name: only
        description: Fix only issues from a specific linter (e.g. misspell, gosec)
        type: string
```

**Usage:**

```sh
lint-fixer                              # lint and fix ./...
lint-fixer --path ./internal/...        # scope to a subtree
lint-fixer --only misspell              # fix only misspell issues
lint-fixer --config .golangci-strict.yml  # use a specific lint config
```

**Diff review:** after each proposed file change you can:
- Press `y` to accept and write this file
- Press `a` to accept this file and all remaining diffs without further prompting
- Press `n` to skip this file
- Press `f` and type feedback (e.g. "fix the root cause, don't add nolint") to refine
  the proposal; the agent will revise and show the diff again

---

## Building any sample

```sh
cd sample-apps/commit-msg
make build          # produces ./bin/commit-msg
make install        # copies to ~/bin/commit-msg
```

Or build directly with tool-builder:

```sh
tool-builder build --config sample-apps/commit-msg/tool.yaml -o ./bin/commit-msg
```

The binary embeds everything. End users only need the binary and `ANTHROPIC_API_KEY`.
