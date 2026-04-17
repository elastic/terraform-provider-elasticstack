## Why

The provider still depends on `github.com/disaster37/go-kibana-rest/v8` even after the legacy Kibana status wiring was removed. The remaining dependency edges are now narrow and mechanical, so this is the right point to fully retire the deprecated module, its vendored fork, and the residual follow-up language that assumes the cleanup is still pending.

## What Changes

- Remove the last first-party `go-kibana-rest` imports from `internal/clients/config` by deleting the legacy `Client.Kibana` surface and making `KibanaOapi` the only Kibana connection shape used by provider wiring.
- Update the provider client factory and test helpers to treat the OpenAPI Kibana config as canonical, including scoped `kibana_connection` validation and acceptance-test variable builders.
- Replace the remaining `github.com/disaster37/go-kibana-rest/v8/kbapi` error assertion in synthetics parameter read with generated-client response handling that preserves existing 404 state-removal behavior.
- Remove `github.com/disaster37/go-kibana-rest/v8` from the root module, delete the `replace` directive and vendored `libs/go-kibana-rest` fork, and tidy the module graph.
- Update docs and OpenSpec requirements that currently describe `go-kibana-rest` removal as residual follow-up work so they reflect the final post-cleanup state.

## Capabilities

### New Capabilities

- `provider-go-module-kibana-clients`: Defines the requirement that the provider module and first-party source tree exclude the deprecated `go-kibana-rest` module and vendored fork once Kibana wiring has fully migrated to `generated/kbapi` and `internal/clients/kibanaoapi`.

### Modified Capabilities

- `provider-kibana-connection`: Tighten the scoped Kibana connection requirements so they no longer allow residual `go-kibana-rest` ownership in config wiring or synthetics read paths after the cleanup completes.
- `provider-client-factory`: Tighten the Kibana scoped client contract so the factory surface is fully OpenAPI-based and no longer carries any legacy Kibana client configuration contract.

## Impact

- **Affected code:** `internal/clients/config/`, `internal/clients/provider_client_factory.go`, `internal/kibana/synthetics/parameter/read.go`, `provider/provider_test.go`, `go.mod`, `go.sum`, and `libs/go-kibana-rest/`.
- **Affected docs/specs:** `dev-docs/high-level/generated-clients.md`, `dev-docs/high-level/coding-standards.md`, `openspec/specs/provider-kibana-connection/spec.md`, `openspec/specs/provider-client-factory/spec.md`, and the change-local delta specs for this cleanup.
- **Verification:** `make build`, targeted tests for config/synthetics behavior or `go test ./...`, repository search for `disaster37/go-kibana-rest`, and OpenSpec validation after the deltas are written.
