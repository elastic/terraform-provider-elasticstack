# Design: Skip SLO acceptance tests on Kibana 8.10.4

## Root cause

Kibana 8.10.4 rejects SLO creation with HTTP 500 / `illegal_argument_exception: must specify
at least one document in [docs]` when the indicator's source index is empty. The failure occurs
during the initial transform backfill setup that Kibana performs server-side. Later 8.x patch
releases tolerate empty indices; the `prevent_initial_backfill` option (>=8.15.0) was introduced
separately to allow skipping this phase entirely.

The three failing tests (`TestAccResourceSlo_36_char_slo_id`,
`TestAccResourceSlo_kql_custom_indicator_basic`, `TestAccResourceSloFromSDK`) all create a
fresh, empty Elasticsearch index and immediately create an SLO against it — hitting this 8.10.4
constraint.

## Design decision

Use version-constraint exclusion (`!=8.10.4`) on the affected tests. This matches the exact
pattern established for 8.11.x:

```
>=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4
```

becomes:

```
>=8.9.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4
```

The `SLOKqlAccTestConstraints` and `SLOKqlFleetAccTestConstraints` constants in `constants.go`
are updated so that every test consuming them picks up the exclusion automatically. The two tests
with inline constraint strings (`TestAccResourceSlo_36_char_slo_id` and
`TestAccResourceSloFromSDK`) are updated directly.

## Scope of `!=8.10.4` vs a range

The research found no evidence that 8.10.0–8.10.3 exhibit the same failure. A single-version
pin is therefore preferred to keep the constraint readable and to avoid accidentally excluding
versions that work correctly. If future CI runs confirm 8.10.0–8.10.3 fail, the constraint
can be widened to `>=8.10.0,<=8.10.4` at that time.

## Alternative considered

**Seed index with a document** — rejected. It would require `local-exec` + `curl` or a new
provider resource, would add CI environment dependencies, and cannot easily be applied to
`TestAccResourceSloFromSDK` step 1 (which uses an external provider version). The 8.10.x line
is old and unlikely to attract new CI matrix targets.

## Open questions

1. Are all of 8.10.0–8.10.3 affected by the same empty-index SLO creation failure, or is it
   specific to 8.10.4? Determines whether a range exclusion is needed vs a single-version pin.
2. Are other SLO acceptance tests beyond the three named ones failing on 8.10.4 with the same
   error? A CI sweep against 8.10.4 would confirm.
3. Does `TestAccResourceSloFromSDK` step 1 (external provider v0.13.1) hit the version
   constraint skip on 8.10.4, or does it run and fail? The 8.10.4 exclusion will resolve this
   either way.

## Affected files

| File | Change |
|------|--------|
| `internal/kibana/slo/constants.go` | Add `!=8.10.4` to `SLOKqlAccTestConstraints` and `SLOKqlFleetAccTestConstraints` |
| `internal/kibana/slo/acc_test.go` | Add `!=8.10.4` to inline constraints in `TestAccResourceSlo_36_char_slo_id` and `TestAccResourceSloFromSDK` |
