# sample-apps

Example tools built with tool-builder. Each subdirectory is a self-contained
tool: a `tool.yaml` config, prompt files, a Makefile, and a README.

These serve two purposes:
1. **Reference implementations** — copy and customize for your own tools.
2. **Test cases** — used to validate tool-builder as it is developed.

## Tools

| Tool | Description |
|---|---|
| [commit-msg](commit-msg/) | Generate a Git commit message from staged changes |
| [test-builder](test-builder/) | Generate or fix Go tests (table-driven, testify, subtests) |
| [lint-fixer](lint-fixer/) | Run golangci-lint and fix every issue with minimal targeted changes |

## Two ways to use a sample

### Run via tool-builder (quick start, no build step)

Good for trying things out or when you want to edit the prompts and see changes immediately.

```sh
# From the repo root:
tool-builder run --config sample-apps/commit-msg/tool.yaml
tool-builder run --config sample-apps/test-builder/tool.yaml generate ./path/to/file.go
tool-builder run --config sample-apps/lint-fixer/tool.yaml
```

### Build a standalone binary (for distribution)

Each sample ships with a `Makefile`. Running `make build` produces a self-contained
binary that embeds the config and all prompt files — end users only need the binary
and an `ANTHROPIC_API_KEY`, no tool-builder install required.

```sh
cd sample-apps/commit-msg
make build          # produces ./bin/commit-msg
./bin/commit-msg    # run from any git repo
```

```sh
cd sample-apps/test-builder
make build          # produces ./bin/test-builder
./bin/test-builder generate ./internal/foo/bar.go
```

## Structure of each sample

```
<tool-name>/
├── Makefile           # build and run targets
├── tool.yaml          # The tool-builder config
├── README.md          # What the tool does and how to use it
└── prompts/
    └── system.md      # The system prompt(s) — the tool's domain knowledge
```
