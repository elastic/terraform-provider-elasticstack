## Purpose

Define an alias-owned rollover index series that Terraform can bootstrap or adopt, observe without
repairing lifecycle changes, and destroy by current concrete alias membership with deletion safeguards.

## ADDED Requirements

### Requirement: Schema and identity

The resource SHALL require string `prefix` and `rollover_alias`, both replacement-only, and expose
optional boolean `deletion_protection` defaulting to `true`. Computed outputs SHALL be string `id`,
nullable string `write_index`, and set-of-string `indices`. Identity SHALL be
`"<cluster_uuid>/<rollover_alias>"`, stable across changes to writer and membership.

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

### Requirement: Unsupported index configuration is excluded

The resource SHALL NOT expose settings, raw settings, shard/replica attributes, mappings, additional
aliases, alias filter/routing configuration, lifecycle-policy selection, rollover conditions,
configurable initial index names, date math, or an initial-generation override. Templates remain
responsible for index settings and mappings.

#### Scenario: Unsupported index configuration is rejected
- **GIVEN** configuration for a rollover index series
- **WHEN** it supplies index settings, mappings, or configurable additional aliases
- **THEN** Terraform SHALL reject those unsupported inputs at schema validation

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

A member explicitly marked `is_write_index=true` SHALL be the effective writer. A sole member with
an unset flag SHALL be an implicit writer; explicit `false` SHALL NOT qualify. Multiple members
without a designated writer SHALL have no effective writer. The provider SHALL NOT choose by name
sorting, generation number, or creation time.

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

When the exact alias is absent, Create SHALL create `<prefix>-000001` with only its write-alias
definition, without settings or mappings. An occupied bootstrap name SHALL cause a collision error,
without adopting, deleting, renaming, or attaching the alias to that object or choosing another
generation. Create SHALL NOT scan the prefix namespace or previously observed members.

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

Create SHALL automatically adopt an ordinary-index alias with exactly one effective writer ending
in `-<digits>`, without an opt-in flag or any mutation. Adoption SHALL include all current members,
regardless of prefix, acquiring deletion ownership once protection is disabled. It SHALL NOT create
indices, rewrite aliases, normalize write flags, or change settings.

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
Shared alias discovery SHALL resolve target types rather than infer them from the `GetAlias` response
shape; any data-stream target SHALL be rejected before adoption.

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

#### Scenario: Numerically suffixed data-stream target rejected
- **WHEN** `rollover_alias` resolves to a data stream named `orders-000001` with `is_write_index=true`
- **THEN** Create SHALL reject the target as a data stream despite its alias-shaped response and numeric name
- **AND** no mutation SHALL be attempted

### Requirement: Read distinguishes health from existence

Read SHALL discover exact-alias membership and writer. Absence SHALL remove state without prefix or
former-member probes. A healthy alias SHALL refresh outputs without warnings. Missing or
non-numeric-suffixed writers SHALL produce warnings only: retain state, refresh members, and expose
null for no writer or the actual incompatible name. Import hydration SHALL use the same Read behavior.

#### Scenario: Alias absence removes state
- **WHEN** `rollover_alias` no longer exists on refresh
- **THEN** the resource SHALL be removed from Terraform state
- **AND** no lookup of `<prefix>-*` or previously known members SHALL be performed

#### Scenario: Alias absence with detached data allows rebootstrap
- **GIVEN** a managed alias whose bootstrap index has been deleted and a surviving former member `orders-000002`
- **AND** an external actor removes the alias from all remaining members
- **WHEN** Terraform refreshes and subsequently applies the unchanged configuration
- **THEN** Read SHALL remove the resource from state without probing former members
- **AND** Create SHALL bootstrap `orders-000001` if that name is available
- **AND** detached `orders-000002` SHALL remain unchanged and outside the new alias

#### Scenario: Alias absence followed by a bootstrap collision does not repair data
- **GIVEN** the alias has been removed but its former bootstrap index `orders-000001` remains
- **WHEN** Read removes the resource from state and the next apply attempts Create
- **THEN** Create SHALL fail on the occupied bootstrap name without attaching an alias or modifying the existing index
- **AND** no prefix or former-member survivor lookup SHALL be performed

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

### Requirement: Read errors do not imply alias absence

Read SHALL resolve alias target types and reject data streams with an error, retaining state.
Authorization, connection, timeout, decoding and unexpected API failures SHALL be errors, never
absence or writer-health warnings. These rules SHALL also apply during import hydration.

#### Scenario: Data-stream target rejected on refresh
- **GIVEN** resource state referring to an alias now targeting a data stream
- **WHEN** Read refreshes the resource
- **THEN** Read SHALL return an unsupported-target error and retain state
- **AND** no index or alias mutation SHALL be attempted

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
Terraform state, followed by a readback of membership and writer. The readback GET is allowed; no
mutating Elasticsearch API call SHALL be tied to those attribute changes. Update SHALL NOT send a
placeholder API call (such as `_validate/query`) to simulate a mutation when no replaceable attribute
has changed.

#### Scenario: Disabling deletion_protection updates state only
- **WHEN** `deletion_protection` is changed from `true` to `false`
- **THEN** Update SHALL perform no mutating Elasticsearch API call as a result of that change
- **AND** Update SHALL read back membership and writer without mutating them
- **AND** state SHALL reflect `deletion_protection = false` after apply

### Requirement: Replacement ordering is a user precondition

Changing either naming field SHALL force replacement. Documentation SHALL require users to ensure
destroy-before-create: same-alias adoption followed by old-instance deletion can delete replacement
members. Terraform Core controls ordering; the provider SHALL NOT claim to enforce it or prevent
direct or propagated `create_before_destroy`.

#### Scenario: Prefix replacement with the same alias destroys before creating
- **WHEN** `prefix` changes from `"orders"` to `"orders-v2"` while `rollover_alias` remains `"orders-write"` and the user has ensured the lifecycle graph orders destroy before create
- **THEN** the old resource instance SHALL be destroyed before the replacement instance is created
- **AND** the documentation SHALL warn that `create_before_destroy` can violate this precondition, including when propagated through dependencies

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

After the protection check, Delete SHALL discover current exact-alias membership, including hidden
and closed indices, and delete only those ordinary concrete names. It SHALL reject data-stream
targets before mutation. It SHALL NOT delete by alias, wildcard, prefix match, or stale inventory.
Detached indices SHALL remain untouched. Genuine alias absence SHALL succeed without survivor probes;
writer health SHALL NOT block member cleanup.

#### Scenario: Delete removes all current concrete members
- **GIVEN** deletion protection has been disabled in a prior apply
- **AND** `rollover_alias` has non-writer member `orders-000002` and writer `orders-000003`
- **WHEN** Delete runs
- **THEN** Delete SHALL delete both `orders-000002` and `orders-000003` by concrete name
- **AND** Delete SHALL NOT issue a delete request against `rollover_alias` or any wildcard pattern

#### Scenario: Hidden and closed members are included
- **WHEN** `rollover_alias` includes a closed member and a hidden member
- **THEN** Delete SHALL discover and delete both

#### Scenario: Data-stream target blocks Delete before mutation
- **WHEN** `rollover_alias` resolves to a numerically suffixed data stream with an explicit writer
- **THEN** Delete SHALL return an error identifying the unsupported data-stream target
- **AND** Delete SHALL NOT delete the data stream or any other target

#### Scenario: Out-of-prefix adopted member is deleted
- **WHEN** an adopted member `unrelated-index-1` (not matching `prefix`) is still attached to `rollover_alias`
- **THEN** Delete SHALL include `unrelated-index-1` in the concrete names it deletes

#### Scenario: Detached prefix-matching indices are excluded from Delete
- **GIVEN** deletion protection has been disabled and `rollover_alias` owns `orders-000003`
- **AND** `orders-000001` and `orders-000002` exist but are detached from the alias
- **WHEN** Delete runs
- **THEN** only currently attached concrete members SHALL be deleted
- **AND** both detached prefix-matching indices SHALL remain unchanged

#### Scenario: Lifecycle-renamed member remains in Delete scope
- **GIVEN** deletion protection has been disabled and ILM has replaced a former member with `shrink-id-orders-000001`
- **AND** the replacement and writer `orders-000002` are attached to `rollover_alias`
- **WHEN** Delete runs
- **THEN** both current concrete members SHALL be deleted regardless of the replacement's name
- **AND** the provider SHALL NOT try to restore or delete through the historical original-name alias

#### Scenario: Concurrently-deleted member does not fail Delete
- **WHEN** a discovered member is already deleted by ILM before the provider's delete request reaches it
- **THEN** Delete SHALL treat that member's absence as success
- **AND** Delete SHALL continue deleting the remaining discovered members

#### Scenario: Alias already absent makes Delete a no-op success
- **WHEN** `rollover_alias` does not exist at Delete time
- **THEN** Delete SHALL succeed without deleting any index
- **AND** Delete SHALL NOT attempt to locate or delete detached prefix-matching indices

### Requirement: Delete retries preserve failures and current membership

Delete SHALL tolerate members already deleted by ILM but surface permission and other non-absence
failures. It SHALL rediscover and delete new members during rollover races within the delete timeout.
Failure or non-convergence SHALL retain state and report an error rather than success while observed
members remain. A subsequent retry SHALL rediscover current membership rather than require stale
members to exist.

#### Scenario: Rollover race during delete is re-discovered
- **WHEN** a rollover introduces a new alias member after initial discovery but before deletion completes, within the delete timeout
- **THEN** Delete SHALL re-discover membership and delete the newly observed member too

#### Scenario: Partial deletion failure retains state for a fresh retry
- **GIVEN** deletion protection has been disabled and the alias has multiple concrete members
- **AND** one member is deleted successfully before a permission error prevents deleting another
- **WHEN** Delete receives the permission error
- **THEN** Delete SHALL report failure and retain resource state
- **AND** it SHALL NOT treat the permission error as absence or report cleanup success
- **WHEN** permission is restored and Delete is retried
- **THEN** it SHALL rediscover current membership and delete remaining members without requiring already-deleted indices to exist

#### Scenario: Unhealthy writer does not block delete
- **WHEN** `rollover_alias` has no effective writer at Delete time
- **THEN** Delete SHALL still proceed to discover and delete current concrete members (subject to the protection check)

### Requirement: Stable composite identity and three-part import

Import SHALL accept `<cluster_uuid>/<rollover_alias>/<prefix>` and explicitly set canonical
`id = <cluster_uuid>/<rollover_alias>`, both names, and protection `true` before Read. The raw
three-part input SHALL NOT become `id`. Read SHALL preserve these fields and subsequently applied
values, independent of writer naming, index UUIDs or generation. Import SHALL reject malformed names
and segments and follow Read's data-stream rejection contract.

#### Scenario: Import populates both required naming attributes
- **WHEN** `terraform import elasticstack_elasticsearch_rollover_index_series.orders '<cluster_uuid>/orders-write/orders'` is run
- **THEN** `rollover_alias` SHALL be set to `"orders-write"` and `prefix` SHALL be set to `"orders"`
- **AND** `id` SHALL be set to `"<cluster_uuid>/orders-write"`
- **AND** `deletion_protection` SHALL be initialized to `true`

#### Scenario: Import followed immediately by destroy remains protected
- **WHEN** a resource is imported and destroy is run without an intervening apply
- **THEN** Delete SHALL fail because imported state has `deletion_protection = true`
- **AND** no index SHALL be deleted

#### Scenario: Import of an adopted series with a differently-named writer
- **WHEN** the imported alias's current writer name does not share `prefix`'s value
- **THEN** import SHALL still succeed, using the imported `prefix` only for any future bootstrap, never to infer or validate the current writer's name

#### Scenario: Malformed import identifier rejected
- **GIVEN** an import identifier with a missing, empty, or extra segment, or an invalid literal naming field
- **WHEN** import is attempted
- **THEN** Terraform SHALL reject the import with an actionable error describing the expected `<cluster_uuid>/<rollover_alias>/<prefix>` format and naming constraints

#### Scenario: Data-stream target rejected during import hydration
- **GIVEN** a well-formed import identifier whose alias targets a data stream
- **WHEN** import hydrates the resource through Read
- **THEN** import SHALL fail with an unsupported-target diagnostic
- **AND** it SHALL NOT adopt or mutate the data stream

### Requirement: Shared operation timeouts

The resource SHALL expose the standard entitycore `timeouts` attribute (create/read/update/delete)
and SHALL use it to bound Create/adoption polling and Delete's re-discovery loop, rather than
implementing resource-specific timeout parsing.

#### Scenario: Delete respects the configured delete timeout
- **GIVEN** deletion protection is disabled and repeated rollovers keep introducing members
- **AND** `timeouts.delete` sets the operation deadline
- **WHEN** membership rediscovery has not converged before that deadline
- **THEN** Delete SHALL stop retrying, return a timeout error, and retain state for a subsequent retry
- **AND** retry delays SHALL NOT extend the operation beyond the configured timeout

### Requirement: Operational cleanup boundaries are documented

Documentation SHALL require quiesced producers and completion of in-flight lifecycle replacements
before destroy. It SHALL disclose that alias-member deletion cannot guarantee cleanup of unattached
transformation targets, detached indices, other series' policy references or backing searchable
snapshots; such data needs separate cleanup.

#### Scenario: Concurrent transformation limits are explicit
- **GIVEN** a user preparing to destroy a series with active lifecycle replacements
- **WHEN** they consult the resource's deletion guidance
- **THEN** it SHALL instruct them to quiesce producers and let in-flight replacements finish
- **AND** it SHALL explain that membership rechecks cannot guarantee discovery of targets not yet attached to the alias

### Requirement: Cleanup never broadens ownership or weakens safeguards

The resource SHALL NOT broaden inventory, remove unrelated policy references, delete snapshots,
stop ILM globally, or weaken cluster deletion safeguards to compensate for incomplete historical
cleanup. Documentation SHALL disclose that surviving references can still block policy deletion.

#### Scenario: Remaining data and policy references are not claimed as cleaned up
- **GIVEN** resource deletion guidance for a series with searchable snapshots or detached or shared-policy indices
- **WHEN** the guidance describes successful series deletion
- **THEN** it SHALL state that snapshots and detached indices require separate cleanup
- **AND** it SHALL state that other references may still block policy deletion without authorizing unrelated cleanup
