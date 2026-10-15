## 1. Bootstrap or adopt an alias-owned series

- [x] 1.1 Expose the registered `elasticstack_elasticsearch_rollover_index_series` resource through
  the Plugin Framework and shared Elasticsearch envelope with the schema, stable identity and
  operation timeouts specified in "Schema and identity" and "Shared operation timeouts". Verify
  provider-schema coverage for every attribute, replacement modifiers on both naming inputs, default
  protection, nullable computed writer, and rejection of unsupported settings, mappings and extra aliases.
- [x] 1.2 Enforce "Naming validation" for both configured and imported names, validating the entire
  generated bootstrap name and applying alias-appropriate constraints. Verify unit coverage for
  length/charset boundaries and literal-name rejection of wildcard, comma, remote-cluster and date-math
  syntax rather than treating those inputs as discovery expressions.
- [x] 1.3 Deliver shared exact-alias discovery and "Effective-writer resolution" without membership
  reconciliation or generation guessing. Verify tests for explicit writers, sole implicit writers,
  explicit false, multiple members with no writer, and non-numeric writer suffixes; resolve target
  types independently of alias-response shape and reject data streams before adoption, Read/import
  hydration or deletion, including numerically suffixed data-stream names with explicit writers.
- [x] 1.4 Deliver "Create bootstraps a missing alias" with an alias-only initial index, fresh state
  readback and no survivor probes. Verify unit/API and directory-backed acceptance coverage for
  template-inherited settings, fixed generation naming, detached later-generation coexistence, and
  occupied index/alias bootstrap names failing without repair, fallback generation or mutation of
  existing objects. Verify interrupted bootstrap can be rediscovered rather than blindly recreated.
- [x] 1.5 Deliver automatic healthy adoption and strict unhealthy-alias rejection without an opt-in
  flag or mutating call, including out-of-prefix members. Verify unit and acceptance cases for each
  adoption branch and live readback of canonical identity, membership and writer. Deliver resource
  documentation explaining that adoption acquires deletion rights over every current member, with
  protection enabled by default and no competing index/alias ownership.

## 2. Observe lifecycle changes without repairing them

- [ ] 2.1 Deliver "Read distinguishes health from existence" and "No configuration-drift repair".
  Verify tests that refresh membership and actual writer, emit warning-only diagnostics for missing
  or non-numeric writers, expose null only when no writer exists, and never restore flags or members.
  Verify data-stream rejection and authorization, timeout, decoding and transport failures retain state
  as errors rather than becoming absence or writer-health warnings.
- [ ] 2.2 Preserve ordinary refresh across external rollover and bootstrap-index deletion. Verify
  directory-backed live acceptance steps that the writer and member outputs follow Elasticsearch
  while the next plan performs no corrective mutation or recreation of the bootstrap index.
- [ ] 2.3 Deliver unconditional state removal for exact-alias absence, without prefix or last-known
  member probes. Verify acceptance sequences for alias removal with detached data, a successful new
  bootstrap when the initial name is free, and collision without repair when it remains occupied;
  assert detached former members are unchanged in both cases.
- [ ] 2.4 Deliver state-only protection/timeout updates with required non-mutating readback. Verify
  API-level assertions that Update performs the required alias discovery but no mutating or placeholder
  API calls, remains successful with writer-health warnings, and preserves the applied Terraform-only
  fields. Document alias-defined existence, automatic rebootstrap, and the absence of survivor cleanup.
- [ ] 2.5 Deliver replacement behavior for changed prefix or alias while documenting "Replacement
  ordering is a user precondition". Verify replacement plans for both inputs and a same-alias prefix
  replacement using normal destroy-before-create ordering. Document that Terraform Core, not the
  provider, controls ordering and that direct or propagated create-before-destroy is unsafe; do not
  introduce an unsupported claim of provider enforcement.

## 3. Destroy only current alias members under explicit protection

- [ ] 3.1 Deliver "Deletion protection guards provider-initiated deletes" using prior applied state.
  Verify live and unit coverage for protected destroy/replacement, a prior apply disabling protection,
  rejection when protection is only disabled in the replacing operation, and external alias removal
  being observed as absence even from protected state.
- [ ] 3.2 Deliver "Delete removes current alias membership, never a prefix sweep" through fresh,
  type-checked discovery and concrete-name deletion. Verify live acceptance coverage with the normal
  destructive-name safeguard enabled for out-of-prefix, hidden, closed and lifecycle-renamed members;
  assert detached prefix-matching indices survive. Verify missing-alias no-op success without survivor
  lookups and concrete-member cleanup regardless of writer health.
- [ ] 3.3 Deliver bounded rediscovery for rollover during cleanup. Verify deterministic unit/mock
  interleavings in which rollover occurs after enumeration and the newly observed member is deleted,
  a member disappears before deletion and is tolerated, and any data-stream target or non-absence
  discovery/deletion error is surfaced before unsafe deletion. Verify no alias/wildcard deletion,
  cluster-safeguard changes, global ILM stop or policy-wide inventory is used.
- [ ] 3.4 Deliver recoverable partial-failure and timeout behavior. Verify deterministic tests for a
  successful member deletion followed by a permission failure retaining state, then a fresh successful
  retry; verify continual new members lead to a timeout error rather than false success. Use the shared
  configured delete deadline or its standard default, including retry delays, with no resource-specific
  minimum or independent deadline extension.
- [ ] 3.5 Deliver the documentation required by "Operational cleanup boundaries are documented".
  Verify guidance explains quiescing producers, waiting for in-flight replacements, and the limits of
  discovering not-yet-attached transformation targets. Explain that direct index deletion leaves
  searchable snapshots and that detached indices or other series can still block shared-policy deletion;
  neither situation authorizes broader cleanup. Include prior-apply protection and replacement warnings
  in the generated resource documentation.

## 4. Import with stable identity and safe Terraform-only state

- [ ] 4.1 Deliver "Stable composite identity and three-part import". Before Read, explicitly populate
  canonical two-part `id`, both required names and protection enabled from the validated three-part
  input; parsing alone must not be treated as state assignment. Verify unit tests for missing, empty,
  extra and invalid naming segments, exact canonical ID rather than raw import input, and preservation
  of imported fields across readback.
- [ ] 4.2 Verify live import after rollover with a writer whose naming differs from the imported
  prefix, without inferring or replacing the configured bootstrap prefix. Verify an immediate destroy
  without an intervening apply remains protected, and that applying protection disabled is preserved by
  later reads and allows intentional cleanup. Verify unsupported data-stream import hydration errors
  without adoption or mutation.
- [ ] 4.3 Deliver import instructions and a self-contained directory-backed example using the explicit
  three-part identifier, canonical state identity and default protection. Verify generated documentation
  and import acceptance checks match the same contract and do not require an opt-in flag for healthy
  Create/adoption.

## 5. Interoperate with supported lifecycle environments

- [ ] 5.1 Deliver a live policy/template/series example and acceptance coverage proving bootstrap
  inheritance and reverse dependency cleanup ordering. Verify deleting the owned series allows its
  unshared policy to be deleted while an independently retained policy reference still prevents policy
  deletion without causing unrelated index cleanup. Keep the rollover alias out of the template's
  aliases section and document settings/mappings as externally template-managed.
- [ ] 5.2 Verify real ILM shrink interoperability in a suitable live acceptance environment: the
  replacement retains alias membership and is removed on series deletion. Use a targeted step move
  only on fixture-owned indices if needed, with cleanup and no global ILM changes. Supplement the live
  case with deterministic mock shrink/searchable-snapshot alias-transfer responses; document mock or
  source-backed snapshot evidence separately from live evidence rather than substituting mocks for
  all real shrink coverage.
- [ ] 5.3 Validate lifecycle acceptance behavior against the provider's supported Elasticsearch version
  matrix and record versions and results as implementation evidence. Do not infer a new minimum from
  the local 9.4.0 experiment, replace matrix validation with merely choosing a minimum, or silently skip
  unsupported behavior; surface evidence and obtain a support decision before narrowing compatibility.
- [ ] 5.4 Deliver the integrated resource, examples and generated documentation with every delta-spec
  scenario mapped to tests. Verify repository OpenSpec validation, build, focused unit suites, lint,
  example validation and targeted live acceptance coverage are green, preserving directory-backed
  fixtures and the normal provider factories. Record any environment-limited verification explicitly
  rather than marking incomplete lifecycle or compatibility coverage complete.
