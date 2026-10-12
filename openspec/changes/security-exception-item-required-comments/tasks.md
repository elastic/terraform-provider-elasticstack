## 1. Schema

- [ ] 1.1 In `internal/kibana/security_exception_item/schema.go`, change `comments` to `Computed`-only (remove `Optional`/`Required` and any write-path plan modifiers), and add nested computed attributes `created_at` and `created_by` alongside the existing `id` and `comment`.
- [ ] 1.2 Add a new optional `required_comments` list-of-strings attribute with validators: reject empty-or-whitespace-only elements, reject duplicate elements (exact, case-sensitive).
- [ ] 1.3 Do not add `RequiresReplace` (stock or custom) to `comments` or `required_comments`.
- [ ] 1.4 Add/adjust a plan modifier on `comments` so it is unknown exactly when apply could append (see REQ-025c), and otherwise carries the prior state value forward unchanged.

## 2. Create

- [ ] 2.1 In `internal/kibana/security_exception_item/models.go`, build the create request's `comments` from `required_comments` (in configuration order, `{comment}`, no `id`) when non-null and non-empty; omit the field otherwise.
- [ ] 2.2 After create, read the item back and populate computed `comments` (`id`, `comment`, `created_at`, `created_by`) for every stored comment.

## 3. Update

- [ ] 3.1 Before building the update request, perform a fresh read of the item's current comments (not the planned value).
- [ ] 3.2 Compute `missing` = texts in `required_comments` not present (exact, case-sensitive) among the freshly-read stored comment texts, preserving configuration order.
- [ ] 3.3 When `missing` is empty, omit `comments` from the update request body entirely, even if other attributes changed.
- [ ] 3.4 When `missing` is non-empty, set the update request's `comments` to exactly `missing`, each as `{comment}` with no `id`; do not echo existing comments.
- [ ] 3.5 After update, read the item back and populate computed `comments` for every stored comment.

## 4. Read / Import

- [ ] 4.1 Ensure read always populates `comments` from the full API comment list (`id`, `comment`, `created_at`, `created_by`), independent of `required_comments`.
- [ ] 4.2 Confirm import does not set `required_comments` (leaves it null) and does not add a state upgrader that copies prior `comments` config/state into it.

## 5. Unit tests

- [ ] 5.1 Add unit tests for the missing-text computation: all present (empty result), one missing, several missing (order follows configuration), duplicate stored texts count once, removal from `required_comments` never produces a payload, null/empty `required_comments` never produces a payload.
- [ ] 5.2 Add unit tests for the `required_comments` validators (empty/whitespace element rejected, duplicate element rejected).
- [ ] 5.3 Add a unit test for the `comments` plan-modifier unknown/known behavior described in REQ-025c.

## 6. Acceptance tests

- [ ] 6.1 Create with two `required_comments`; assert computed `comments` has both texts, ids, `created_at`, `created_by`.
- [ ] 6.2 Second apply with no configuration change is a no-op; comment count and ids unchanged (regression test for the duplication bug).
- [ ] 6.3 Add a third required text; assert in-place update, original two ids survive, new text is last.
- [ ] 6.4 Remove a required text from configuration; assert the attribute updates, the Kibana comment and its id remain, and the resource is not replaced.
- [ ] 6.5 Change one required text's value; assert the old comment remains, the new text is appended, and the resource is not replaced.
- [ ] 6.6 Insert a comment directly via the Kibana API (out-of-band); assert it appears in `comments` after refresh and, when required texts are all still present, the next plan is a no-op.
- [ ] 6.7 Configure duplicate strings in `required_comments`; assert validation fails at plan time.
- [ ] 6.8 Import an item with existing comments; assert `comments` is populated and `required_comments` stays unset.

## 7. Documentation and changelog (implementer follow-up, out of scope for this proposal)

- [ ] 7.1 Update the resource's generated documentation to describe `required_comments`, the append-only/computed nature of `comments`, and the migration snippet from `proposal.md`.
- [ ] 7.2 Add a changelog entry under "Breaking changes" describing the `comments` → `required_comments` migration.

## 8. Supersession

- [ ] 8.1 Confirm PR #3572 (id-passthrough + conditional `RequiresReplace`) is not merged; this change supersedes it.
