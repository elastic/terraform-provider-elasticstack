## Context

The verify-label workflow currently documents a hybrid bootstrap model: explicit agent runtimes in workflow frontmatter, an `actions/setup-go` step for the runner environment, and Node setup described in terms of the `package.json` engine range. The requested change simplifies that contract by treating the workflow source as the single runtime declaration point for the review environment while still preserving repository-level drift detection against `go.mod` and `package.json`.

This change spans the OpenSpec requirement, the workflow source and compiled output, and the make-based guardrails that keep workflow runtime declarations aligned with repository metadata. Because the change affects both documented behavior and validation expectations, a short design is useful before implementation.

## Goals / Non-Goals

**Goals:**
- Define the review workflow requirement around explicit pinned workflow runtimes only.
- Remove `actions/setup-go` from the required bootstrap path.
- Keep Go runtime drift detection against `go.mod`.
- Add Node runtime validation so workflow `runtimes.node.version: "24"` is checked against `package.json` `engines.node`.
- Preserve Terraform CLI availability requirements for the review environment.

**Non-Goals:**
- Changing the repository's declared Go version in `go.mod`.
- Changing the repository's declared Node engine range in `package.json`.
- Redesigning unrelated workflow verification, review, archive, or label-cleanup behavior.

## Decisions

### Use workflow frontmatter as the only runtime declaration for review setup

The modified requirement will treat `runtimes.go.version` and `runtimes.node.version` in the workflow frontmatter as the authoritative toolchain declarations for the review environment.

Alternative considered: keep the previous split model where Go is pinned twice, once in frontmatter and once through `actions/setup-go`. Rejected because it duplicates intent, increases drift risk, and conflicts with the goal of describing only the runtime declarations that actually need to remain.

### Pin the required versions to `go 1.26.1` and `node 24`

The requirement will state the concrete values the workflow must declare: `runtimes.go.version == "1.26.1"` and `runtimes.node.version == "24"`. The Go value stays aligned with `go.mod`, and the Node value is intentionally narrower than the engine range while still needing to satisfy it.

Alternative considered: continue specifying Node behavior only as "satisfies the engine range." Rejected because the requested behavior is an explicit workflow runtime pin, not a dynamic resolution from the range.

### Extend make-based runtime drift checks to cover Node compatibility

The repository already has make targets that compare the workflow's Go runtime declaration to `go.mod`. Implementation should extend this validation family so the workflow's pinned Node runtime is also checked against `package.json` `engines.node`, and the same check path used by CI lint verification covers both languages.

Alternative considered: rely on workflow execution failures to catch an unsupported Node pin. Rejected because the repo already uses preflight drift checks for Go, and Node should follow the same deterministic maintenance path.

## Risks / Trade-offs

- Fixed workflow Node pin can drift from `package.json` engine updates -> Mitigation: add make-based validation and keep it in the existing lint/check path.
- Removing `actions/setup-go` could surprise maintainers who still think runner setup is separate from agent runtime setup -> Mitigation: update the requirement text and workflow comments together so the single-source-of-truth model is explicit.
- Pinned runtime values require manual updates when repository toolchains change -> Mitigation: preserve the sync/check flow for Go and add equivalent validation expectations for Node.
