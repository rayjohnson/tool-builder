# TODO

Items to work on next, roughly in priority order.
Cross off items as they are completed; add new ones as they come up.

---

## In progress / next up

- [ ] **MCP server support**
  Allow tool-builder tools to connect to MCP (Model Context Protocol) servers
  and expose their tools to the agent. The `tool.yaml` would declare MCP servers
  to connect to (by command or URL); tool-builder discovers their tool manifests
  at runtime and wires them alongside shell/file/TUI tools. This gives generated
  tools access to the whole MCP ecosystem (filesystem, git, databases, etc.)
  without hand-coding each integration.

- [ ] **Additional TUI tools inspired by gum/charmbracelet**
  Gum (https://github.com/charmbracelet/gum) is a CLI, not a library, but the
  underlying charmbracelet stack is. Add new tool-builder TUI tools:
  - `table` — render tabular data from the agent (rows + headers → pretty table)
  - `pager` — scrollable text display (for long output like logs or file contents)
  - `filter` — fuzzy-search filter over a list; finer than `list_select`
  - `log` — emit a styled status line (info/warn/error/success) to the terminal
  Consider `charmbracelet/huh` (https://github.com/charmbracelet/huh) for
  multi-field form interactions — it IS a proper Go library (unlike gum itself).

- [ ] **LipGloss styled output tool**
  Add a `print_styled` tool (backed by charmbracelet/lipgloss) that lets the
  agent emit richly formatted, colored text to the terminal. Useful for tools
  that want a structured summary or status output beyond a plain text line.
  lipgloss is already an indirect dep; just needs a thin tool wrapper.

- [x] **Flag values injected into command prompt**
  Values declared in the config's `flags` that the user explicitly passes are
  injected into the agent's first user message. Uses `cobra.Flag.Changed` so
  only user-set flags appear — defaults are not injected.

- [ ] **Validate sample apps end-to-end with `build`**
  Build each sample app as a binary and run it against real code to confirm
  the full build-and-run pipeline works correctly.

---

## Medium priority — polish

- [ ] **`output_mode: interactive`** — accept/reject/refine loop
  Currently `interactive` behaves the same as `confirm` (diff + yes/no).
  The intended behavior: after showing the diff, let the user type feedback
  to refine the proposal before accepting. The lint-fixer sample uses this mode.

- [ ] **`file_access` exclude patterns**
  Allow `dir:` and `glob:` entries to have an `exclude` list (e.g. skip
  `vendor/`, `node_modules/`).

- [ ] **Ambient context: `CLAUDE.md`**
  Generated tools should optionally read a `CLAUDE.md` from the working
  directory and inject it as additional system context. Useful for project-
  specific conventions the tool author didn't know about in advance.

- [ ] **Second provider: Google Gemini**
  Add `google/` under `internal/provider/`. Validates the multi-provider design.

---

## Lower priority

- [ ] **Sample app: `rtfm`** — souped-up man page powered by Context7 MCP
  A `man`-like tool that answers "how do I use X?" questions with live,
  accurate documentation via the Context7 MCP server
  (https://github.com/upstash/context7). Unlike static man pages, Context7
  serves up-to-date library docs for any package. The tool takes a natural
  language query, resolves the relevant library via Context7, fetches the docs,
  and returns a concise, example-driven answer. Depends on MCP support landing
  first.

- [ ] **`tool-builder init`** — interactive config scaffolding wizard
- [ ] **URL prompt caching** — `cache: runtime` opt-in for tools that want
  fresh URL content on every run instead of the build-time-embedded snapshot
- [ ] **Homebrew tap** — `brew install rayjohnson/tap/tool-builder` distribution

---

## Done

- [x] Published to GitHub; tagged v0.1.1 → v0.1.4
- [x] `pkg/runtime` — public entry point for generated binaries
- [x] `cmd/build.go` — `tool-builder build` subcommand
- [x] `internal/codegen` — `main.go` generator with `//go:embed` and sorted keys
- [x] Generated `go.mod` with version pinning and dev `replace` directive
- [x] System prompt assembly (`internal/systemprompt`)
- [x] Streaming agent loop (`internal/runner`)
- [x] File tools — `read_file` / `write_file` scoped to `file_access`
- [x] Shell tool-use execution with per-command arg allowlist
- [x] `output_mode: confirm` — diff + yes/no before writing
- [x] TUI tools — `list_select`, `confirm`, `text_input`, `text_editor`
- [x] Sample apps: commit-msg, test-builder, lint-fixer (with Makefiles)
- [x] Removed `tool-builder run` command — build-only model
- [x] docs/ — overview, config reference, TUI tools, examples
