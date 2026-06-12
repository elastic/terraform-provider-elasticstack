## 1. Transport safety net

- [x] 1.1 In `internal/clients/kibanaoapi/client.go`, replace the multi-branch `if` chain in `transport.RoundTrip` with a `switch` statement using `req.Header.Set` throughout, priority order: `BearerToken > APIKey > BasicAuth`
- [x] 1.2 Verify that the Fleet client path is also covered (it uses `kibanaoapi.NewClientWithLabel`, so it shares the transport — no separate change needed)
- [x] 1.3 Add `internal/clients/kibanaoapi/transport_test.go`: test cases for (a) APIKey + BasicAuth set → only ApiKey header; (b) BearerToken set alongside others → only Bearer header; (c) BasicAuth only → only Basic header; (d) no auth fields set → no Authorization header

## 2. Config layer — shared clearing and counting helpers

- [x] 2.1 Add a `clearConflictingAuth` helper in `internal/clients/config/` (e.g. `auth.go`) that encodes the 3-branch clearing matrix: given a method-detection result (BasicAuth / APIKey / BearerToken), clear the fields belonging to all other methods. Use this helper in all four clearing sites (Kibana schema, Kibana env, Fleet schema, Fleet env) to avoid duplicating the same 3-branch logic.
- [x] 2.2 Add an `authMethodCount` helper in the same file that counts distinct populated auth method groups. Count BasicAuth only when `Username != ""` (to match the transport: `Password` alone sends no header and should not be counted as an active method). Use this single helper for both the Kibana and Fleet warning checks.

## 3. Config layer — Kibana schema clearing

- [x] 3.1 In `internal/clients/config/kibana_oapi.go`, add method-scoped auth clearing at the start of the Kibana-block application in `buildKibanaOapiConfigFromFramework`: before writing Kibana block auth fields, detect which auth method the Kibana block introduces and invoke `clearConflictingAuth` to remove conflicting fields inherited from the ES base config
- [x] 3.2 Ensure same-method partial composition is preserved (e.g. Kibana block sets only `password` while `username` was inherited — only conflicting-method fields are cleared, not same-method fields from lower-priority sources)

## 4. Config layer — `withFleetBlockFallback` method-level auth guard

- [x] 4.1 In `internal/clients/config/kibana_oapi.go`, add a method-level auth guard to `withFleetBlockFallback`: before filling any auth field (`Username`, `Password`, `APIKey`, `BearerToken`) from the Fleet block, check whether any of those fields is already non-empty in the Kibana config; if so, skip all auth field filling from Fleet entirely
- [x] 4.2 Confirm that URL, CA certs, and TLS (`Insecure`) filling in `withFleetBlockFallback` is unchanged — these are not subject to the auth guard

## 5. Config layer — Kibana env clearing

- [x] 5.1 In `internal/clients/config/kibana_oapi.go`, add method-scoped auth clearing in `withNonURLEnvironmentOverrides`: before applying env-var overrides, detect which auth method the env introduces (using `os.LookupEnv` to distinguish "not set" from "set to empty string") and invoke `clearConflictingAuth` to remove conflicting fields
- [x] 5.2 Preserve same-method partial composition: `KIBANA_PASSWORD` in env must not clear `username` from the provider schema (they belong to the same BasicAuth method)

## 6. Config layer — Fleet schema clearing

- [x] 6.1 In `internal/clients/config/fleet.go`, add method-scoped auth clearing at the start of the Fleet-block application in `newFleetConfigFromFramework`: detect which auth method the Fleet block introduces and invoke `clearConflictingAuth` to remove conflicting fields inherited from the Kibana config
- [x] 6.2 Preserve same-method partial composition

## 7. Config layer — Fleet env clearing

- [x] 7.1 In `internal/clients/config/fleet.go`, add method-scoped auth clearing in `withEnvironmentOverrides`: detect which Fleet auth env var groups are set (using `os.LookupEnv`) and invoke `clearConflictingAuth` before applying env values
- [x] 7.2 Preserve same-method partial composition: `FLEET_PASSWORD` in env must not clear `username` from the Fleet provider block (same BasicAuth method)

## 8. Diagnostic warnings

- [x] 8.1 In `newProviderKibanaOapiConfigFromFramework` and `newKibanaOapiConfigFromFramework`, after final config assembly, emit `diags.AddWarning` when `authMethodCount > 1`. Warning title: "Multiple Kibana authentication methods configured". Body directs the user to check environment variables for conflicting auth settings.
- [x] 8.2 In `newFleetConfigFromFramework`, after final config assembly, emit `diags.AddWarning` when `authMethodCount > 1`. Warning title: "Multiple Fleet authentication methods configured". Body directs the user to check Fleet environment variables for conflicting auth settings.

## 9. Unit tests — Kibana path

- [x] 9.1 In `internal/clients/config/kibana_oapi_test.go`, add test: ES APIKey + Kibana username/password → resolved config has username/password only, no APIKey
- [x] 9.2 Add test: ES APIKey + Kibana APIKey → resolved config has Kibana APIKey, no username/password
- [x] 9.3 Add test: ES APIKey + no Kibana auth block → resolved config inherits ES APIKey (unchanged behavior)
- [x] 9.4 Add test: `KIBANA_PASSWORD` env + provider `username` → resolved config has both fields set (same method, partial composition preserved)
- [x] 9.5 Add test: `KIBANA_API_KEY` env + provider `username`/`password` → resolved config has APIKey only, BasicAuth cleared
- [x] 9.6 Add test: env-level conflict (`KIBANA_API_KEY` and `KIBANA_USERNAME` both set) → warning diagnostic emitted with title "Multiple Kibana authentication methods configured"
- [x] 9.7 Add test: exactly one auth method set → no warning emitted
- [x] 9.8 Add test: `KIBANA_BEARER_TOKEN` env + provider `username`/`password` → resolved config has BearerToken only, BasicAuth cleared
- [x] 9.9 Add test: Kibana block sets only `username` (single-field BasicAuth detection) → ES APIKey is cleared, partial BasicAuth is preserved
- [x] 9.10 **Update or remove** the existing test `"kibana username with fleet password uses both blocks field-by-field"` (`kibana_oapi_test.go:520`): this test asserts that `withFleetBlockFallback` fills `Password` from the Fleet block even when the Kibana block has already set `Username`. This behavior is being intentionally changed by task 4.1 — with the new method-level auth guard, `Password` from Fleet will NOT be filled once Kibana has `Username` set. The test must be updated to assert the new expected behavior: resolved config has `Username = "kibana-user"` and `Password = ""`.
- [x] 9.11 Add test: Kibana BasicAuth set + Fleet block has `api_key` → `withFleetBlockFallback` does NOT set `APIKey`
- [x] 9.12 Add test: Kibana `APIKey` set (inherited from ES) + Fleet block has `username`/`password` → `withFleetBlockFallback` does NOT fill `Username`/`Password`
- [x] 9.13 Add test: no Kibana auth set + Fleet block has `username`/`password` → `withFleetBlockFallback` fills both fields

## 10. Unit tests — Fleet path

- [x] 10.1 In `internal/clients/config/fleet_test.go`, add test: Kibana BasicAuth + Fleet APIKey block → resolved fleet config has APIKey only
- [x] 10.2 Add test: Kibana APIKey + `FLEET_PASSWORD` env + provider Fleet `username` → resolved fleet config has BasicAuth (username from provider, password from env), no APIKey
- [x] 10.3 Add test: `FLEET_API_KEY` env + Fleet provider `username`/`password` → resolved fleet config has APIKey only
- [x] 10.4 Add test: env-level conflict (`FLEET_API_KEY` and `FLEET_USERNAME` both set) → warning diagnostic emitted with title "Multiple Fleet authentication methods configured"
- [x] 10.5 Add test: Fleet inherits Kibana config that already has multiple auth methods set → warning emitted after Fleet env clearing
- [x] 10.6 Add test: `FLEET_BEARER_TOKEN` env + Fleet provider BasicAuth → resolved fleet config has BearerToken only

## 11. Regression check

- [ ] 11.1 Confirm surviving tests in `kibana_oapi_test.go` and `fleet_test.go` still pass (common case: ES and Kibana share same credentials, no explicit Kibana auth block → inheritance still works)
- [ ] 11.2 Run `make build` to confirm the provider compiles without errors
