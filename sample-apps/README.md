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

## Using a sample

Build a self-contained binary with `make build`, then install it to `~/bin` with `make install`:

```sh
cd sample-apps/commit-msg
make install        # builds and copies to ~/bin/commit-msg
commit-msg          # run from any git repo
```

```sh
cd sample-apps/test-builder
make install        # builds and copies to ~/bin/test-builder
test-builder generate ./internal/foo/bar.go
```

The binary embeds the config and all prompt files. End users only need the
binary and an `ANTHROPIC_API_KEY` — no tool-builder install required.

## Structure of each sample

```
<tool-name>/
├── Makefile           # build and install targets
├── tool.yaml          # The tool-builder config
├── README.md          # What the tool does and how to use it
└── prompts/
    └── system.md      # The system prompt(s) — the tool's domain knowledge
```
