# sample-apps

Example tools built with tool-builder. Each subdirectory is a self-contained
tool: a `tool.yaml` config, prompt files, and a README.

These serve two purposes:
1. **Reference implementations** — copy and customize for your own tools.
2. **Test cases** — used to validate tool-builder as it is developed.

## Tools

| Tool | Description |
|---|---|
| [test-builder](test-builder/) | Generate or fix Go tests (table-driven, testify, subtests) |
| [lint-fixer](lint-fixer/) | Run golangci-lint and fix every issue with minimal targeted changes |

## Running a sample

```sh
# From the repo root:
tool-builder run --config sample-apps/test-builder/tool.yaml generate ./path/to/file.go
tool-builder run --config sample-apps/lint-fixer/tool.yaml
```

## Structure of each sample

```
<tool-name>/
├── tool.yaml          # The tool-builder config
├── README.md          # What the tool does and how to use it
└── prompts/
    └── system.md      # The system prompt(s) — the tool's domain knowledge
```
