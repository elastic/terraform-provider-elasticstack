## Purpose

Continuously finds Kibana API request fields that Terraform resources built on the Kibana API can reach but never use, keeps only the gaps worth exposing, and files one grounded issue per gap group for research, without flooding the tracker or re-filing decided gaps.

## ADDED Requirements

### Requirement: Scanner runs on schedule, on Kibana client changes, and on demand
The Kibana API gap scanner workflow SHALL run on a daily schedule, on pushes to the default branch that change files under `generated/kbapi/`, and on `workflow_dispatch`. Every trigger SHALL apply the same issue-slot backpressure, triage budget, and memory rules. Runs SHALL be serialised: a run SHALL NOT start its analysis while another scanner run is in progress, so concurrent triggers cannot race on issue slots or memory.

#### Scenario: Generated Kibana client changes on main
- **GIVEN** a commit on the default branch that changes a file under `generated/kbapi/`
- **WHEN** the commit is pushed
- **THEN** the gap scanner workflow SHALL run

#### Scenario: Unrelated change on main
- **GIVEN** a commit on the default branch that changes no file under `generated/kbapi/`
- **WHEN** the commit is pushed
- **THEN** the gap scanner workflow SHALL NOT run because of that push

#### Scenario: Schedule and push trigger at the same time
- **GIVEN** a scanner run in progress
- **WHEN** another trigger fires
- **THEN** the second run SHALL NOT analyse or file until the first run has finished

#### Scenario: Maintainer dispatches the scanner
- **GIVEN** open issue slots are available
- **WHEN** a maintainer runs the workflow via `workflow_dispatch`
- **THEN** the scanner SHALL behave exactly as a scheduled run

### Requirement: Scanner covers resources built on the Kibana API only
The scanner SHALL analyse every Terraform resource registered by the provider whose implementation calls the Kibana API through the generated Kibana client, whatever its type-name prefix. This includes `elasticstack_kibana_*`, `elasticstack_fleet_*`, and `elasticstack_apm_*` resources built on that client. It SHALL NOT analyse data sources or resources that do not call the Kibana API, and it SHALL NOT report Kibana API endpoints that no resource calls.

#### Scenario: Fleet resource is analysed
- **GIVEN** the provider registers the resource `elasticstack_fleet_agent_policy`
- **WHEN** the scanner runs
- **THEN** that resource SHALL be included in candidate discovery

#### Scenario: APM resource on the Kibana API is analysed
- **GIVEN** the resource `elasticstack_apm_agent_configuration`, which calls the Kibana API
- **WHEN** the scanner runs
- **THEN** that resource SHALL be included in candidate discovery

#### Scenario: Elasticsearch resource is not analysed
- **GIVEN** a resource that calls only Elasticsearch APIs
- **WHEN** the scanner runs
- **THEN** that resource SHALL NOT produce candidates

#### Scenario: Data source is not analysed
- **GIVEN** the provider registers a data source with an `elasticstack_kibana_` prefix
- **WHEN** the scanner runs
- **THEN** that data source SHALL NOT produce candidates

#### Scenario: Endpoint used by no resource
- **GIVEN** a Kibana API endpoint that no Kibana or Fleet resource calls
- **WHEN** the scanner runs
- **THEN** no candidate SHALL be produced for that endpoint

### Requirement: Candidate discovery is deterministic and proves absence
For each in-scope resource, the scanner SHALL deterministically, without any LLM involvement, determine the Kibana API request-body fields and query parameters of the operations that resource calls, and report as a candidate every such field that the resource's implementation never uses. A candidate SHALL only be reported when its absence from the implementation is proven by the analysis. The following SHALL NOT be candidates:
- response-only fields;
- fields marked deprecated in the Kibana API specification;
- fields inside a request structure the resource passes through whole from an opaque user-supplied value (for example, a JSON-string attribute), which count as covered;
- fields the resource uses, including fields it only ever sets to fixed values.

Running discovery twice on the same repository state SHALL produce identical candidates.

#### Scenario: Unused request field becomes a candidate
- **GIVEN** a resource whose create operation accepts a request-body field that the resource's implementation never uses
- **WHEN** discovery runs
- **THEN** that field SHALL be reported as a candidate for that resource

#### Scenario: Used field is not a candidate
- **GIVEN** a request-body field that the resource's implementation reads or writes
- **WHEN** discovery runs
- **THEN** that field SHALL NOT be reported as a candidate

#### Scenario: Deprecated field is excluded
- **GIVEN** an unused request-body field marked deprecated in the Kibana API specification
- **WHEN** discovery runs
- **THEN** that field SHALL NOT be reported as a candidate

#### Scenario: Opaque passthrough counts as covered
- **GIVEN** a request structure that the resource populates whole from a user-supplied JSON value
- **WHEN** discovery runs
- **THEN** no field inside that structure SHALL be reported as a candidate

#### Scenario: Discovery is repeatable
- **GIVEN** an unchanged repository state
- **WHEN** discovery runs twice
- **THEN** both runs SHALL report the same candidates with the same identities

### Requirement: Candidates are grouped per resource and top-level request field with stable identity
Candidates SHALL be grouped into one gap group per (resource, top-level request field), where a top-level request field is a direct property of an operation's request body or a query parameter. A top-level field of the same kind (request-body property or query parameter) that appears under the same name in several of the resource's operation bodies or request-body variants (for example in both a create and an update body, or in every variant of a union body) SHALL form one group; the group's members keep their own specification locations. Each group SHALL have a stable key made of the resource type name and the Kibana API specification location of the top-level field (the lexicographically smallest location when the group spans several), and a structural fingerprint computed from the definitions of that field, including all nested fields, at every location the group spans. When a location is added that sorts before the current smallest one, or the smallest one is removed, the group's key changes and it is triaged once more; this is accepted. The fingerprint SHALL change when the field's structure changes (types, properties, enum values, required-ness, or deprecation) and SHALL NOT change when only descriptions or examples change.

#### Scenario: Nested unused fields share one group
- **GIVEN** a resource that leaves two nested fields under the same top-level request field unused
- **WHEN** candidates are grouped
- **THEN** both fields SHALL belong to a single gap group for that top-level field

#### Scenario: Same field in several bodies shares one group
- **GIVEN** a resource whose create and update request bodies both accept the same top-level field, which the resource never uses
- **WHEN** candidates are grouped
- **THEN** a single gap group SHALL exist for that field, keyed by the smaller of the two specification locations
- **AND** the group SHALL list the unused fields at both locations

#### Scenario: Description-only change keeps the fingerprint
- **GIVEN** a gap group with a recorded fingerprint
- **WHEN** the Kibana API specification changes only the description of a field in that group
- **THEN** the group's fingerprint SHALL be unchanged

#### Scenario: New nested field changes the fingerprint
- **GIVEN** a gap group with a recorded fingerprint
- **WHEN** the Kibana API specification adds a nested property under that top-level field
- **THEN** the group's fingerprint SHALL change

### Requirement: Triage may only remove candidates, and suppresses only with high confidence
An LLM triage step SHALL judge each gap group selected for triage and return exactly one verdict: `file` or `suppress`. Triage SHALL `suppress` only when it is highly confident the group should not be exposed in Terraform (for example, server-managed values, internal or telemetry fields, or settings that make no sense as infrastructure-as-code) and SHALL give a short reason. Any group triage is not highly confident about SHALL be filed. Every verdict, `file` or `suppress`, SHALL carry a short reason. Triage SHALL NOT introduce gaps, fields, or groups that discovery did not report, and a suppressed group SHALL NOT be filed.

#### Scenario: Server-managed field is suppressed
- **GIVEN** a gap group for a server-managed revision counter
- **WHEN** triage judges it with high confidence as not user-configurable
- **THEN** the verdict SHALL be `suppress` with a short reason
- **AND** no issue SHALL be filed for that group

#### Scenario: Uncertain group is filed
- **GIVEN** a gap group whose usefulness to Terraform users is unclear
- **WHEN** triage judges it
- **THEN** the verdict SHALL be `file`

#### Scenario: Triage cannot add gaps
- **GIVEN** a triage response that names a field not present in any discovered group
- **WHEN** the scanner processes triage output
- **THEN** that field SHALL NOT appear in any filed issue or memory record

### Requirement: Triage is ordered and budgeted per run
Each run SHALL compute available issue slots as the cap of 3 minus the number of open issues labelled `kibana-api-gap`, floored at zero. When no slots are available, the run SHALL NOT triage. Otherwise, the run SHALL consider only gap groups that have no current verdict in memory, ordered first by resource, with resources that have never had a group triaged first and then resources in order of when a group of theirs was last triaged (oldest first), ties broken by resource type name; and within a resource by group size (most unused fields first), ties broken by group key, and SHALL triage them one at a time until either the available slots are filled with `file` verdicts or 15 groups have been triaged in that run. `suppress` verdicts SHALL NOT consume a slot.

#### Scenario: No slots available
- **GIVEN** three open issues labelled `kibana-api-gap`
- **WHEN** the scanner runs
- **THEN** no triage SHALL occur and no issue SHALL be filed

#### Scenario: Suppressions do not consume slots
- **GIVEN** two available slots
- **AND** the first three ordered groups are triaged as `suppress`, `file`, `file`
- **WHEN** the run triages groups
- **THEN** two issues SHALL be filed
- **AND** triage SHALL stop after the third group

#### Scenario: Triage budget caps a run of suppressions
- **GIVEN** available slots
- **AND** every ordered group would be triaged as `suppress`
- **WHEN** the run triages groups
- **THEN** the run SHALL stop after 15 triaged groups

### Requirement: Filed gap groups become grounded issues dispatched to research
For each `file` verdict, the scanner SHALL create one issue whose title is prefixed with `[kibana-api-gap] ` and names the resource and the top-level field, labelled `kibana-api-gap` and `triaged`. The issue body SHALL identify the resource, list every unused field in the group with its Kibana API specification location, state the Kibana version in which the field was introduced when the specification provides one, include triage's reason for filing, and end with a visible, machine-readable marker line carrying the group key and fingerprint (not an HTML comment, which the safe-output sanitizer strips). After issue creation, the scanner SHALL dispatch `research-factory` once per created issue, as defined by the `ci-research-factory-dispatch-fanout` capability. The scanner SHALL NOT apply any factory trigger label.

#### Scenario: Filed issue is grounded
- **GIVEN** a `file` verdict for a group of two unused fields
- **WHEN** the issue is created
- **THEN** the issue body SHALL list both fields with their specification locations
- **AND** the issue SHALL carry the labels `kibana-api-gap` and `triaged`

#### Scenario: Version is unknown
- **GIVEN** a filed field whose specification entry has no introduced-in version
- **WHEN** the issue is created
- **THEN** the issue SHALL still be created, without an introduced-in version for that field

#### Scenario: Research is requested automatically
- **GIVEN** a run that creates two issues
- **WHEN** issue creation completes
- **THEN** `research-factory` SHALL be dispatched once for each of the two issues

### Requirement: Verdicts persist in repository memory and expire on structural change
The scanner SHALL persist one memory record per triaged gap group on a dedicated repository-memory branch. Each record SHALL contain the group key, the structural fingerprint, the verdict, the date, and either a reference to the issue filed for it (`file`) or the short reason (`suppress`). Every suppressed group SHALL have a record with its reason; this record is the sole audit trail for suppressions. A group with a record SHALL count as covered, and SHALL NOT be triaged or filed again, while its current fingerprint equals the recorded fingerprint, regardless of the state of any issue filed for it. When the current fingerprint differs from the recorded one, the record SHALL be treated as expired and the group SHALL be eligible for triage as if new. Groups that no longer appear in discovery SHALL NOT be triaged. The memory SHALL also record, per resource, when a group of that resource was last triaged, for ordering. Each filed issue SHALL carry its group key in a form the scanner can recognise, and the `file` record's reference to the issue SHALL be that group key.

#### Scenario: Suppression is recorded with a reason
- **GIVEN** a group triaged as `suppress` with reason "server-managed counter"
- **WHEN** the run persists memory
- **THEN** the memory SHALL contain a record for that group with verdict `suppress` and that reason

#### Scenario: Filed group is not re-filed after its issue is closed
- **GIVEN** a group recorded as `file` with a reference to issue #123
- **AND** issue #123 has been closed without the gap being fixed
- **AND** the group's fingerprint is unchanged
- **WHEN** the scanner runs
- **THEN** the group SHALL NOT be triaged or filed again

#### Scenario: Structural change expires a suppression
- **GIVEN** a group recorded as `suppress`
- **WHEN** the Kibana API specification adds an enum value to a field in that group
- **THEN** the group SHALL be eligible for triage again

#### Scenario: Maintainer audits suppressions
- **GIVEN** the scanner has suppressed groups in earlier runs
- **WHEN** a maintainer reads the memory branch
- **THEN** every suppressed group SHALL appear with its key, verdict, date, and reason

### Requirement: Scanner failures never file unproven or duplicate issues
If candidate discovery fails, the run SHALL create no issues and SHALL leave memory unchanged. A run's verdicts and issues SHALL be all-or-nothing with respect to the triage step: if triage does not complete successfully (for example, the agent fails or times out), the run SHALL create no issues and SHALL persist no memory, so every group it considered remains eligible for a later run. When triage completes successfully, memory SHALL be persisted after issue creation has been attempted, and every group with a `file` verdict SHALL have a `file` record. If issue creation fails or is blocked for such a group, the `file` record SHALL still stand and the group SHALL remain covered until its fingerprint changes; the scanner SHALL NOT look up issue existence or state to detect this, and this is an accepted limitation. If memory cannot be persisted after successful triage, the run SHALL fail visibly.

#### Scenario: Discovery fails
- **GIVEN** candidate discovery exits with an error
- **WHEN** the run continues
- **THEN** no issue SHALL be created
- **AND** the memory branch SHALL be unchanged

#### Scenario: Triage is interrupted
- **GIVEN** a run whose triage agent fails after judging four groups
- **WHEN** the run finishes
- **THEN** no issue SHALL be created by that run
- **AND** the memory branch SHALL be unchanged
- **AND** those four groups SHALL be eligible for triage in a later run

#### Scenario: Memory cannot be written
- **GIVEN** a run whose triage completed successfully
- **WHEN** persisting memory fails
- **THEN** the workflow run SHALL fail visibly

#### Scenario: Issue creation fails after the verdict is recorded
- **GIVEN** a run that completed triage with a `file` verdict for a group
- **AND** issue creation for that group failed
- **WHEN** later scanner runs occur and the group's fingerprint is unchanged
- **THEN** the group SHALL NOT be triaged or filed again
