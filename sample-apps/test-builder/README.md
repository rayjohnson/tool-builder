# test-builder

Generates or fixes Go test files following idiomatic best practices:
table-driven tests, testify assertions, subtests, and proper coverage of
happy paths and error cases.

## Usage

### Run directly (no build step)

```sh
# Generate tests for a source file
tool-builder run --config tool.yaml generate ./internal/mypackage/foo.go

# Fix a failing test file
tool-builder run --config tool.yaml fix ./internal/mypackage/foo_test.go
```

Run from the root of the Go repo you want to test.

### Build a standalone binary

```sh
make build
# produces ./bin/test-builder

# Then run it from any Go repo — no tool-builder required at runtime:
./bin/test-builder generate ./internal/mypackage/foo.go
./bin/test-builder fix ./internal/mypackage/foo_test.go
```

The binary embeds the config and all prompt files. Distribute it to your team
and they only need an `ANTHROPIC_API_KEY` — no Go toolchain or tool-builder install.

## Requirements

- `ANTHROPIC_API_KEY` set in environment
- `go` in PATH
- `tool-builder` in PATH (only needed for `run` or `make build`, not the built binary)

## What it does

**generate**: Reads the target source file, identifies exported functions and
methods, and writes a `_test.go` file alongside it. Uses table-driven tests,
testify, and subtests. Runs `go test` after writing to confirm the tests pass.

**fix**: Reads an existing test file (and its corresponding source file),
diagnoses failing or non-compiling tests, and proposes targeted fixes.

## Customizing

Copy this directory and edit `prompts/system.md` to add your project's specific
conventions — internal test helpers, custom matchers, mock libraries, etc.
