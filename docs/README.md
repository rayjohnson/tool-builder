# tool-builder documentation

## Contents

### [Overview](overview.md)
Start here. Explains the build model, what the agent loop actually does, and the six
core concepts: system prompts, file access, commands, tool use, output mode, and the
interaction constraint that makes TUI tools necessary.

### [Config reference](config-reference.md)
Every field in `tool.yaml` — types, required vs. optional, defaults, and validation
rules. Use this when writing or debugging a config.

### [TUI tools](tui-tools.md)
The four interactive terminal tools the agent can call mid-task to collect input from
the user: `list_select`, `confirm`, `text_input`, and `text_editor`. Covers input
schemas, return values, keyboard controls, and how to instruct the agent to use them
in system prompts.

### [Examples](examples.md)
Complete annotated `tool.yaml` configs for all three sample apps — commit-msg,
test-builder, and lint-fixer — with design rationale for each choice.
