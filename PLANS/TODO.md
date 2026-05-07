# TODO

Items I think we should work on next, roughly in priority order.
Cross off items as they are completed; add new ones as they come up.

---

## High priority — pivot to generated binaries

- [ ] **Move `internal/` to `pkg/`**
  The runner, config, provider, and systemprompt packages must be publicly
  importable by generated binaries. Rename all four packages and update every
  import path. This is a prerequisite for everything else in this section.

- [ ] **`pkg/runtime` — entry point for generated binaries**
  New package. Provides `Run(embeds Embeds, args []string) error` where
  `Embeds` carries the pre-loaded config YAML and a map of prompt file paths
  to their byte contents. This is what every generated binary's `main()` calls.
  The runner already does this work; `runtime` is a thin public wrapper.

- [ ] **`cmd/build.go` — the `build` subcommand**
  `tool-builder build --config mytool.yaml [--output ./bin/mytool]`
  Steps:
  1. Load and validate config
  2. Collect all prompt files referenced by `file:` sources; fetch `url:` sources
  3. Write a temp directory: config, prompt files, generated `main.go`, `go.mod`
  4. Run `go build -o <output>` in the temp dir
  5. Move binary to output path; clean up temp dir

- [ ] **`main.go` code generator** (`pkg/codegen`)
  Generates the `main.go` for the built binary using `text/template`.
  Output includes:
  - `//go:embed` directives for config YAML and every prompt file
  - `main()` that calls `runtime.Run(embeds, os.Args[1:])`
  - Cobra CLI wired to the config's commands, flags, and args

- [ ] **Generated `go.mod`**
  Pin `github.com/rayjohnson/tool-builder` to the version that built it
  (injected via ldflags at release time; falls back to `main` for dev builds).
  The generated binary is then reproducible and independent of future
  tool-builder changes.

---

## Still needed from before (interpreter work)

- [x] **System prompt assembly** (`pkg/systemprompt`)
- [x] **Streaming agent loop** (`pkg/runner`)
- [x] **File tools** — read/write scoped to `file_access`
- [x] **`output_mode: confirm`**
- [x] **Shell tool-use execution**
- [x] **Sample apps: commit-msg, test-builder, lint-fixer**
- [ ] **Flag values injected into command prompt**
  The runner ignores flag values (e.g. `--hint` on commit-msg). Values
  declared in the config should be injected into the prompt context.
- [ ] **Validate sample apps end-to-end with `build`**
  Build each sample app as a binary and run it against real code.

---

## Medium priority — polish

- [ ] **`output_mode: interactive`** — accept/reject/refine loop
  After `confirm` is working, add the richer mode where the user can type
  feedback to refine a proposed change before accepting.

- [ ] **`file_access` exclude patterns**
  Allow `dir:` and `glob:` entries to have an `exclude` list (e.g. skip
  `vendor/`, `node_modules/`).

- [ ] **Ambient context: `CLAUDE.md`**
  Generated tools should optionally read a `CLAUDE.md` from the working
  directory and inject it as additional system context. Useful for project-
  specific conventions the tool author didn't know about in advance.

- [ ] **Second provider: Google Gemini**
  Add `google/` under `pkg/provider/`. Validates the multi-provider design.

---

## Lower priority

- [ ] **`tool-builder init`** — interactive config scaffolding wizard
- [ ] **URL prompt caching at build time** is currently fetch-and-embed;
  add a `cache: runtime` opt-in for tools that want fresher content.
