# Delta spec: `kibana-security-exception-item` — `required_comments` and append-only `comments` log

Extends the main spec at `openspec/specs/kibana-security-exception-item/spec.md`.

## MODIFIED Requirements

### Requirement: Mapping — comments on create and update (REQ-025)

The `comments` attribute SHALL be `Computed`-only and SHALL NOT be settable in configuration; it represents the full comment log Kibana has stored for the item, including comments created outside Terraform. The resource SHALL instead accept an optional `required_comments` attribute (list of strings) that names comment texts that MUST be present somewhere in the log.

On create, when `required_comments` is known and non-empty, the resource SHALL include each text in the create request body, in configuration order, using the create-comment API shape (`{comment}`, no `id`). When `required_comments` is null or empty, the resource SHALL omit `comments` from the create request body.

On update, the resource SHALL first perform a fresh read of the item's current comments, then compute `missing` as the `required_comments` texts (exact, case-sensitive string match) not present among the currently-stored comment texts. When `missing` is empty, the resource SHALL omit `comments` from the update request body, even when other attributes changed. When `missing` is non-empty, the resource SHALL set the update request's `comments` to exactly the `missing` texts, in configuration order, each using the update-comment API shape (`{comment}`, no `id`); it SHALL NOT echo back existing comments or their ids.

Computed comment `id`, `comment`, `created_at`, and `created_by` values SHALL be populated from the API response on every read, for every comment currently stored (not only ones Terraform created).

#### Scenario: Comment id populated from API

- GIVEN `required_comments` configured with texts not yet present
- WHEN the resource is read after create or update
- THEN each comment in state's `comments` SHALL include the Kibana-assigned `id`, the stored `comment` text, `created_at`, and `created_by`

#### Scenario: No-op update omits comments

- GIVEN every text in `required_comments` is already present among the item's stored comments
- WHEN an update is triggered by an unrelated attribute change (e.g. `name`)
- THEN the update request body SHALL NOT include a `comments` field

#### Scenario: Update sends only missing texts

- GIVEN stored comments are `alpha`, `beta` and `required_comments` is `["alpha", "gamma"]`
- WHEN update runs
- THEN the update request's `comments` SHALL be exactly `[{comment: "gamma"}]`, with no `id` and no echo of `alpha` or `beta`

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

Neither `comments` nor `required_comments` SHALL carry a `RequiresReplace` plan modifier, and no custom plan modifier SHALL force replacement based on a diff in either attribute. A change to `required_comments` — addition, removal, reordering, or text edit — SHALL always be planned as an in-place update, never a destroy-and-recreate, regardless of the shape of the corresponding change to the stored comment log.

#### Scenario: Removing a required text does not replace the resource

- GIVEN state has `required_comments = ["alpha", "beta"]` and stored comments include both texts
- WHEN configuration changes `required_comments` to `["alpha"]`
- THEN the plan SHALL show an in-place update (not a destroy/create) and SHALL NOT remove `beta` from the stored comment log

#### Scenario: Editing a required text does not replace the resource

- GIVEN state has `required_comments = ["alpha"]` and a stored comment with text `alpha`
- WHEN configuration changes `required_comments` to `["alpha-revised"]`
- THEN the plan SHALL show an in-place update, and apply SHALL append `alpha-revised` as a new comment while leaving the original `alpha` comment stored

### Requirement: Plan-time unknown handling for `comments` (REQ-025c)

`comments` SHALL be planned as unknown whenever apply could append at least one new comment: on create, and on update when any `required_comments` text (known at plan time) is not already present among the prior state's stored comment texts. When every `required_comments` text is already present in the prior state's stored comment texts, `comments` SHALL be planned as unchanged from the prior state value.

#### Scenario: Plan is unknown when a new required text will be appended

- GIVEN state's stored comments are `alpha` and configuration changes `required_comments` to include `beta` (not previously present)
- WHEN Terraform plans the update
- THEN `comments` SHALL be shown as unknown (`(known after apply)`) in the plan

#### Scenario: Plan keeps comments unchanged when nothing will be appended

- GIVEN state's stored comments are `alpha`, `beta` and `required_comments` is reduced from `["alpha", "beta"]` to `["alpha"]`
- WHEN Terraform plans the update
- THEN `comments` SHALL be shown as unchanged from state, and the plan SHALL show only the `required_comments` edit

### Requirement: Out-of-band comments do not force drift or replacement (REQ-025d)

When a read observes comments in the stored log that were not created by Terraform (added directly against the Kibana API), the resource SHALL include them in state's `comments` and SHALL NOT treat their presence as a difference requiring an update or replacement, provided every text in `required_comments` is still present in the log.

#### Scenario: Out-of-band comment tolerated

- GIVEN an analyst adds a comment directly through the Kibana API between two Terraform runs, and every `required_comments` text remains present
- WHEN Terraform runs a plan
- THEN the plan SHALL be a no-op with respect to `comments` and `required_comments`, and the out-of-band comment SHALL appear in state's `comments` after the next read

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
