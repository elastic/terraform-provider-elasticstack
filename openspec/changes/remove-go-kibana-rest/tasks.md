## 1. Remove legacy Kibana config wiring

- [x] 1.1 Delete `Client.Kibana` and the legacy `kibanaConfig` type from `internal/clients/config`, leaving `KibanaOapi` as the only Kibana connection surface built by env, SDK, and Framework config paths.
- [x] 1.2 Update any config-related tests or helper assertions to use `kibanaoapi.Config` field names and semantics instead of legacy `kibana.Config` names.

## 2. Re-anchor factory and consumers on OpenAPI config

- [x] 2.1 Update `internal/clients/provider_client_factory.go` so Kibana scoped client validation and construction rely on `cfg.KibanaOapi` rather than `cfg.Kibana`.
- [x] 2.2 Update remaining first-party consumers such as `provider/provider_test.go` to read Kibana connection details from the OpenAPI config surface only.

## 3. Remove the final legacy import path

- [x] 3.1 Refactor `internal/kibana/synthetics/parameter/read.go` to replace the legacy `kbapi.APIError` assertion with response-based 404 handling using the generated Kibana client response object.
- [x] 3.2 Confirm repository-wide that no first-party Go source still imports `github.com/disaster37/go-kibana-rest/v8` or its `kbapi` subpackage.

## 4. Clean up module and repository artifacts

- [x] 4.1 Remove the `go-kibana-rest` `require` and `replace` directives from `go.mod`, delete `libs/go-kibana-rest`, and run `go mod tidy`.
- [x] 4.2 Update contributor docs and OpenSpec text that still describe `go-kibana-rest` removal as pending follow-up work.

## 5. Verify and finalize

- [x] 5.1 Run `make build` and targeted tests for the touched areas, or `go test ./...` if targeted coverage is insufficient, and fix any regressions caused by the cleanup.
- [x] 5.2 Run repository search for `disaster37/go-kibana-rest` and OpenSpec validation for this change so the tree and change artifacts reflect the final removed state.
