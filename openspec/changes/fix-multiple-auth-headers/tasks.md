## 1. Transport safety net

- [ ] 1.1 In `internal/clients/kibanaoapi/client.go`, replace the multi-branch `if` chain in `transport.RoundTrip` with a `switch` statement using `req.Header.Set` throughout, priority order: `BearerToken > APIKey > BasicAuth`
- [ ] 1.2 Verify that the Fleet client path is also covered (it uses `kibanaoapi.NewClientWithLabel`, so it shares the transport — no separate change needed)

## 2. Config layer — Kibana schema clearing

- [ ] 2.1 In `internal/clients/config/kibana_oapi.go`, add method-scoped auth clearing at the start of the Kibana-block application in `buildKibanaOapiConfigFromFramework`: before writing Kibana block auth fields, detect which auth method the Kibana block introduces and clear fields from conflicting methods that were inherited from the ES base config
- [ ] 2.2 Ensure same-method partial composition is preserved (e.g. Kibana block sets only `password` while `username` was inherited — only conflicting-method fields are cleared, not same-method fields from lower-priority sources)

## 3. Config layer — Kibana env clearing

- [ ] 3.1 In `internal/clients/config/kibana_oapi.go`, add method-scoped auth clearing in `withNonURLEnvironmentOverrides`: before applying env-var overrides, detect which auth method the env introduces (using `os.LookupEnv` to distinguish "not set" from "set to empty string") and clear fields from conflicting methods
- [ ] 3.2 Preserve same-method partial composition: `KIBANA_PASSWORD` in env must not clear `username` from the provider schema (they belong to the same BasicAuth method)

## 4. Config layer — Fleet schema clearing

- [ ] 4.1 In `internal/clients/config/fleet.go`, add method-scoped auth clearing at the start of the Fleet-block application in `newFleetConfigFromFramework`: detect which auth method the Fleet block introduces and clear conflicting fields inherited from the Kibana config
- [ ] 4.2 Preserve same-method partial composition

## 5. Config layer — Fleet env clearing

- [ ] 5.1 In `internal/clients/config/fleet.go`, add method-scoped auth clearing in `withEnvironmentOverrides`: detect which Fleet auth env var groups are set and clear fields from conflicting methods before applying env values

## 6. Diagnostic warnings

- [ ] 6.1 Add a helper function (e.g. `authMethodCount(c kibanaOapiConfig) int`) that returns the count of distinct populated auth method groups (BasicAuth, APIKey, BearerToken)
- [ ] 6.2 In `newProviderKibanaOapiConfigFromFramework` and `newKibanaOapiConfigFromFramework`, after final config assembly, emit `diags.AddWarning` when `authMethodCount > 1`
- [ ] 6.3 Add an equivalent helper and warning in `newFleetConfigFromFramework` for the fleet config (or reuse a shared helper if the structs permit)

## 7. Unit tests — Kibana path

- [ ] 7.1 In `internal/clients/config/kibana_oapi_test.go`, add test: ES APIKey + Kibana username/password → resolved config has username/password only, no APIKey
- [ ] 7.2 Add test: ES APIKey + Kibana APIKey → resolved config has Kibana APIKey, no username/password
- [ ] 7.3 Add test: ES APIKey + no Kibana auth block → resolved config inherits ES APIKey (unchanged behavior)
- [ ] 7.4 Add test: `KIBANA_PASSWORD` env + provider `username` → resolved config has both fields set (same method, partial composition preserved)
- [ ] 7.5 Add test: `KIBANA_API_KEY` env + provider `username`/`password` → resolved config has APIKey only, BasicAuth cleared
- [ ] 7.6 Add test: warning diagnostic is emitted when resolved config still carries multiple auth methods
- [ ] 7.7 Add test: no warning when exactly one auth method is set

## 8. Unit tests — Fleet path

- [ ] 8.1 In `internal/clients/config/fleet_test.go`, add test: Kibana BasicAuth + Fleet APIKey block → resolved fleet config has APIKey only
- [ ] 8.2 Add test: Kibana APIKey + `FLEET_PASSWORD` env + provider Fleet `username` → resolved fleet config has BasicAuth (username from provider, password from env), no APIKey
- [ ] 8.3 Add test: `FLEET_API_KEY` env + Fleet provider `username`/`password` → resolved fleet config has APIKey only
- [ ] 8.4 Add test: warning diagnostic emitted on fleet config with multiple auth methods after resolution

## 9. Regression check

- [ ] 9.1 Confirm existing tests in `kibana_oapi_test.go` and `fleet_test.go` still pass (common case: ES and Kibana share same credentials, no explicit Kibana auth block → inheritance still works)
- [ ] 9.2 Run `make build` to confirm the provider compiles without errors
