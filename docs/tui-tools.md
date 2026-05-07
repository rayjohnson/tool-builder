# TUI tools

TUI tools are interactive terminal UI components the agent can call mid-task to collect
input from the user. They pause the agentic loop until the user responds, then return
the result to the agent as a tool result.

## Why TUI tools exist

The agent's text output ends the conversation turn immediately. An agent cannot ask a
question in plain text and wait for a reply — the moment it finishes generating, the turn
is over. TUI tools solve this: they are proper tool calls that pause execution, render
a UI in the terminal, and return structured data back to the agent.

**Rule for system prompt authors:** any time the agent needs a decision from the user
before taking an action, it must call a TUI tool — not ask a question in text. Use
`confirm` for yes/no decisions. Use `list_select` when the user must pick from a list.

## Enabling TUI tools

Add tool names to `tool_use.tui` in the config:

```yaml
tool_use:
  enabled: true
  tui:
    - list_select
    - confirm
    - text_input
    - text_editor
```

Only listed tools are available to the agent. Omit tools the agent does not need.

---

## list_select

Shows an interactive selection list and returns the chosen items.

**Use when:** you have a set of options and want the user to decide which ones to act on —
files to include, items to process, people to notify, etc.

### Input

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | yes | Heading shown above the list |
| `items` | array of strings | yes | The options to display |
| `single_select` | bool | no | If true, user picks exactly one item; default false (multi-select) |

### Return value

A JSON-encoded array of the selected strings. Empty array `[]` if the user selected
nothing or cancelled.

### Keyboard controls

| Key | Action |
|---|---|
| ↑ / k | Move up |
| ↓ / j | Move down |
| Space | Toggle selection |
| a | Select all |
| n | Select none |
| Enter | Confirm selection |
| q | Cancel (returns empty array) |

### Example system prompt usage

```markdown
If there are unstaged files, use the `list_select` tool to let the user pick which ones
to include. Set `title` to "Select files to include in this commit". Set `items` to the
list of unstaged file paths. Leave `single_select` unset (multi-select is the default).
```

---

## confirm

Asks the user a yes/no question and returns their answer. A single keypress is sufficient;
no Enter needed.

**Use when:** the agent is about to take an irreversible or significant action — staging
files, running `git commit`, deleting data, sending a message, calling an external API.
Also use it whenever you would otherwise ask a yes/no question in text.

### Input

| Field | Type | Required | Description |
|---|---|---|---|
| `question` | string | yes | The yes/no question to display |
| `default_yes` | bool | no | If true, Enter defaults to yes; default false (Enter = no) |

### Return value

The string `"yes"` or `"no"`.

### Example system prompt usage

```markdown
After writing the commit message, use the `confirm` tool to ask "Run git commit?" —
if the answer is "yes", run `git commit -m "<message>"`. If "no", stop.

**Never** ask yes/no questions in plain text — use `confirm` every time you need the
user to make a decision before you act.
```

---

## text_input

Prompts the user to type a short text value using a full-featured text field with cursor
movement and editing support.

**Use when:** you need free-form input — a name, email address, search query, a custom
message, or any other short string the agent cannot infer from context.

### Input

| Field | Type | Required | Description |
|---|---|---|---|
| `prompt` | string | yes | The label or question shown above the input field |
| `placeholder` | string | no | Placeholder text shown in the empty field |

### Return value

The typed text as a string. Empty string if the user cancelled.

### Example system prompt usage

```markdown
If the user did not provide a `--hint` flag, use the `text_input` tool to ask for one:
- Set `prompt` to "What is this commit about? (optional hint)"
- Set `placeholder` to "e.g. fix the nil panic in config.Load"
Use the returned value as additional context when writing the commit message.
```

---

## text_editor

Opens the user's preferred text editor (`$EDITOR`) with an initial draft and returns the
edited content after the user saves and closes.

**Use when:** the content is long enough that a text field is awkward — a commit message
body, an email draft, a generated config snippet, a document section. The agent provides
a draft; the user refines it.

### Input

| Field | Type | Required | Description |
|---|---|---|---|
| `content` | string | yes | The initial draft to open in the editor |
| `filename` | string | no | Filename hint for syntax highlighting (e.g. `commit-msg.txt`, `query.sql`) |

### Return value

The saved file contents as a string after the editor closes.

Editor resolution order: `$EDITOR`, `$VISUAL`, `vi`, `nano`. If none of these is available,
returns an error.

### Example system prompt usage

```markdown
After generating the commit message, use the `text_editor` tool to let the user review
and edit it before committing:
- Set `content` to the generated message
- Set `filename` to `commit-msg.txt`
Use the returned text as the final commit message.
```

---

## Calling pattern

TUI tools serialize with other terminal output via a mutex — only one TUI interaction runs
at a time. You do not need to coordinate this in your system prompt; it is handled
automatically.

The agent may call TUI tools multiple times in a session. For example, `list_select` might
be called once at the start to collect files, and `confirm` called later before committing.
Each call is independent.
