## Context

Kibana has treated exception-item comments as append-only since `elastic/kibana#73129` (July 2020). `created_by`/`created_at` are stamped server-side from the authenticated API-key owner and cannot be supplied by the client. The update request body accepts only `comment` and an optional `id`; there is no operation to delete a single comment.

The following was measured directly against a local Kibana **9.4.0** stack (`PUT /api/exception_lists/items`), starting each case from an item whose comments were `alpha`, `beta` unless noted:

| PUT `comments` | HTTP | Result |
|---|---|---|
| Field omitted, `name` changed | 200 | Comments and ids unchanged. Name changed. |
| `[]` | 200 | Comments and ids unchanged. |
| Only `[{comment: "gamma"}]` | 200 | `gamma` appended. `alpha` and `beta` ids kept. |
| Only `[{comment: "gamma"}, {comment: "delta"}]` | 200 | Both appended. Existing ids kept. |
| Existing texts with no ids, plus `gamma` | 200 | `alpha` and `beta` duplicated, then `gamma`. |
| Existing comments with their ids, plus new texts at the end | 200 | New texts appended. Existing ids kept. |
| One existing id omitted, plus `gamma` at the end | 200 | Omitted comment kept. `gamma` appended. |
| A new comment placed before existing id-bearing comments | 400 | No change. `item "comments" are append only`. |
| Item that had no comments, then `[{comment: "gamma"}]` | 200 | `gamma` stored. |
| `[{comment: "alpha"}]` when `alpha` already exists, no id | 200 | A second `alpha` is appended. Kibana does not dedupe. |

An append does not need the existing comments echoed back on Kibana 9.4.0: sending only the new texts appends them. Echoing existing ids, then the new texts, is also a correct append on that version, and it is the body older servers require (elastic/kibana#145575). Sending an existing text again with no id is exactly what duplicates the log. Omitting `comments`, or sending `[]`, leaves the log untouched.

The sibling resource, `elasticstack_kibana_security_exception_list`, has no `comments` attribute; this change is scoped to `security_exception_item` only.

## Goals / Non-Goals

**Goals:**
- Stop the duplication-on-every-apply bug (1 → 2 → 4 → 8 comments) without ever recreating the exception item for a comment-only change.
- Make a second `apply` with no configuration change a true no-op for `comments`.
- Tolerate comments added directly in Kibana (out-of-band) without forcing replacement or reporting drift against `required_comments`.
- Support the one operation the API actually allows in place — appending new text — as a cheap, visible, in-place update.

**Non-Goals:**
- Editing or deleting an individual existing comment (not supported by the Kibana API regardless of provider modeling).
- Preserving `created_by`/`created_at` across a replacement triggered by an unrelated immutable attribute (e.g. `list_id`, `type`) — irreducible at the public API layer.
- Enforcing comment order after create.
- Sending `_version` on update (this resource does not send it today, and this change does not add it).
- Fixing or extending `elasticstack_kibana_security_exception_list` (no `comments` attribute exists there).

## Decisions

### Decision: model `comments` as a computed log and `required_comments` as a "must be present" set (Option D)

`comments` becomes fully computed and reflects whatever Kibana actually stores, including comments added outside Terraform. `required_comments` is a separate, optional, user-owned list of strings; its only contract is that each listed text exists somewhere in the log after apply. Removing a string from `required_comments` stops requiring it going forward but never deletes the underlying Kibana comment, and never forces replacement.

On update, the provider performs a fresh read during apply and computes the set difference between `required_comments` and the currently-stored comment texts (exact, case-sensitive match). A stored text counts whether or not Terraform wrote it. If the fresh read fails or cannot provide the complete current log, the update aborts with the read diagnostic and sends no PUT; the provider never falls back to planned or empty comments. When nothing is missing, `comments` is omitted from the PUT entirely, so an unrelated attribute change (name, description, entries, tags, …) never touches the comment log. When something is missing, the PUT appends those texts and never resends a stored text without its `id`.

Two bodies do that. Only the missing texts, each `{comment}` with no `id`, is used on Kibana 8.6.0 and later. Kibana 8.5.x and earlier use the echo-with-ids body: every stored comment as `{id, comment}` in stored order, then the missing texts. This boundary follows the fix for elastic/kibana#145575, which was backported to 8.6.0. The server version must be successfully discovered before selecting a body when an append is needed; a version-discovery diagnostic aborts before PUT. A new comment is never placed before an id-bearing entry.

The fresh-read-then-PUT sequence is not atomic. Since this design does not send `_version` or use conditional writes, another apply or Kibana user can append a required text after the fresh read but before the PUT; Kibana may then append a duplicate. This accepted race cannot be prevented by a retry without risking another duplicate. API failures are returned as diagnostics without automatic retry; a subsequent read reports whatever Kibana stored.

`comments` is unknown in the plan on create, when `required_comments` is unknown, and when a known required text is absent from state. Otherwise the prior log is kept, so a post-apply read of a longer list cannot disagree with a plan that still showed the old one. An empty API log is stored as null, the same as this resource's other empty collections. Comment `created_at` uses `2006-01-02T15:04:05.000Z`.

No `RequiresReplace` plan modifier is added to `comments` or `required_comments`. Comment drift — an edit, a removal, a reorder, or an out-of-band addition — never destroys and recreates the exception item. A change to `space_id`, `item_id`, `list_id`, `type`, or `namespace_type` still replaces it under the existing rule, and the new item starts from `required_comments` only.

### Decision: upgrade existing comment state without inferring requirements

The resource schema version advances from 0 to 1. Its v0-to-v1 state upgrader preserves each existing comment's `id` and `comment` and adds null `created_at` and `created_by` values so Terraform can decode the expanded nested object before refresh. It leaves the new `required_comments` attribute null. The upgrader does not copy prior comment state into `required_comments`: v0 state represents the full observed Kibana log, not the subset the user intended to require.

### Rejected: Option A — id passthrough + `UseStateForUnknown` + conditional `RequiresReplace` on the full list

This is the design proposed in the issue body and independently recommended by the implementation-research comment (spine `full-fidelity-append-only-plan-modifier`, confidence medium). It treats `comments` as a list Terraform fully owns: pass `id` through on update, add `stringplanmodifier.UseStateForUnknown()` to `comment.id`, and add a custom `planmodifier.List` on `comments` that permits only a pure trailing append and forces `-/+ destroy and replace` for anything else (edit, removal, reorder), with a warning about `created_by` collapsing to the API-key owner.

This matches the API's actual capability when Terraform is the only writer and `comments` in configuration is the complete, authoritative log. It does not match how comments are used in practice: analysts add comments through the Kibana UI. After the next refresh, state holds the full server log (Terraform's list plus every out-of-band comment) while configuration still only lists the subset Terraform's author cares about. Under Option A that mismatch is indistinguishable from an intentional shrink, so every such refresh forces a destroy-and-recreate: the exception is unenforced for the duration of the replace, every comment's `created_by` is re-stamped to the API-key owner (losing the original analyst attribution for comments that were never even written by Terraform), and the item's generated id changes. Copying each UI-added comment back into HCL to keep the plan quiet only holds until the next out-of-band comment.

### Rejected: Option B — `RequiresReplace` on any change to `comments`

Considered as a simpler variant of A in the implementation-research comment: any change to `comments`, including a pure append, forces replacement. This has the same operational cost as A — destroy-and-recreate, `created_by` collapse, id churn — even for the one operation (append) that the API supports cleanly in place. Rejected for the same reason as A, with no offsetting simplicity benefit significant enough to justify destroying the resource for the common case.

### Rejected: Option C — silent plan-time reconciliation without replacement

A plan modifier that rewrites the planned `comments` value to whatever Kibana will actually store — keeping state entries verbatim and accepting only new trailing no-id entries — while emitting a warning when an edit or removal is discarded, but never forcing replacement. This avoids destroying the resource, but it silently no-ops an explicit user edit or deletion of a `comments` entry forever: `terraform apply` reports success while doing something other than the diff the user asked for. Option D is preferred because it makes the "must be present" contract on `required_comments` explicit and separate from the observed log, rather than quietly discarding part of a single conflated attribute.

### Deviation from the implementation-research comment

The implementation-research comment's `### Recommendation` (Approach A) and the issue body's own three-part fix are not implemented. Both were explicitly superseded by human direction on this issue (`@tobio`, 2026-10-12), which identifies the same root failure mode described above — out-of-band Kibana comments making `comments` not fully Terraform-owned — and selects Option D instead. Per that direction, PR #3572 (which implements Option A) is superseded and must not be merged.

## Open questions

Copied verbatim from the implementation-research comment; the first (blocking) question is resolved below, the others are not applicable to the selected design.

- Can the "pure append" boundary be verified live (multi-append in one Update; `comments = null` → single new comment) before merge? **(blocking)**
- Should the `RequiresReplace` warning also fire when `comments` goes empty-in-state → non-empty-in-plan on an existing/imported resource?
- Does PR #3551 exist under another number or fork? Unresolvable via `gh` in this checkout.

**Resolution**: the blocking question is closed by the measurement table in `## Context` above (multi-append in one `PUT`, and an initially-empty log receiving its first comment, were both verified against a live Kibana 9.4.0 cluster as part of human direction on this issue). The second question does not apply: this design never adds a `RequiresReplace` modifier to `comments`, so there is no empty→non-empty replacement case to reconsider. The third question (PR #3551's existence) remains unresolved and is immaterial to this design, which supersedes both PR #3551 and PR #3572.

## Migration Plan

This is a one-release breaking configuration change: configurations that set `comments` fail to apply starting with the release that ships this change. The v0-to-v1 state upgrader only makes the expanded computed comment object decodable by adding null audit fields; it leaves `required_comments` null and does not migrate the full historical server log into user intent. Users who want a text enforced move it from `comments` onto `required_comments` and delete the `comments` block. Stored comments stay in Kibana either way. The changelog entry is the user-facing migration note. Resource docs do not repeat it.

Syncing this change into `openspec/specs/kibana-security-exception-item/spec.md` also replaces the `comments` stanza in that file's schema block. Requirement deltas do not rewrite that fence on their own, and leaving it in place would describe the old optional list after REQ-025 has changed.

## Rollout considerations

Flagged for the implementer, out of scope for this proposal:
- Changelog entry under "Breaking changes", including how to move enforced texts onto `required_comments`.
- Schema descriptions for `required_comments` and the computed `comments` log. tfplugindocs writes those into `docs/resources/kibana_security_exception_item.md`. No migration snippet in the description or the example.
