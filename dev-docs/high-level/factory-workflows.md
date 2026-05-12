# Factory workflows

This repository uses agentic GitHub Actions workflows — "factories" — to automate research, proposal authoring, and implementation for issues. Each factory is triggered by a dedicated label.

## Labels

| Label | Color | Description |
|-------|-------|-------------|
| `change-factory` | *(existing)* | Trigger the change-factory agent to author an OpenSpec proposal PR for this issue |
| `code-factory` | *(existing)* | Trigger the code-factory agent to implement this issue in a linked pull request |
| `research-factory` | `#e6b8a2` | Trigger the research-factory agent to author an implementation-research block for this issue |

> **Note for maintainers:** The `research-factory` label must be created in the repository settings before the workflow can be triggered by label events. Use the color `#e6b8a2` (warm orange) and the description shown above.

---

## `research-factory`

### What it does

`research-factory` adds a deep-research pass **before** `change-factory`. It enriches issues with an implementation-research block that compares at least two candidate approaches, surfaces open questions, and grounds decisions in Elastic documentation and a full repository checkout.

### How to trigger

1. **Label trigger:** Apply the `research-factory` label to an existing issue.
2. **Dispatch trigger:** Run the `research-factory-issue` workflow via `workflow_dispatch` and provide the `issue_number` input. This path is intended for automated chaining (e.g. from a future issue classifier).

### What the block looks like

The agent appends or rewrites a single gated section in the issue body, delimited by:

```markdown
<!-- implementation-research:start -->
## Implementation research

_(provenance header with run timestamp, run link, and social-contract notice)_

### Problem framing
...

### Approaches considered
#### Approach A
...
#### Approach B
...

### Recommendation
...

### Open questions
...

### Out of scope
...

### References
...
<!-- implementation-research:end -->
```

Mandatory subsections in order:

- `## Implementation research` — H2 heading with provenance and social-contract notice
- `### Problem framing`
- `### Approaches considered` — two or more `#### ` H4 children
- `### Recommendation`
- `### Open questions`
- `### Out of scope`
- `### References`

### Social contract

- The block is **regenerated on every run**. Edits you make inside the block are read as input on the next re-run but are **not preserved verbatim**.
- For durable feedback, post a comment or edit content **outside** the block.
- To trigger a fresh research pass, re-apply the `research-factory` label.

### How `change-factory` uses it

When an implementation-research block is present in the issue body, `change-factory` treats it as the **exclusive** authoritative scope:

- `### Recommendation` becomes the proposal spine.
- `### Open questions` is copied into `design.md` as `## Open questions`.
- `### Approaches considered` is treated as already-evaluated context.

When the block is absent, `change-factory` falls back to its default behavior: the issue title and body are authoritative.

`change-factory` will **not** modify or rewrite the implementation-research block itself.

### What the workflow does NOT do

- It does **not** open pull requests.
- It does **not** write code or modify repository files.
- It does **not** apply the `change-factory` label — promotion from research to proposal is a **human action**.
