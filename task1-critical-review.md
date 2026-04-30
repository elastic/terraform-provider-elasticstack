WARNING (fixed)
- Evidence: `internal/entitycore/resource_envelope.go` originally implemented `Create`/`Update` as empty methods on the shared envelope. Because the type satisfies `resource.Resource`, any concrete resource that accidentally forgot to override one of those methods would still compile and then silently no-op at runtime.
- Risk: This is a sharp edge for the planned embedding pattern; a migration mistake would look like a successful apply with no state mutation instead of a clear provider error.
- Recommended fix: Make the envelope defaults fail loudly by appending error diagnostics, while still allowing concrete resources to override them through normal method promotion/precedence rules.
- Status: Fixed. `Create`/`Update` now return explicit diagnostics in `internal/entitycore/resource_envelope.go`, and `internal/entitycore/resource_envelope_test.go` now verifies both the defensive defaults and that concrete overrides still win.

No remaining CRITICAL/WARNING/SUGGESTION findings in the reviewed files.