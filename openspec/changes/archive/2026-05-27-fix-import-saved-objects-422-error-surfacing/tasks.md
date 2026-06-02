# Tasks: Surface Kibana error messages for saved objects import failures

## Status

- [x] TASK-001 — `ReportKibanaBoomHTTPError` added in `internal/diagutil/http.go`
- [x] TASK-002 — `TestReportKibanaBoomHTTPError` added in `internal/diagutil/http_test.go`
- [x] TASK-003 — `ImportSavedObjects` `default:` branch now uses the new helper
- [x] TASK-004 — `TestImportSavedObjects_422Response` added in `internal/clients/kibanaoapi/saved_objects_import_test.go`
- [x] TASK-005 — Delta spec scenario added under REQ-002

## TASK-001: Add `ReportKibanaBoomHTTPError` to `diagutil`

**File**: `internal/diagutil/http.go`

Add a package-level unexported struct `kibanaBoomError` and an exported function
`ReportKibanaBoomHTTPError(statusCode int, summary string, body []byte) fwdiag.Diagnostics`.

The function must:
1. Attempt to `json.Unmarshal` `body` into `kibanaBoomError`.
2. When unmarshalling succeeds and `boom.Message` is non-empty, return a single
   `fwdiag.NewErrorDiagnostic(summary, boom.Message)`.
3. Otherwise call and return `ReportUnknownHTTPError(statusCode, body)`.

Import `"encoding/json"` (add to the existing import block).

## TASK-002: Add unit tests for `ReportKibanaBoomHTTPError`

**File**: `internal/diagutil/http_test.go`

Add a `TestReportKibanaBoomHTTPError` function (or table-driven `TestReportKibanaBoomHTTPError_*`
subtests) covering:

| Case | Body | Expected summary contains | Expected detail contains |
|---|---|---|---|
| Valid Boom body | `{"error":"Unprocessable Entity","message":"Doc belongs to newer Kibana [10.3.0]"}` | caller-supplied summary | `"Doc belongs to newer Kibana [10.3.0]"` |
| Invalid JSON | `not json` | `"Unexpected status code"` | raw body string |
| Empty message | `{"error":"Unprocessable Entity","message":""}` | `"Unexpected status code"` | raw body string |

Use `assert.Contains` (or `assert.Equal`) from `github.com/stretchr/testify/assert`.

## TASK-003: Update `ImportSavedObjects` to use the new helper

**File**: `internal/clients/kibanaoapi/saved_objects_import.go`

In the `switch resp.StatusCode()` block, change the `default:` branch from:

```go
default:
    return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
```

to:

```go
default:
    return nil, diagutil.ReportKibanaBoomHTTPError(
        resp.StatusCode(),
        "failed to import saved objects",
        resp.Body,
    )
```

No other changes to this file.

## TASK-004: Add `TestImportSavedObjects_422Response` unit test

**File**: `internal/clients/kibanaoapi/saved_objects_import_test.go`

Add a new test function `TestImportSavedObjects_422Response` following the pattern of
`TestImportSavedObjects_400Response`:

1. Start an `httptest.NewServer` that returns HTTP 422 with body:
   ```json
   {"statusCode":422,"error":"Unprocessable Entity","message":"Document \"abc\" belongs to a more recent version of Kibana [10.3.0] when the last known version is [7.9.3]."}
   ```
2. Construct a Kibana client via `clients.NewAcceptanceTestingKibanaScopedClient()`.
3. Call `kibanaoapi.ImportSavedObjects` with empty params and assert:
   - `diags.HasError()` is true
   - `result` is nil
   - At least one diagnostic detail exactly equals the expected Boom `message` string (not the raw
     JSON body)

## TASK-005: Update delta spec — REQ-002 scenario for 422 responses

**File**: `openspec/changes/fix-import-saved-objects-422-error-surfacing/specs/kibana-import-saved-objects/spec.md`

Add a new scenario under **REQ-002** describing the 422 Unprocessable Entity behaviour. See the
delta spec file in this change for the exact wording.
