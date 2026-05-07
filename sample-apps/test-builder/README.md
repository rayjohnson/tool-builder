# test-builder

Generates or fixes Go test files following idiomatic best practices:
table-driven tests, testify assertions, subtests, and proper coverage of
happy paths and error cases.

## Usage

```sh
cd sample-apps/test-builder
make build          # produces ./bin/test-builder
make install        # copies to ~/bin/test-builder

# Then run it from any Go repo:
test-builder generate ./internal/mypackage/foo.go
test-builder fix ./internal/mypackage/foo_test.go
```

## Requirements

- `ANTHROPIC_API_KEY` set in environment
- `go` in PATH

## What it does

**generate**: Reads the target source file, identifies exported functions and
methods, and writes a `_test.go` file alongside it. Uses table-driven tests,
testify, and subtests. Runs `go test` after writing to confirm the tests pass.

**fix**: Reads an existing test file (and its corresponding source file),
diagnoses failing or non-compiling tests, and proposes targeted fixes.

## Customizing

Copy this directory and edit `prompts/system.md` to add your project's specific
conventions — internal test helpers, custom matchers, mock libraries, etc.
