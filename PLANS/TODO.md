# TODO

Items I think we should work on next, roughly in priority order.
Cross off items as they are completed; add new ones as they come up.

---

## High priority — needed to have a working end-to-end tool

- [ ] **System prompt assembly** (`internal/systemprompt`)
  Load and concatenate system prompts from all three sources: inline `text`,
  local `file` (relative to the config file), and `url` (fetched once per
  session, cached in memory). This is a prerequisite for the runner.

- [ ] **Streaming Anthropic client**
  The current `anthropic.Client.Send` does a blocking request and returns
  when done. For interactive CLI use, streaming is essential UX — the user
  should see tokens arrive in real time. Add a `Stream` method that writes
  to an `io.Writer` as chunks arrive.

- [ ] **Runner / agent loop** (`internal/runner`)
  The core of the tool. Given a loaded config and a command invocation:
  1. Assemble the full system prompt
  2. Read positional arg files and inject them as the first user message
  3. Run the streaming conversation loop (user input → LLM → output)
  4. When the LLM proposes a file write, hand off to the output handler
  Wire `cmd/run.go` up to this once it exists.

- [ ] **File tools** — read and write scoped to `file_access`
  Give the agent `read_file` and `write_file` tool-use functions. Enforce
  that every path the agent tries to access is within the declared
  `file_access.read` / `file_access.write` patterns. Reject out-of-scope
  paths with a clear error rather than silently allowing or denying.

- [ ] **`output_mode: confirm`** — diff + yes/no before writing
  When the agent produces a file write, show a unified diff and prompt the
  user to confirm before touching the filesystem. This is the default mode
  and the most important one to get right first.

---

## Medium priority — makes the tool usable in practice

- [ ] **Sample app: `commit-msg`** (`sample-apps/commit-msg/`)
  A simple single-command tool that reads `git diff --staged` and generates
  a commit message. Good first real test because it has no file writes —
  just output to stdout. Validates the system prompt + streaming pipeline
  before we tackle file-write flows. Add to sample-apps alongside the
  existing test-builder and lint-fixer.

- [ ] **Validate sample apps end-to-end**
  Once the runner exists, run `sample-apps/test-builder` and
  `sample-apps/lint-fixer` against real code as integration tests.
  Document any config schema gaps discovered during that exercise.

- [ ] **`output_mode: interactive`** — accept/reject/refine loop
  After `confirm` is working, add the richer interaction mode where the user
  can type feedback to refine a proposed change before accepting it.
  Can borrow the TUI approach from prpolish.

- [ ] **Shell tool-use execution** (`internal/tooluse`)
  When `tool_use.enabled: true`, expose the declared shell commands as
  LLM tool-use functions. Validate each invocation against the allowlist
  before executing. Capture and return stdout/stderr to the LLM.

- [ ] **Config tests** (`internal/config/config_test.go`)
  Table-driven tests for `Load`, `validate`, `CheckEnv`, `Provider`,
  `ModelID`. Use small inline YAML strings — no file fixtures needed.

- [ ] **README**
  Installation (build from source for now), usage (`tool-builder run`),
  the config file format (link to `PLANS/config-schema.md` for now),
  and requirements (`ANTHROPIC_API_KEY`).

---

## Lower priority — polish and future-proofing

- [ ] **`examples/` directory**
  Ship a handful of example `.yaml` config files in the repo so users
  have something to start from. At minimum: `commit-msg.yaml`, `gotest.yaml`.

- [ ] **Second provider: Google Gemini**
  Add `google/` under `internal/provider/` using the Google Generative AI
  Go SDK. The `provider.Provider` interface is already provider-agnostic;
  this is purely additive. Good validation that the multi-provider design
  actually works.

- [ ] **`file_access` exclude patterns**
  Allow `dir:` and `glob:` entries to have an `exclude` list (e.g., skip
  `vendor/`, `node_modules/`, `dist/`). Needed before anyone points a
  tool at a real repo with a working directory full of ignored files.

- [ ] **URL system prompt caching across runs**
  Currently URL prompts are fetched fresh every invocation. Add an optional
  `cache: always` mode that stores the fetched content in a local cache
  (keyed by URL + ETag/Last-Modified) so repeated runs don't hit the network.

- [ ] **`build` subcommand** (phase 2)
  `tool-builder build --config <file>` emits a standalone binary with the
  config and runtime embedded. Deferred until the interpreted runner is
  proven. Needs design work on how to embed prompt files and the runtime
  into the generated binary.

- [ ] **`tool-builder init`** — interactive config scaffolding
  A wizard that asks questions and writes a starter `tool.yaml`. Lowers
  the barrier for new tool authors.
