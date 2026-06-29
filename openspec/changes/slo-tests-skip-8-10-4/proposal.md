# Proposal: Skip SLO acceptance tests on Kibana 8.10.4

## Summary

Three SLO acceptance tests fail on Elastic Stack **8.10.4** because the Kibana SLO backend
returns HTTP 500 (`illegal_argument_exception: must specify at least one document in [docs]`)
when creating an SLO whose source index is empty. This is a known server-side constraint on
8.10.x that was resolved in later patch releases. The provider code is not at fault.

The fix is to add `!=8.10.4` to the version-constraint strings for the three affected tests,
following the existing pattern already established for 8.11.x Kibana SLO bugs.

## Failing tests

| Test | Current constraint | Proposed addition |
|------|--------------------|-------------------|
| `TestAccResourceSlo_36_char_slo_id` | `>=8.9.0,!=8.11.{0–4},<8.16.0` | `!=8.10.4` |
| `TestAccResourceSlo_kql_custom_indicator_basic` | `SLOKqlAccTestConstraints` (>=8.9.0,!=8.11.{0–4}) | `!=8.10.4` added to constant |
| `TestAccResourceSloFromSDK` | `>=8.9.0,!=8.11.{0–4}` | `!=8.10.4` |

`TestAccResourceSlo_kql_custom_indicator_basic` uses `SLOKqlAccTestConstraints` from
`constants.go`; updating the shared constant also covers any other test that references it
(e.g. the step-level `skipKqlSLO` and `skipKqlSLOFleetStep` closures). `SLOKqlFleetAccTestConstraints`
starts at `>=8.10.0` and should receive the same `!=8.10.4` exclusion.

## Approach

**Version constraint exclusion** — the same surgical approach the codebase already uses for
8.11.x. No Terraform provider logic, no new resources, no test data changes.

Files to change:
- `internal/kibana/slo/constants.go` — add `!=8.10.4` to `SLOKqlAccTestConstraints` and
  `SLOKqlFleetAccTestConstraints`.
- `internal/kibana/slo/acc_test.go` — add `!=8.10.4` to the inline constraint strings in
  `TestAccResourceSlo_36_char_slo_id` and `TestAccResourceSloFromSDK`.

## Out of scope

- Fixing the underlying Kibana 8.10.4 behaviour (upstream issue).
- Seeding test indices with documents to work around the constraint (adds complexity and CI
  dependencies not warranted for an old minor version).
- Adding a new `elasticstack_elasticsearch_document` provider resource.

## References

- Issue: #3951
- `internal/kibana/slo/acc_test.go`
- `internal/kibana/slo/constants.go`
