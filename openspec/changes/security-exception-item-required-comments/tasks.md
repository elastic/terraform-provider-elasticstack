## 1. Schema

- [ ] 1.1 In `internal/kibana/security_exception_item/schema.go`, change `comments` to `Computed`-only (remove `Optional`/`Required` and any write-path plan modifiers), and add nested computed attributes `created_at` and `created_by` alongside the existing `id` and `comment`. Format `created_at` as `2006-01-02T15:04:05.000Z`.
- [ ] 1.2 Add a new optional `required_comments` list-of-strings attribute with validators: reject empty-or-whitespace-only elements, reject duplicate elements (exact, case-sensitive). Schema descriptions cover the computed log and the presence set, and do not include a migration snippet.
- [ ] 1.3 Do not add `RequiresReplace` (stock or custom) to `comments` or `required_comments`. A comment-only plan stays in-place. `space_id`, `item_id`, `list_id`, `type`, and `namespace_type` still replace under REQ-010, and the new item's initial comments are `required_comments` only.
- [ ] 1.4 Add/adjust a plan modifier on `comments` so it is unknown on create, when `required_comments` is unknown, and when a known required text is absent from state (REQ-025c). Otherwise carry the prior state value forward unchanged.

## 2. Create

- [ ] 2.1 In `internal/kibana/security_exception_item/models.go`, build the create request's `comments` from `required_comments` (in configuration order, `{comment}`, no `id`) when non-null and non-empty; omit the field otherwise.
- [ ] 2.2 After create, read the item back and populate computed `comments` (`id`, `comment`, `created_at`, `created_by`) for every stored comment.

## 3. Update

- [ ] 3.1 Before building the update request, perform a fresh read of the item's current comments (not the planned value).
- [ ] 3.2 Compute `missing` = texts in `required_comments` not present (exact, case-sensitive) among the freshly-read stored comment texts, preserving configuration order.
- [ ] 3.3 When `missing` is empty, omit `comments` from the update request body entirely, even if other attributes changed.
- [ ] 3.4 When `missing` is non-empty, append those texts in configuration order as `{comment}` with no `id`. On Kibana versions that predate elastic/kibana#145575, echo every stored comment as `{id, comment}` in stored order and then append the missing texts. Never resend a stored text without its `id`, and never place a new comment before an id-bearing entry.
- [ ] 3.5 After update, read the item back and populate computed `comments` for every stored comment. Store an empty API log as null.

## 4. Read / Import

- [ ] 4.1 Ensure read always populates `comments` from the full API comment list (`id`, `comment`, `created_at`, `created_by`), independent of `required_comments`. Format `created_at` as `2006-01-02T15:04:05.000Z`. Store an empty API log as null.
- [ ] 4.2 Confirm import does not set `required_comments` (leaves it null) and does not add a state upgrader that copies prior `comments` config/state into it.

## 5. Unit tests

- [ ] 5.1 Add unit tests for the missing-text computation: all present (empty result), one missing, several missing (order follows configuration), duplicate stored texts count once, an analyst-written exact text counts as present, removal from `required_comments` never produces a payload, null/empty `required_comments` never produces a payload. Cover both allowed append bodies: missing texts only, and stored comments echoed with ids followed by the missing texts. Assert a stored text is never sent without its id.
- [ ] 5.2 Add unit tests for the `required_comments` validators (empty/whitespace element rejected, duplicate element rejected).
- [ ] 5.3 Add a unit test for the `comments` plan-modifier unknown/known behavior described in REQ-025c, including unknown `required_comments` making `comments` unknown.

## 6. Acceptance tests

- [ ] 6.1 Create with two `required_comments`; assert computed `comments` has both texts, ids, `created_at`, `created_by`.
- [ ] 6.2 Second apply with no configuration change is a no-op; comment count and ids unchanged (regression test for the duplication bug).
- [ ] 6.3 Add a third required text; assert in-place update, original two ids survive, new text is last.
- [ ] 6.4 Remove a required text from configuration; assert the attribute updates, the Kibana comment and its id remain, and the resource is not replaced.
- [ ] 6.5 Change one required text's value; assert the old comment remains, the new text is appended, and the resource is not replaced.
- [ ] 6.6 Insert a comment directly via the Kibana API (out-of-band). After refresh, assert it appears in `comments`. With configuration unchanged and every required text still present, assert the following plan is a no-op.
- [ ] 6.7 Configure duplicate strings in `required_comments`; assert validation fails at plan time.
- [ ] 6.8 Import an item with existing comments; assert `comments` is populated and `required_comments` stays unset.
- [ ] 6.9 Rewrite `TestAccResourceExceptionItem_Comments` and `testdata/TestAccResourceExceptionItem_Comments/comments_create` so configuration uses `required_comments` instead of a `comments` block.
- [ ] 6.10 When the stored log already contains a required text written outside Terraform, assert apply does not append a second copy.

## 7. Documentation and changelog (implementer follow-up, out of scope for this proposal)

- [ ] 7.1 Regenerate `docs/resources/kibana_security_exception_item.md` from schema descriptions. The page describes `required_comments` and the computed `comments` log. It does not include a migration snippet.
- [ ] 7.2 Add a changelog entry under "Breaking changes" describing the move from a configured `comments` block to `required_comments`. That entry is the user-facing migration note.
- [ ] 7.3 When syncing this change, replace the `comments` stanza in `openspec/specs/kibana-security-exception-item/spec.md` with the schema refinement in the delta, including `required_comments`.

## 8. Supersession

- [ ] 8.1 Confirm PR #3572 (id-passthrough + conditional `RequiresReplace`) is not merged; this change supersedes it.
