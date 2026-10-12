## Why

`elasticstack_kibana_security_exception_item.comments` is modeled today as an ordinary mutable list of `{ id (computed), comment }`. The Kibana exception-item API (`PUT /api/exception_lists/items`) does not support that model: it is append-only. `toUpdateRequest` in `internal/kibana/security_exception_item/models.go` sends every configured comment's text on every update and never sends `id`, so Kibana appends a fresh copy of each existing comment on every `terraform apply` — a real item has been observed growing 1 → 2 → 4 → 8 comments across three consecutive applies with no user-side change. Attempting to edit, reorder, or remove a comment through configuration is either silently dropped by Kibana (producing a `produced inconsistent result after apply` error) or rejected outright with `400 item "comments" are append only` when a new entry is not placed at the end.

Two conditional-replacement designs were evaluated and rejected (see `design.md`): passing `id` through plus a plan modifier that forces `-/+ destroy and replace` whenever the diff is not a pure trailing append (the issue body's proposal, and the independent implementation-research comment's recommendation). Both treat `comments` as an attribute Terraform fully owns. In practice analysts add comments directly in Kibana; after a refresh, Terraform's state holds the full server log while the user's configuration only ever listed the subset it cares about. Under either rejected design, that mismatch reads as a shrink and forces every subsequent apply to destroy and recreate the exception item just to converge a list Terraform never fully owned, dropping the exception for the duration of the replace, re-stamping every `created_by` to the API-key owner, and changing the item's generated id.

## What Changes

- **BREAKING**: `comments` becomes a computed-only attribute (`id`, `comment`, `created_at`, `created_by`) representing the full comment log Kibana has stored for the item, including comments added outside Terraform. Configurations that currently set `comments` will fail to apply and must migrate (see Migration below).
- Add a new optional `required_comments` attribute: a list of comment-text strings that Terraform ensures are present somewhere in the log. It is a "must contain" set, not the log itself and not an exclusive list. Validated as non-empty, non-whitespace-only strings with no duplicate entries (exact, case-sensitive comparison).
- **Create**: POST `required_comments`, in configuration order, as `{comment}` entries with no `id`, when the list is non-null and non-empty; omit `comments` from the create body otherwise.
- **Update**: Before building the update request, compute `missing` as the `required_comments` texts not already present (by exact text) among the item's currently-stored comment texts, determined from a fresh read taken during apply (not from the planned `comments` value). When `missing` is empty, omit `comments` from the PUT body entirely — even when some other attribute changed. When `missing` is non-empty, send only those texts, in configuration order, each as `{comment}` with no `id`; never echo existing comments or ids back to the API.
- **Plan**: `comments` is planned as unknown whenever apply could append (any configured `required_comments` text not already present in state); otherwise the prior state value is carried through unchanged, so reordering or removing a `required_comments` entry — which never touches the Kibana log — does not perturb `comments` in the plan.
- **No `RequiresReplace`** is added to either `comments` or `required_comments`. Comment drift of any kind (edit, remove, reorder) never destroys and recreates the exception item under this design — only a change to an existing immutable attribute (e.g. `list_id`, `type`) does, per existing REQ-010 behavior.
- Removing a string from `required_comments` only stops Terraform from requiring it; the corresponding Kibana comment, if it had been created, is left in place. Terraform never issues a delete for an individual comment — the public API offers no such operation.

### Migration (breaking change, one release)

```hcl
# Before
comments = [
  { comment = "Exception approved in ticket SEC-123" },
  { comment = "Scoped to linux hosts only" },
]

# After
required_comments = [
  "Exception approved in ticket SEC-123",
  "Scoped to linux hosts only",
]
```

Comments already stored in Kibana are unaffected either way; they surface under the now-computed `comments` attribute after the next read. `required_comments` does not need to list them to avoid deletion — nothing in this resource ever deletes a comment.

## Capabilities

### Modified Capabilities

- `kibana-security-exception-item`: `comments` becomes computed-only (full server log, including out-of-band comments); a new `required_comments` optional attribute drives an append-only reconciliation on create/update; no replacement is ever forced by comment changes.

## Impact

- **Affected code**: `internal/kibana/security_exception_item/schema.go` (new `required_comments` attribute; `comments` becomes `Computed`-only with new `created_at`/`created_by` nested attributes), `internal/kibana/security_exception_item/models.go` (comment-to-API mapping, missing-text computation, plan modifiers for `comments`), `internal/kibana/security_exception_item/resource.go` or equivalent CRUD wiring (fresh read during update before building the PUT body).
- **Acceptance tests**: `internal/kibana/security_exception_item/acc_test.go` needs new cases for no-op second apply, incremental append, removal-without-delete, edit-without-delete, out-of-band comment tolerance, and import behavior (see `tasks.md`).
- **Documentation**: the resource's generated docs (`docs/resources/kibana_security_exception_item.md.tmpl` or equivalent) need a breaking-change note and usage guidance for `required_comments`; out of scope for this proposal-only change but flagged in `tasks.md` for the implementer.
- **Changelog**: a breaking-change entry is required; flagged in `tasks.md` for the implementer (`make` changelog tooling, not authored here).
- Out of scope: `elasticstack_kibana_security_exception_list` (no `comments` attribute today), editing or deleting a single existing comment (unsupported by the Kibana API), preserving `created_by`/`created_at` across a replacement triggered by an unrelated immutable attribute, enforcing comment order after create, and PR #3572 (the id-passthrough + conditional-`RequiresReplace` fix), which this proposal supersedes and which must not be merged.
