## 1. Client layer

- [x] 1.1 Add `ListFleetServerHosts(ctx, client, spaceID)` to `internal/clients/fleet/fleet_server_host.go`; non-200 responses are errors and a response with `total` greater than the number of items is an error
- [x] 1.2 Unit tests in `internal/clients/fleet/fleet_server_host_test.go`: full list, empty list, space-aware path, truncated list, 404 and 500

## 2. Data source package `internal/fleet/serverhostds`

- [x] 2.1 `schema.go`: `host_id`, `name` and `default` as `Optional+Computed` selectors with one `ExactlyOneOf`, `LengthAtLeast(1)` on the strings and `Equals(true)` on `default`; `space_id`; computed `id`, `hosts`, `proxy_id` and `ssl` without `key`
- [x] 2.2 `models.go`: model, `GetResourceID` placeholder when `host_id` is not the selector, `populateFromAPI`
- [x] 2.3 `read.go`: lookup by `host_id` (get), by `name` and by `default` (list and filter), not-found and ambiguous-match errors listing the matching `host_id`s
- [x] 2.4 `data_source.go`: wire schema and read into `entitycore.NewKibanaDataSource`
- [x] 2.5 Unit tests in `read_test.go` for each selector, the error cases, the `ssl` mapping and the schema

## 3. Registration and docs

- [x] 3.1 Register `serverhostds.NewDataSource` in `provider/plugin_framework.go`
- [x] 3.2 Add `examples/data-sources/elasticstack_fleet_server_host/data-source.tf` and put it on the `TestAccExamples_planOnly` skip list
- [x] 3.3 Generate `docs/data-sources/fleet_server_host.md`

## 4. Acceptance tests (`internal/fleet/serverhostds/acc_test.go`, except 4.2)

- [x] 4.1 Lookup by `host_id`, then an update of the host and a lookup by `name` on the same data source
- [x] 4.2 Lookup by `default = true`, and `default` reported as `true` on a lookup by `host_id`. This test is `TestAccDataSourceFleetServerHost_default` in `internal/fleet/serverhost/acc_test.go`: it changes the stack-wide default host, so it has to run in the same package as the resource test that does the same, where non-parallel tests are serialized
- [x] 4.3 `proxy_id` (8.7.1)
- [x] 4.4 `ssl` with all exposed fields and with `certificate` only; `ssl.key` is never in state (9.1.0)
- [x] 4.5 `space_id` omitted, set to the host's space and set to another space all return the host (9.1.0)
- [x] 4.6 Two hosts with the same name: error listing both `host_id`s
- [x] 4.7 Not found by `host_id` and by `name`
- [x] 4.8 Plan-time validation: no selector, two selectors, `default = false`, empty `name`

## 5. Validation

- [x] 5.1 Run unit tests, the new acceptance tests and golangci-lint
- [x] 5.2 Run `openspec validate --all`
