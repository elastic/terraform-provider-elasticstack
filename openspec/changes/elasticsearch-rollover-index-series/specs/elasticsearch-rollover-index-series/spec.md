## ADDED Requirements

### Requirement: Schema and identity

The `elasticstack_elasticsearch_rollover_index_series` resource SHALL require `prefix` (string,
`RequiresReplace`) and `rollover_alias` (string, `RequiresReplace`), SHALL expose optional
`deletion_protection` (bool, default `true`), and SHALL expose computed `id` (string),
`write_index` (string, nullable), and `indices` (set of string). `id` SHALL be set to
`"<cluster_uuid>/<rollover_alias>"`. The resource SHALL NOT expose `settings`, `settings_raw`,
shard/replica attributes, mappings, additional aliases, alias filter/routing configuration,
lifecycle-policy selection, rollover conditions, a configurable initial index name, date math, or an
initial-generation override.

#### Scenario: Computed id is cluster-and-alias scoped
- **WHEN** the resource is created or adopted against cluster UUID `abc123` with `rollover_alias = "orders-write"`
- **THEN** `id` SHALL equal `"abc123/orders-write"`
- **AND** `id` SHALL NOT change when the effective writer or alias membership changes

#### Scenario: prefix change forces replacement
- **WHEN** `prefix` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: rollover_alias change forces replacement
- **WHEN** `rollover_alias` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: Unsupported attributes are rejected
- **WHEN** config sets an attribute outside `prefix`, `rollover_alias`, `deletion_protection`, or `timeouts`
- **THEN** Terraform SHALL reject the configuration at schema validation

### Requirement: Naming validation

The resource SHALL validate `prefix` and `rollover_alias` as literal Elasticsearch names — rejecting
wildcard, comma-separated, remote-cluster-qualified, or date-math expressions. The resource SHALL
validate the full generated bootstrap name `"<prefix>-000001"` (including total length) against
Elasticsearch index-naming rules, and SHALL validate `rollover_alias` against Elasticsearch
alias-naming rules rather than assuming identical constraints to index names.

#### Scenario: Wildcard prefix rejected
- **WHEN** `prefix = "orders-*"` is set
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Bootstrap name length validated in full
- **WHEN** `prefix` is long enough that `"<prefix>-000001"` exceeds the Elasticsearch index-name length limit
- **THEN** Terraform SHALL reject the plan with a validation error naming the generated bootstrap index name

#### Scenario: Comma-separated rollover_alias rejected
- **WHEN** `rollover_alias = "orders-write,other-alias"` is set
- **THEN** Terraform SHALL reject the plan with a validation error

### Requirement: Effective-writer resolution

The resource SHALL resolve the effective writer of `rollover_alias` using the following precedence,
and SHALL NOT select a writer by sorting names, by generation number, or by creation time:
1. A member with `is_write_index=true` is the writer.
2. A sole member with an unset write flag is the implicit writer.
3. A sole member with `is_write_index=false` and no other member marked true is NOT a writer (no
   effective writer).
4. Multiple members with no member marked `is_write_index=true` have no effective writer.

#### Scenario: Explicit writer flag
- **WHEN** `rollover_alias` has two members and one has `is_write_index=true`
- **THEN** `write_index` SHALL equal that member's concrete name

#### Scenario: Sole member with unset write flag is the implicit writer
- **WHEN** `rollover_alias` has exactly one member with no `is_write_index` value set
- **THEN** `write_index` SHALL equal that member's concrete name

#### Scenario: Sole member with explicit false has no writer
- **WHEN** `rollover_alias` has exactly one member with `is_write_index=false`
- **THEN** `write_index` SHALL be null

#### Scenario: Multiple members without a designated writer
- **WHEN** `rollover_alias` has two or more members and none has `is_write_index=true`
- **THEN** `write_index` SHALL be null

### Requirement: Create bootstraps a missing alias

When `rollover_alias` does not exist, Create SHALL send
`PUT /<prefix>-000001` with a body containing only
`{"aliases": {"<rollover_alias>": {"is_write_index": true}}}` — no index settings or mappings. If
`<prefix>-000001` already exists as an index, alias, or other incompatible object, Create SHALL fail
with the collision error and SHALL NOT adopt, delete, rename, or attach the alias to that object.
Create SHALL NOT scan `<prefix>-*` or any previously observed member names to decide whether
bootstrapping is allowed.

#### Scenario: Successful bootstrap of a missing alias
- **WHEN** `rollover_alias = "orders-write"` does not exist and `prefix = "orders"`
- **THEN** `PUT /orders-000001` SHALL be called with body `{"aliases": {"orders-write": {"is_write_index": true}}}`
- **AND** after Create, `write_index` SHALL equal `"orders-000001"` and `indices` SHALL equal `["orders-000001"]`

#### Scenario: Bootstrap name collision fails without mutation
- **WHEN** `rollover_alias` does not exist and `orders-000001` already exists as an unrelated index
- **THEN** Create SHALL fail with an actionable collision error
- **AND** no alias mutation SHALL be attempted on `orders-000001`

#### Scenario: Detached prefix-matching index does not block or satisfy bootstrap
- **WHEN** `rollover_alias` does not exist, `orders-000001` does not exist, but a detached `orders-000002` exists with no alias attached
- **THEN** Create SHALL bootstrap `orders-000001` normally
- **AND** the detached `orders-000002` SHALL NOT be adopted, renamed, or deleted

### Requirement: Create adopts a healthy existing alias automatically

When `rollover_alias` exists with exactly one effective writer whose name ends in `-<digits>`, Create
SHALL adopt the series automatically, without requiring an opt-in adoption flag, and SHALL NOT create
an index, add or rewrite any alias, normalize write flags, or mutate member settings. All current
members SHALL be adopted regardless of whether their names match `prefix`. Automatic adoption SHALL
acquire the ability to delete every then-current alias member once `deletion_protection` is disabled.

#### Scenario: Adoption of a healthy alias with a numerically-suffixed writer
- **WHEN** `rollover_alias = "orders-write"` already exists with writer `orders-000003`
- **THEN** Create SHALL perform no mutating API call
- **AND** `write_index` SHALL equal `"orders-000003"` and `indices` SHALL include `orders-000003`

#### Scenario: Adoption includes out-of-prefix members
- **WHEN** `rollover_alias` exists with a non-writer member named `unrelated-index-1` that does not match `prefix`
- **THEN** Create SHALL adopt the series including `unrelated-index-1` in `indices`
- **AND** `unrelated-index-1` SHALL be included in this resource's Delete scope once protection is disabled

### Requirement: Create rejects an unhealthy existing alias

When `rollover_alias` exists but has no effective writer, has a writer whose name does not end in
`-<digits>`, or is an unsupported alias type (for example a data-stream alias), Create SHALL fail
with an actionable diagnostic and SHALL NOT mutate any index or alias to force a healthy state.

#### Scenario: No effective writer blocks adoption
- **WHEN** `rollover_alias` exists with two members and neither is marked `is_write_index=true`
- **THEN** Create SHALL fail with a diagnostic identifying the missing effective writer
- **AND** no mutation SHALL be attempted

#### Scenario: Non-numeric writer name blocks adoption
- **WHEN** `rollover_alias` exists with a sole writer named `orders-current` (no `-<digits>` suffix)
- **THEN** Create SHALL fail with a diagnostic explaining that automatic rollover cannot derive a successor name
- **AND** no mutation SHALL be attempted

#### Scenario: Data-stream alias rejected
- **WHEN** `rollover_alias` resolves to a data-stream alias rather than an ordinary index alias
- **THEN** Create SHALL fail with a diagnostic stating the resource manages ordinary index aliases only

### Requirement: Read distinguishes health from existence

Read SHALL call the alias lookup for `rollover_alias`. When the alias is absent, Read SHALL remove
the resource from state. When the alias exists with a valid, numerically-suffixed writer, Read SHALL
refresh `indices` and `write_index` without warnings. When the alias exists but is unhealthy (no
effective writer, or a writer name lacking `-<digits>`), Read SHALL emit a warning diagnostic, retain
the resource in state, refresh `indices`, and set `write_index` to null or to the actual
non-conforming writer name (never hidden). Read SHALL treat authorization, connection, timeout,
decoding, or other unexpected API failures as errors, never as alias absence, and SHALL NOT discard
state on such errors. Read SHALL NOT query `<prefix>-*` or any previously observed membership when
the exact alias is absent.

#### Scenario: Alias absence removes state
- **WHEN** `rollover_alias` no longer exists on refresh
- **THEN** the resource SHALL be removed from Terraform state
- **AND** no lookup of `<prefix>-*` or previously known members SHALL be performed

#### Scenario: Bootstrap index deleted but alias survives
- **WHEN** `orders-000001` has been deleted by ILM but `rollover_alias` remains attached to `orders-000002`
- **THEN** Read SHALL report normal state with `write_index = "orders-000002"`
- **AND** Read SHALL NOT attempt to recreate or restore `orders-000001`

#### Scenario: Unhealthy alias produces a warning, not an error
- **WHEN** `rollover_alias` exists but has no effective writer on refresh
- **THEN** Read SHALL emit a warning diagnostic
- **AND** the resource SHALL remain in state with `write_index = null`

#### Scenario: Writer with non-conforming name is exposed, not hidden
- **WHEN** `rollover_alias`'s sole writer is named `orders-current`
- **THEN** Read SHALL emit a warning diagnostic
- **AND** `write_index` SHALL be set to `"orders-current"`

#### Scenario: API failure is not treated as absence
- **WHEN** the alias lookup fails with an authorization or transport error
- **THEN** Read SHALL return an error
- **AND** the resource SHALL NOT be removed from state

### Requirement: No configuration-drift repair

Read SHALL never restore `is_write_index=true` on the bootstrap index, never move the alias to a
guessed latest generation, and never re-add a member that has been removed from the alias. Changes to
the computed `indices` and `write_index` attributes observed between plans SHALL NOT be treated as
drift requiring an Update API call.

#### Scenario: Writer change between plans is not corrected
- **WHEN** `write_index` changes from `orders-000001` to `orders-000002` due to an external rollover
- **THEN** the next plan SHALL show only the computed attribute change
- **AND** no API call SHALL be made to restore `orders-000001` as the writer

### Requirement: Update has no in-place Elasticsearch mutation

Update SHALL persist changes to `deletion_protection` and the operation-timeouts attribute to
Terraform state, followed by a readback of membership and writer, without any Elasticsearch API call
tied to those attribute changes. Update SHALL NOT send a placeholder API call (such as
`_validate/query`) to simulate a mutation when no replaceable attribute has changed.

#### Scenario: Disabling deletion_protection updates state only
- **WHEN** `deletion_protection` is changed from `true` to `false`
- **THEN** no Elasticsearch API call SHALL be made as a result of that change
- **AND** state SHALL reflect `deletion_protection = false` after apply

### Requirement: Replacement requires destroy-before-create

Changing `prefix` or `rollover_alias` SHALL force replacement. The resource SHALL NOT support
`create_before_destroy` for this replacement: destroy SHALL run before the replacement create is
attempted, because automatic adoption could otherwise let the new instance adopt the alias while the
old instance still owns it, causing the old instance's destroy to delete the new instance's members.

#### Scenario: rollover_alias replacement destroys before creating
- **WHEN** `rollover_alias` changes from `"orders-write"` to `"orders-write-v2"` and both reference distinct existing resources
- **THEN** Terraform SHALL destroy the resource instance for `"orders-write"` before creating the instance for `"orders-write-v2"`

### Requirement: Deletion protection guards provider-initiated deletes

`deletion_protection` SHALL default to `true`. While enabled, Delete SHALL fail with an actionable
error and SHALL NOT mutate any index or alias, including during a forced replacement. Disabling
protection SHALL require a prior, separate `terraform apply` before a destroy or replacement that
deletes the resource is attempted in the same operation.

#### Scenario: Delete blocked while protection is enabled
- **WHEN** `terraform destroy` is run with `deletion_protection = true` in the last-applied config
- **THEN** Delete SHALL fail with an error directing the user to set `deletion_protection = false` and apply first
- **AND** no index SHALL be deleted

#### Scenario: Protection must be disabled in a prior apply
- **WHEN** a single operation both sets `deletion_protection = false` and replaces the resource
- **THEN** Delete SHALL still fail, because the unprotected state was not established by a prior, separate apply

#### Scenario: External alias deletion bypasses protection but is still observed safely
- **WHEN** an external actor deletes `rollover_alias` directly while the Terraform resource still has `deletion_protection = true`
- **THEN** the next Read SHALL remove the resource from state
- **AND** this removal SHALL NOT be treated as a protection violation

### Requirement: Delete removes current alias membership, never a prefix sweep

Delete SHALL, after passing the protection check, discover the exact alias's current concrete
membership (including hidden and closed indices) and delete exactly those concrete index names. Delete
SHALL NOT delete by alias name, by wildcard, or by prefix match. Delete SHALL treat a concrete member
already removed by ILM between discovery and deletion as a success (not an error) while still
surfacing permission or other non-absence failures. Delete SHALL re-discover membership and repeat
deletion within the delete timeout if a rollover introduces new members during the operation. If
cleanup cannot converge before timeout, Delete SHALL error and retain state for retry.

#### Scenario: Delete removes all current concrete members
- **WHEN** `rollover_alias` currently has members `orders-000002` (writer) and `orders-000003` (writer)
- **THEN** Delete SHALL delete both `orders-000002` and `orders-000003` by concrete name
- **AND** Delete SHALL NOT issue a delete request against `rollover_alias` or any wildcard pattern

#### Scenario: Hidden and closed members are included
- **WHEN** `rollover_alias` includes a closed member and a hidden member
- **THEN** Delete SHALL discover and delete both

#### Scenario: Out-of-prefix adopted member is deleted
- **WHEN** an adopted member `unrelated-index-1` (not matching `prefix`) is still attached to `rollover_alias`
- **THEN** Delete SHALL include `unrelated-index-1` in the concrete names it deletes

#### Scenario: Concurrently-deleted member does not fail Delete
- **WHEN** a discovered member is already deleted by ILM before the provider's delete request reaches it
- **THEN** Delete SHALL treat that member's absence as success
- **AND** Delete SHALL continue deleting the remaining discovered members

#### Scenario: Alias already absent makes Delete a no-op success
- **WHEN** `rollover_alias` does not exist at Delete time
- **THEN** Delete SHALL succeed without deleting any index
- **AND** Delete SHALL NOT attempt to locate or delete detached prefix-matching indices

#### Scenario: Rollover race during delete is re-discovered
- **WHEN** a rollover introduces a new alias member after initial discovery but before deletion completes, within the delete timeout
- **THEN** Delete SHALL re-discover membership and delete the newly observed member too

#### Scenario: Unhealthy writer does not block delete
- **WHEN** `rollover_alias` has no effective writer at Delete time
- **THEN** Delete SHALL still proceed to discover and delete current concrete members (subject to the protection check)

### Requirement: Stable composite identity and three-part import

The resource SHALL use `"<cluster_uuid>/<rollover_alias>"` as its Terraform state `id`, independent
of the writer, any index UUID, or generation number. Import SHALL accept the three-part identifier
`"<cluster_uuid>/<rollover_alias>/<prefix>"`, populating `rollover_alias` and `prefix` from the parsed
parts; all other attributes SHALL be populated by the subsequent Read per the Create/adoption and Read
contracts above. `deletion_protection` SHALL default to `true` for imported resources, identically to
normal Create.

#### Scenario: Import populates both required naming attributes
- **WHEN** `terraform import elasticstack_elasticsearch_rollover_index_series.orders '<cluster_uuid>/orders-write/orders'` is run
- **THEN** `rollover_alias` SHALL be set to `"orders-write"` and `prefix` SHALL be set to `"orders"`
- **AND** `id` SHALL be set to `"<cluster_uuid>/orders-write"`

#### Scenario: Import of an adopted series with a differently-named writer
- **WHEN** the imported alias's current writer name does not share `prefix`'s value
- **THEN** import SHALL still succeed, using the imported `prefix` only for any future bootstrap, never to infer or validate the current writer's name

#### Scenario: Malformed import identifier rejected
- **WHEN** an import ID is supplied without three `/`-separated segments
- **THEN** Terraform SHALL reject the import with an actionable error describing the expected `<cluster_uuid>/<rollover_alias>/<prefix>` format

### Requirement: Shared operation timeouts

The resource SHALL expose the standard entitycore `timeouts` attribute (create/read/update/delete)
and SHALL use it to bound Create/adoption polling and Delete's re-discovery loop, rather than
implementing resource-specific timeout parsing.

#### Scenario: Delete respects the configured delete timeout
- **WHEN** `timeouts.delete` is configured and membership re-discovery has not converged before it elapses
- **THEN** Delete SHALL stop retrying, return a timeout error, and retain state for a subsequent retry
