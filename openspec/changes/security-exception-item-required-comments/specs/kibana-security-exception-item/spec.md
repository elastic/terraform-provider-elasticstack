# Delta spec: `kibana-security-exception-item` — `required_comments` and append-only `comments` log

Extends the main spec at `openspec/specs/kibana-security-exception-item/spec.md`.

## Schema refinements

When this change is synced, replace the `comments` stanza in the canonical schema block with a computed log, and add `required_comments`. An empty log is null, matching the other empty API collections in this spec.

```hcl
  required_comments = <optional, list(string)>        # texts that must be present; exact match; duplicates rejected

  comments {                                          # computed; full Kibana log, including out-of-band comments
    id         = <computed, string>
    comment    = <computed, string>
    created_at = <computed, string>                   # 2006-01-02T15:04:05.000Z
    created_by = <computed, string>
  }
```

## MODIFIED Requirements

### Requirement: Mapping — comments on create and update (REQ-025)

The `comments` attribute SHALL be `Computed`-only and SHALL NOT be settable in configuration; it represents the full comment log Kibana has stored for the item, including comments created outside Terraform. The resource SHALL instead accept an optional `required_comments` attribute (list of strings) that names comment texts that MUST be present somewhere in the log.

On create, when `required_comments` is known and non-empty, the resource SHALL include each text in the create request body, in configuration order, using the create-comment API shape (`{comment}`, no `id`). When `required_comments` is null or empty, the resource SHALL omit `comments` from the create request body.

On update, the resource SHALL first perform a fresh read of the item's current comments, then compute `missing` as the `required_comments` texts (exact, case-sensitive string match) not present among the currently-stored comment texts. A stored text satisfies a required text whether or not Terraform created that comment. When `missing` is empty, the resource SHALL omit `comments` from the update request body, even when other attributes changed. When `missing` is non-empty, the resource SHALL send a `comments` array that appends exactly those missing texts, in configuration order, each as `{comment}` with no `id`. The resource SHALL NOT include a stored comment text in that array unless the entry also carries that comment's stored `id`. Any id-bearing entries SHALL keep stored order and SHALL precede the new no-id entries.

Two bodies meet that rule. The first is only the missing texts. The second is every stored comment echoed as `{id, comment}` in stored order, followed by the missing texts. Sending only the missing texts was verified on Kibana 9.4.0. Kibana versions that predate elastic/kibana#145575 (merged 2022-11-18) reject an update whose `comments` array contains only the new comment; on those versions the resource SHALL use the echo-with-ids body. The resource SHALL NOT resend a stored text with no `id`.

Computed comment `id`, `comment`, `created_at`, and `created_by` values SHALL be populated from the API response on every read, for every comment currently stored (not only ones Terraform created). `created_at` SHALL be formatted as `2006-01-02T15:04:05.000Z`, the same layout as REQ-030. When the API returns no comments, `comments` SHALL be null in state, not an empty list.

#### Scenario: Comment id populated from API

- GIVEN `required_comments` configured with texts not yet present
- WHEN the resource is read after create or update
- THEN each comment in state's `comments` SHALL include the Kibana-assigned `id`, the stored `comment` text, `created_at`, and `created_by`

#### Scenario: No-op update omits comments

- GIVEN every text in `required_comments` is already present among the item's stored comments
- WHEN an update is triggered by an unrelated attribute change (e.g. `name`)
- THEN the update request body SHALL NOT include a `comments` field

#### Scenario: Update appends a missing text without resending stored texts bare

- GIVEN stored comments are `alpha` and `beta`, and `required_comments` is `["alpha", "gamma"]`
- WHEN update runs
- THEN the update request SHALL append `gamma` as `{comment}` with no `id`
- AND the request SHALL NOT include `alpha` or `beta` unless that entry also carries the stored id for that comment
- AND any id-bearing entries SHALL precede `gamma`

#### Scenario: Empty comment log is null

- GIVEN the API returns no comments for the item
- WHEN state is written
- THEN `comments` SHALL be null in state, not an empty list

#### Scenario: Comment created_at uses the Kibana timestamp layout

- GIVEN the API returns a comment `created_at`
- WHEN state is written
- THEN that value SHALL be the string form `2006-01-02T15:04:05.000Z`

## ADDED Requirements

### Requirement: `required_comments` schema and validation (REQ-025a)

`required_comments` SHALL be an optional list of strings. Each element SHALL be validated at plan time as a non-empty string that is not only whitespace. The list SHALL be validated at plan time to reject duplicate elements under exact, case-sensitive string equality. Elements SHALL NOT be trimmed or otherwise normalized before comparison or transmission.

#### Scenario: Whitespace-only text rejected

- GIVEN `required_comments` contains an element that is empty or only whitespace
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for `required_comments`

#### Scenario: Duplicate text rejected

- GIVEN `required_comments` contains the same string twice
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for `required_comments`

### Requirement: No forced replacement on comment changes (REQ-025b)

Neither `comments` nor `required_comments` SHALL carry a `RequiresReplace` plan modifier, and no custom plan modifier SHALL force replacement based on a diff in either attribute. A plan that changes only `required_comments` or the computed `comments` log — addition, removal, reordering, or a text edit of `required_comments`, or an out-of-band change to the log — SHALL be an in-place update. A plan that also changes `space_id`, `item_id`, `list_id`, `type`, or `namespace_type` SHALL still require replacement under REQ-010. That replacement creates a new item whose initial comments are `required_comments` under the create rules in REQ-025. Comments that existed only on the destroyed item are not copied.

#### Scenario: Removing a required text does not replace the resource

- GIVEN state has `required_comments = ["alpha", "beta"]` and stored comments include both texts
- WHEN configuration changes `required_comments` to `["alpha"]`
- THEN the plan SHALL show an in-place update (not a destroy/create) and SHALL NOT remove `beta` from the stored comment log

#### Scenario: Editing a required text does not replace the resource

- GIVEN state has `required_comments = ["alpha"]` and a stored comment with text `alpha`
- WHEN configuration changes `required_comments` to `["alpha-revised"]`
- THEN the plan SHALL show an in-place update, and apply SHALL append `alpha-revised` as a new comment while leaving the original `alpha` comment stored

#### Scenario: An immutable-field change still replaces the item

- GIVEN an item whose stored comments include an out-of-band comment and whose `required_comments` is `["alpha"]`
- WHEN configuration changes `list_id` and leaves `required_comments` unchanged
- THEN the plan SHALL require replacement under REQ-010
- AND the created item's initial comments SHALL be `alpha` only

### Requirement: Plan-time unknown handling for `comments` (REQ-025c)

`comments` SHALL be planned as unknown on create, when `required_comments` is unknown, and on update when any known `required_comments` text is not already present among the prior state's stored comment texts. When `required_comments` is known and every one of its texts is already present in the prior state's stored comment texts, `comments` SHALL be planned as unchanged from the prior state value. Null and empty `required_comments` are known and name nothing missing, so they do not by themselves make `comments` unknown.

#### Scenario: Plan is unknown when a new required text will be appended

- GIVEN state's stored comments are `alpha` and configuration changes `required_comments` to include `beta` (not previously present)
- WHEN Terraform plans the update
- THEN `comments` SHALL be shown as unknown (`(known after apply)`) in the plan

#### Scenario: Plan keeps comments unchanged when nothing will be appended

- GIVEN state's stored comments are `alpha`, `beta` and `required_comments` is reduced from `["alpha", "beta"]` to `["alpha"]`
- WHEN Terraform plans the update
- THEN `comments` SHALL be shown as unchanged from state, and the plan SHALL show only the `required_comments` edit

#### Scenario: Unknown required_comments makes comments unknown

- GIVEN `required_comments` is unknown at plan time
- WHEN Terraform plans the update
- THEN `comments` SHALL be shown as unknown (`(known after apply)`) in the plan

### Requirement: Out-of-band comments do not force drift or replacement (REQ-025d)

When a read observes comments in the stored log that were not created by Terraform (added directly against the Kibana API), the resource SHALL include them in state's `comments` and SHALL NOT treat their presence as a difference requiring an update or replacement, provided every text in `required_comments` is still present in the log.

#### Scenario: Refresh records an out-of-band comment

- GIVEN an analyst adds a comment directly through the Kibana API, and every `required_comments` text remains present
- WHEN Terraform refreshes the resource
- THEN state's `comments` SHALL include that comment
- AND `required_comments` SHALL be unchanged

#### Scenario: Plan after that refresh is a no-op

- GIVEN the refresh has stored the out-of-band comment and configuration is unchanged
- WHEN Terraform plans
- THEN the plan SHALL be a no-op

#### Scenario: An analyst-written text satisfies the requirement

- GIVEN the stored log already contains the exact text `alpha`, written outside Terraform
- AND `required_comments` is `["alpha"]`
- WHEN update runs
- THEN the update request SHALL omit `comments`
- AND apply SHALL NOT append a second `alpha`

### Requirement: Import does not populate `required_comments` (REQ-025e)

On import, the resource SHALL populate `comments` from the API read, including every stored comment with its `id`, `comment`, `created_at`, and `created_by`. The resource SHALL NOT populate `required_comments` from the API; it SHALL remain null/unset after import until configuration sets it. No state upgrader SHALL copy a prior schema version's `comments` configuration or state into `required_comments`.

#### Scenario: Import leaves required_comments unset

- GIVEN an exception item with two stored comments is imported
- WHEN import completes and the first read runs
- THEN state SHALL contain `comments` with both stored comments and SHALL contain `required_comments = null`

#### Scenario: Next apply after import only appends missing texts

- GIVEN an imported resource with `required_comments` subsequently set in configuration to a text not already stored
- WHEN the next apply runs
- THEN only that missing text SHALL be appended; previously-stored comments SHALL be left untouched
