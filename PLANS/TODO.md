# TODO

Items I think we should work on next, roughly in priority order.
Cross off items as they are completed; add new ones as they come up.

---

## High priority — needed to have a working end-to-end tool

- [x] **System prompt assembly** (`internal/systemprompt`)
  Loads from `text`, `file` (relative to config), and `url` (session-cached).

- [x] **Streaming Anthropic client**
  Uses `BetaToolRunnerStreaming.AllStreaming`; prints text tokens in real time.

- [x] **Runner / agent loop** (`internal/runner`)
  Assembles system prompt, injects arg files as initial message, runs the
  streaming conversation loop via the SDK's `BetaToolRunnerStreaming`.

- [x] **File tools** — read and write scoped to `file_access`
  `read_file` and `write_file` tools with scope validation against
  `file_access.read` / `file_access.write` patterns (glob + dir supported).

- [x] **`output_mode: confirm`** — diff + yes/no before writing
  Shows old vs. new content with +/- coloring; prompts y/N before writing.

- [ ] **End-to-end smoke test**
  Run one of the sample apps against a real file and verify the full pipeline
  works: system prompt loads, file injects, LLM streams, write_file confirms.
  Known gaps to discover: error messages, edge cases in arg parsing.

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
