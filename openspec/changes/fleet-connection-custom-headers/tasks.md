## 1. Fleet connection block gains a `headers` attribute

- [ ] 1.1 Add a `headers` map-of-string attribute to `GetFleetFWConnectionBlock()`
      (`internal/schema/connection.go`), matching the exact `Optional`/`Sensitive`/`types.StringType`
      contract already used for `kibana.headers`, reusing the existing `attrHeaders` constant and adding a
      Fleet-specific markdown description
- [ ] 1.2 Add `Headers types.Map` (`tfsdk:"headers"`) to `FleetConnection`
      (`internal/clients/config/provider.go`)
- [ ] 1.3 Regenerate/update provider documentation so the `fleet` block's generated docs describe the new
      `headers` attribute

## 2. Fleet headers are validated with the exact same reserved-name list as Kibana

- [ ] 2.1 Extract the reserved-header-name and case-insensitive-duplicate validation currently inline in
      `buildKibanaOapiConfigFromFramework` (`internal/clients/config/kibana_oapi.go`) into a shared
      helper parameterized by an entity label, without changing Kibana's existing validation behavior or
      diagnostics
- [ ] 2.2 Rename `reservedKibanaHeaderNames` to a name that reflects shared use across Kibana and Fleet
      (e.g. `reservedConnectionHeaderNames`), with no change to the set of reserved names
- [ ] 2.3 Confirm all existing Kibana header validation tests (reserved-name rejection, case-insensitive
      duplicate rejection, whitespace trimming) still pass unchanged against the extracted helper

## 3. Fleet decodes, validates, and sends its own configured headers

- [ ] 3.1 Decode `fleet.headers` in `newFleetConfigFromFramework` (`internal/clients/config/fleet.go`)
      using the same `typeutils.IsKnown` + `ElementsAs` pattern already used for `kibana.headers`
- [ ] 3.2 Validate decoded Fleet headers with the shared helper from task 2.1, using an entity label of
      "Fleet", and return the resulting diagnostics on error before building the rest of the Fleet config
- [ ] 3.3 Assign the validated, trimmed headers to `config.Headers` only when non-empty, after
      `kibanaCfg.toFleetConfig()` has already cleared any inherited Kibana headers, so `fleet.headers` is
      independent of `kibana.headers` in both directions
- [ ] 3.4 Confirm the existing Fleet config-build tests in `internal/clients/config/fleet_test.go` continue
      to assert `fleetCfg.Headers` is `nil` when `fleet.headers` is omitted, null, or an empty map

## 4. No automatic Kibana-to-Fleet header forwarding is introduced

- [ ] 4.1 Confirm no endpoint-comparison or same-host inference logic is added anywhere in the config
      build pipeline as part of this change (Approach B is explicitly out of scope per human direction)
- [ ] 4.2 Add/extend a config-build unit test asserting that `kibana.headers` values never appear in
      `fleetCfg.Headers` even when `kibana.endpoints` and `fleet.endpoint` resolve to the same host

## 5. Test coverage proves headers reach the wire and stay independent

- [ ] 5.1 Add unit tests to `internal/clients/config/fleet_test.go`: a configured `fleet.headers` value
      reaches `fleetCfg.Headers`; reserved-name rejection and case-insensitive duplicate rejection behave
      identically to the existing Kibana-header tests; `kibana.headers` and `fleet.headers` configured to
      different values stay independent in both directions
- [ ] 5.2 Add an acceptance test alongside `TestFleetConfiguration_excludesKibanaHeaders`
      (`provider/acc_test.go`) that points `fleet_endpoint` at a local recording HTTP server, configures
      `fleet.headers`, and asserts the configured header is actually observed on a recorded Fleet request
      (not just that `terraform apply` succeeded), reusing the existing env-isolation pattern
      (`PreferConfiguredKibanaEndpointEnvVar`, explicit `FLEET_ENDPOINT` handling)
- [ ] 5.3 Extend that acceptance test (or add a sibling step) to assert that a `kibana.headers` value
      configured in the same test run does not appear on the recorded Fleet request
- [ ] 5.4 Run the new acceptance coverage across the provider's standard supported-Stack-version test
      matrix; do not add a Stack-version gate to the `fleet.headers` attribute itself
