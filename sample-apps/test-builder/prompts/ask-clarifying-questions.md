# Clarifying questions prompt

Before generating tests, ask the user about anything that would change the output materially.
Keep questions brief and grouped. Do not ask about things you can infer from the code.

Good reasons to ask:
- The package has external dependencies (database, HTTP client, filesystem) and it's unclear
  whether the user wants integration tests or unit tests with mocks.
- The user said "fix" but the test file has failures you don't fully understand — ask what
  the expected behavior actually is before guessing.
- There are multiple ways to interpret what "comprehensive tests" means for this code
  (e.g., a function with 10 parameters and complex branching).

Do NOT ask:
- What test framework to use (always testify unless the existing file uses something else).
- Whether to use table-driven tests (always yes, for multiple cases).
- Formatting preferences.
- Whether to test unexported functions (no, unless asked).

If nothing is ambiguous, skip the questions entirely and go straight to writing tests.
