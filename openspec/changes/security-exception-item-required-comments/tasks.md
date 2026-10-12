## 1. Create and read the computed comment log

- [ ] 1.1 Make `comments` computed-only with `id`, `comment`, `created_at`, and `created_by`, and add optional validated `required_comments`; verify schema tests cover timestamp format, whitespace, and duplicates.
- [ ] 1.2 Create items from `required_comments` in configuration order and populate the full computed log on authoritative read; verify create/read tests cover empty and non-empty logs.
- [ ] 1.3 Read the full comment log independently of `required_comments`, representing an empty log as null; verify read tests include out-of-band entries and the exact `2006-01-02T15:04:05.000Z` layout.
- [ ] 1.4 Replace the canonical spec's old writable `comments` schema sketch with the computed log and `required_comments`; verify the sketch matches the delta schema.

## 2. Reconcile missing comments in-place on update

- [ ] 2.1 Plan `comments` as unknown only when create, unknown required texts, or missing required texts may change the log; otherwise preserve prior state, with no replacement for comment-only changes. Verify plan-modifier tests include required-text reordering.
- [ ] 2.2 Reconcile updates against a fresh complete comment read, omitting `comments` when nothing is missing and appending only missing texts in configuration order; verify unit tests cover matching, missing, removal, analyst-written text, and both payload shapes.
- [ ] 2.3 Select the update payload by server version: echo stored ids for 8.5 and earlier, missing-only for 8.6 and later; abort before PUT on read or version-discovery failure. Verify tests prove 8.5, 8.6+, and both diagnostic paths.
- [ ] 2.4 Preserve stored comment order and append new texts without resending stored texts without ids; verify acceptance tests cover no-op apply, incremental append, unchanged stored ids, edits, removals, and reordering without replacement.
- [ ] 2.5 Document the non-atomic concurrent-append limitation and no-automatic-retry behavior; verify the design and delta describe the possible duplicate outcome.

## 3. Preserve imported and upgraded comment history

- [ ] 3.1 Upgrade v0 comment state to schema version 1 by retaining `id` and `comment`, adding null audit fields, and leaving `required_comments` null; verify state-upgrade tests cover existing nested comments and empty logs.
- [ ] 3.2 Keep import from inferring `required_comments` from the API log; verify acceptance tests show imported comments remain observable and only missing configured texts are appended.
- [ ] 3.3 Verify acceptance tests cover out-of-band comments, validation errors, and immutable-field replacement while preserving the required-comments lifecycle.
- [ ] 3.4 Rewrite `TestAccResourceExceptionItem_Comments` and its create fixture to configure `required_comments`; regenerate resource documentation and add the breaking-change changelog entry.
