# Design: Surface Kibana error messages for saved objects import failures

## Background

The Kibana Saved Objects Import API uses the [Boom](https://hapi.dev/family/boom/) error format for
non-200 responses: `{"statusCode": N, "error": "<HTTP reason>", "message": "<detail>"}`. When
`ImportSavedObjects` falls through to the `default:` branch for any non-200, non-400 status, it
calls `diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)`, which uses the raw JSON
bytes as the diagnostic detail. The result is an opaque JSON blob shown to the user instead of the
human-readable `message` field.

The 400 case already parses the JSON specially (via `resp.JSON400`); this fix brings equivalent
behaviour to non-typed status codes by parsing the Boom envelope at the `diagutil` layer.

## Architecture

### New helper: `diagutil.ReportKibanaBoomHTTPError`

Location: `internal/diagutil/http.go`

```go
type kibanaBoomError struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}

// ReportKibanaBoomHTTPError attempts to parse body as a Kibana Boom error.
// If body contains a non-empty "message" field, that field is used as the
// diagnostic detail under the caller-supplied summary. Otherwise it falls
// back to ReportUnknownHTTPError.
func ReportKibanaBoomHTTPError(statusCode int, summary string, body []byte) fwdiag.Diagnostics {
    var boom kibanaBoomError
    if err := json.Unmarshal(body, &boom); err == nil && boom.Message != "" {
        return fwdiag.Diagnostics{
            fwdiag.NewErrorDiagnostic(summary, boom.Message),
        }
    }
    return ReportUnknownHTTPError(statusCode, body)
}
```

The `summary` parameter lets each call site provide context-specific wording.

### Call-site change in `ImportSavedObjects`

Location: `internal/clients/kibanaoapi/saved_objects_import.go`

The `default:` branch changes from:

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

No schema changes. Diagnostic detail changes from raw JSON to the extracted `message` field when
the Boom envelope is well-formed.

## Testing strategy

### Unit tests for `diagutil.ReportKibanaBoomHTTPError` (`internal/diagutil/http_test.go`)

| Test case | Input | Expected |
|---|---|---|
| Valid Boom body | `{"error":"Unprocessable Entity","message":"Doc belongs to newer Kibana"}` | Diagnostics with the supplied summary and extracted message as detail |
| Fallback — invalid JSON | `not json` | Falls back to `ReportUnknownHTTPError` behaviour (summary is generic HTTP status) |
| Fallback — empty message | `{"error":"Unprocessable Entity","message":""}` | Falls back to `ReportUnknownHTTPError` |

### New integration test (`internal/clients/kibanaoapi/saved_objects_import_test.go`)

`TestImportSavedObjects_422Response`: mock server returns HTTP 422 with a Boom body. Assert that:
- `diags.HasError()` is true
- one diagnostic detail exactly equals the expected Boom `message` value

## Open questions

- **Does `compatibility_mode=true` prevent the 422 by converting to a 200-with-errors path in all cases?** If so, the fix covers the remaining cases where `compatibility_mode` is not set; no change to scope.
- **Should `importError.String()` expose more sub-fields from per-object errors (`missing_references`, `conflict` details)?** That is a separate UX improvement for the 200-with-errors case, outside this bug's scope.
- **Should the 400 branch also be refactored to use `ReportKibanaBoomHTTPError` for consistency?** The 400 case is already specially-cased via `resp.JSON400`; alignment is optional and deferred.

## Assumptions

- The Kibana Boom format is stable across the versions the provider supports. The fallback to
  `ReportUnknownHTTPError` handles any deviation gracefully.
- Updating other `kibanaoapi` callers is deferred to a follow-up change; the minimal required
  change set is bounded to the four files listed in `tasks.md`.
