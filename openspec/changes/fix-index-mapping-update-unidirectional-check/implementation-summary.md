---
schema: gc.build.implementation-summary.v1
workflow:
  id: ma-5x7q
  formula: openspec-implementation
methodology:
  pack: openspec
  name: openspec-implementation
producer:
  formula: openspec-implementation
  stage: implement
  attempt: 1
status: complete
trace:
  upstream:
    - path: openspec/changes/fix-index-mapping-update-unidirectional-check/proposal.md
      hash: sha256:39c17ce77fa96a761e67907fce765094bebebb95d298d611c8b1bf78ed305ea5
    - path: openspec/changes/fix-index-mapping-update-unidirectional-check/design.md
      hash: sha256:d89705fc911a8299570690506695c5f57af7d3b84d55f0e31e0b756da2c13d19
    - path: openspec/changes/fix-index-mapping-update-unidirectional-check/tasks.md
      hash: sha256:e562f61eb9263c8e730611c539f089ec488a6c83448ffb2588256d64c15d758c
    - path: openspec/changes/fix-index-mapping-update-unidirectional-check/specs/elasticsearch-index/spec.md
      hash: sha256:ee276d84465b8160a94e204645b0982e0d4fc8731849e347b789f1a9f424d536
    - path: beads/ma-w001
      hash: bead:ma-w001
  coverage:
    - id: REQ-015
      status: covered
    - id: REQ-016
      status: covered
    - id: REQ-017
      status: covered
    - id: REQ-018
      status: covered
---

## Summary

Workflow root `ma-5x7q` (formula: `openspec-implementation`) fixed the silent
mapping-update bug (`elastic/protections-cloud#19769`) in the
`elasticstack_elasticsearch_index` resource.

The change is on branch `openspec/fix-index-mapping-update-unidirectional-check`
and was committed as `e99e11c5f fix(elasticsearch-index): use unidirectional
mapping update-decision check`.

All 13 tasks across four slices were completed: helper + unit tests (slice A),
wire + acceptance tests (slice B), spec sync (slice C), and build validation
(slice D). Acceptance tests `TestAccResourceIndexMappingsUpdateRegression` and
`TestAccResourceIndexUseExistingAdoptMappings` passed against a live Elastic
Stack 9.4.0.

## Intended Behavior

`updateMappings` in `internal/elasticsearch/index/index/update.go` now calls
`planMappings.RequiresMappingsUpdate(ctx, stateMappings)` instead of
`planMappings.StringSemanticEquals(ctx, stateMappings)`.

`RequiresMappingsUpdate` (added to `internal/elasticsearch/index/mappings_value.go`)
makes a **unidirectional** decision: it returns `true` (PUT needed) when the plan
contains any field not already present in state, and `false` when state is already
a non-drifting superset of plan. This preserves the existing tolerance for
Elasticsearch/template-injected extras while fixing the case where a user adds a
new field — previously treated as "equal" and silently skipped.

`create.go`'s `adoptExistingIndexOnCreate` calls the same `updateMappings` helper,
so the adoption path also picks up the fix with no additional change.

## Changed Files

| File | Change |
| --- | --- |
| `internal/elasticsearch/index/mappings_value.go` | Added `RequiresMappingsUpdate` method with unidirectional semantics |
| `internal/elasticsearch/index/mappings_value_test.go` | Added `TestMappingsValue_RequiresMappingsUpdate` unit tests (6 cases) |
| `internal/elasticsearch/index/index/update.go` | Replaced `StringSemanticEquals` with `RequiresMappingsUpdate`, inverted branch |
| `internal/elasticsearch/index/index/acc_test.go` | Added `TestAccResourceIndexMappingsUpdateRegression` and `TestAccResourceIndexUseExistingAdoptMappings` |
| `internal/elasticsearch/index/index/testdata/TestAccResourceIndexMappingsUpdateRegression/` | New test data: create + update configs |
| `internal/elasticsearch/index/index/testdata/TestAccResourceIndexUseExistingAdoptMappings/` | New test data: create config with `use_existing` adoption |
| `openspec/changes/fix-index-mapping-update-unidirectional-check/specs/elasticsearch-index/spec.md` | Updated REQ-015–REQ-018 and adoption requirement with unidirectional scenarios |

## Verification

Unit tests (`go test ./internal/elasticsearch/index/...`): all pass.

Acceptance tests run against Elastic Stack 9.4.0 (local Docker):

```
--- PASS: TestAccResourceIndexUseExistingAdoptMappings (3.74s)
--- PASS: TestAccResourceIndexMappingsUpdateRegression (5.82s)
PASS
ok  github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/index  6.733s
```

`go build ./...` and `go vet ./...`: no errors.

`openspec validate fix-index-mapping-update-unidirectional-check --type change`: all checks pass.

## Remaining Risks

None. The change is narrowly scoped: a single method addition and one call-site
swap. `StringSemanticEquals` and `MappingsSemanticallyEqual` are untouched.
