# Tasks: Skip SLO acceptance tests on Kibana 8.10.4

## Task 1 — Update shared SLO KQL acceptance test constraints

**File:** `internal/kibana/slo/constants.go`

Update `SLOKqlAccTestConstraints` and `SLOKqlFleetAccTestConstraints` to include `!=8.10.4`.

Before:
```go
var SLOKqlAccTestConstraints = mustKqlAccConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
var SLOKqlFleetAccTestConstraints = mustKqlAccConstraint(">=8.10.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
```

After:
```go
var SLOKqlAccTestConstraints = mustKqlAccConstraint(">=8.9.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
var SLOKqlFleetAccTestConstraints = mustKqlAccConstraint(">=8.10.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
```

**Acceptance:** `go build ./internal/kibana/slo/...` succeeds; no test file references the
old constant values.

---

## Task 2 — Update inline constraint in TestAccResourceSlo_36_char_slo_id

**File:** `internal/kibana/slo/acc_test.go`

In `TestAccResourceSlo_36_char_slo_id`, update the inline `version.NewConstraint` call to
include `!=8.10.4`.

Before:
```go
slo36CharConstraints, err := version.NewConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4,<8.16.0")
```

After:
```go
slo36CharConstraints, err := version.NewConstraint(">=8.9.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4,<8.16.0")
```

**Acceptance:** `go vet ./internal/kibana/slo/...` succeeds.

---

## Task 3 — Update inline constraint in TestAccResourceSloFromSDK

**File:** `internal/kibana/slo/acc_test.go`

In `TestAccResourceSloFromSDK`, update the inline `version.NewConstraint` call to include
`!=8.10.4`.

Before:
```go
sloConstraints, err := version.NewConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
```

After:
```go
sloConstraints, err := version.NewConstraint(">=8.9.0,!=8.10.4,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
```

**Acceptance:** `go vet ./internal/kibana/slo/...` succeeds.

---

## Verification

After all three tasks:

1. `go build ./...` succeeds.
2. `go vet ./internal/kibana/slo/...` succeeds.
3. On a CI run targeting Kibana 8.10.4, the three previously-failing tests are now skipped
   rather than failing.
4. On a CI run targeting any other supported version (e.g. 8.12.x, 8.13.x, 8.14.x, 8.15.x),
   the same three tests still run and pass.
