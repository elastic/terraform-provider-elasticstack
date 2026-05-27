# Proposal: Surface Kibana error messages for saved objects import failures

## Problem

When importing Kibana saved objects via `elasticstack_kibana_import_saved_objects`, the Kibana API
can return HTTP 422 Unprocessable Entity — for example, when the exported objects belong to a more
recent Kibana version than the target cluster. In that case Kibana includes a detailed,
human-readable message in the JSON response body (e.g.
`"Unprocessable Entity: Document \"...\" belongs to a more recent version of Kibana [10.3.0] when
the last known version is [7.9.3]."`), but the provider diagnostic shows only the raw JSON body
under the generic summary `"Unexpected status code from server: got HTTP 422"`.

## Proposed solution

Add a shared helper `diagutil.ReportKibanaBoomHTTPError` that parses the Kibana
[Boom](https://hapi.dev/family/boom/) error envelope (`{"statusCode":N,"error":"...","message":"..."}`).
When the envelope is valid and `message` is non-empty, it returns a named diagnostic with the
caller-supplied summary and the extracted message as the detail. When parsing fails or `message` is
absent, it falls back to the existing `ReportUnknownHTTPError` behavior.

`ImportSavedObjects` is updated to use this helper in its `default:` branch so that the
422 scenario surfaces the Kibana message instead of raw JSON.

## Scope

**In scope:**
- New `diagutil.ReportKibanaBoomHTTPError` function in `internal/diagutil/http.go`
- Unit tests covering the success path and the fallback path in `internal/diagutil/http_test.go`
- Updated `default:` branch in `internal/clients/kibanaoapi/saved_objects_import.go`
- New `TestImportSavedObjects_422Response` test in `internal/clients/kibanaoapi/saved_objects_import_test.go`
- Delta spec update to REQ-002 in `openspec/changes/fix-import-saved-objects-422-error-surfacing/specs/kibana-import-saved-objects/spec.md`

**Out of scope:**
- Updating other `kibanaoapi` callers (`source_map.go`, `responses.go`, `synthetics_monitor.go`) — recommended as follow-up
- Schema changes to `elasticstack_kibana_import_saved_objects`
- Sub-field capture from per-object `error` maps (separate UX improvement)
- Support for `POST /api/saved_objects/_resolve_import_errors`

## Related issue

Closes #795
