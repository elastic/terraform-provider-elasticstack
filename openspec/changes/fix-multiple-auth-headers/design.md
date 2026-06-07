## Context

The provider sends duplicate `Authorization` headers when Elasticsearch and Kibana/Fleet are configured with different auth mechanisms. Two bugs cooperate to produce this behavior:

1. `buildKibanaOapiConfigFromFramework` (`kibana_oapi.go:59`) starts from `base.toKibanaOapiConfig()`, which copies all ES credentials — including `APIKey` — as the starting point for the Kibana config. When the Kibana provider block subsequently sets `Username`/`Password`, it does not clear `APIKey`. Both remain set simultaneously.

2. `transport.RoundTrip` (`kibanaoapi/client.go:113`) applies all set auth methods independently using a mix of `Header.Set` and `Header.Add`, so both methods reach the wire as separate `Authorization` headers.

The same inheritance path is taken by `NewFromEnv` (`env.go:44`), used in acceptance tests. The Fleet path is also affected: `newFleetConfigFromFramework` (`fleet.go:31`) starts from the already-resolved Kibana config and can inherit the same mixed-auth state.

The correct priority model confirmed by `@tobio` is `ENV > RESOURCE > PROVIDER` for all blocks. Priority is source-based, not method-based.

## Goals / Non-Goals

**Goals:**
- `Config` structs delivered to Kibana and Fleet clients carry exactly one auth mechanism.
- Source priority `ENV > RESOURCE > PROVIDER` is correctly implemented at every config-resolution layer.
- Partial env+schema auth combinations (e.g. `KIBANA_PASSWORD` in env + `username` from schema) are correctly preserved — this is a confirmed valid use case.
- Diagnostic warnings surface previously-silent auth precedence decisions.
- Transport switch provides defense-in-depth regardless of config state.
- Both Kibana and Fleet paths are covered.
- `NewFromEnv` (acceptance-test path) is fixed as a side effect.

**Non-Goals:**
- Extending `TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT` to auth fields (deferred follow-up).
- Full source-aware config rewrite (incrementally implemented instead).
- Elasticsearch-facing auth changes.
- SDK v2 concerns (confirmed fully removed).

## Decisions

### Decision 1: Method-scoped clearing at each priority boundary

When a higher-priority source introduces an auth method, clear fields from *conflicting* methods inherited from lower-priority sources. Same-method fields are preserved, allowing partial auth construction across sources (e.g. `KIBANA_PASSWORD` in env + `username` from provider schema both belong to BasicAuth — they should cooperate, not conflict).

**Kibana schema layer** (`buildKibanaOapiConfigFromFramework`):

```go
// Method-scoped clearing: if the Kibana provider block sets any auth field,
// clear fields from conflicting auth methods inherited from the ES base.
kibUsesBasicAuth := kibConfig.Username.ValueString() != "" || kibConfig.Password.ValueString() != ""
kibUsesAPIKey    := kibConfig.APIKey.ValueString() != ""
kibUsesBearer    := kibConfig.BearerToken.ValueString() != ""

if kibUsesBasicAuth {
    config.APIKey = ""
    config.BearerToken = ""
}
if kibUsesAPIKey {
    config.Username = ""
    config.Password = ""
    config.BearerToken = ""
}
if kibUsesBearer {
    config.Username = ""
    config.Password = ""
    config.APIKey = ""
}
// ... then apply the Kibana block's own values as before
```

**Kibana env layer** (`withNonURLEnvironmentOverrides`):

```go
_, hasUser   := os.LookupEnv("KIBANA_USERNAME")
_, hasPass   := os.LookupEnv("KIBANA_PASSWORD")
_, hasKey    := os.LookupEnv("KIBANA_API_KEY")
_, hasBearer := os.LookupEnv("KIBANA_BEARER_TOKEN")

if hasUser || hasPass { k.APIKey = ""; k.BearerToken = "" }
if hasKey             { k.Username = ""; k.Password = ""; k.BearerToken = "" }
if hasBearer          { k.Username = ""; k.Password = ""; k.APIKey = "" }

// ... then apply env values as before
```

**Fleet schema layer** (`newFleetConfigFromFramework`):

```go
fleetUsesBasicAuth := fleetCfg.Username.ValueString() != "" || fleetCfg.Password.ValueString() != ""
fleetUsesAPIKey    := fleetCfg.APIKey.ValueString() != ""
fleetUsesBearer    := fleetCfg.BearerToken.ValueString() != ""

if fleetUsesBasicAuth { config.APIKey = ""; config.BearerToken = "" }
if fleetUsesAPIKey    { config.Username = ""; config.Password = ""; config.BearerToken = "" }
if fleetUsesBearer    { config.Username = ""; config.Password = ""; config.APIKey = "" }

// ... then apply fleet block values as before
```

**Fleet env layer** (`withEnvironmentOverrides`):

```go
_, hasUser   := os.LookupEnv("FLEET_USERNAME")
_, hasPass   := os.LookupEnv("FLEET_PASSWORD")
_, hasKey    := os.LookupEnv("FLEET_API_KEY")
_, hasBearer := os.LookupEnv("FLEET_BEARER_TOKEN")

if hasUser || hasPass { c.APIKey = ""; c.BearerToken = "" }
if hasKey             { c.Username = ""; c.Password = ""; c.BearerToken = "" }
if hasBearer          { c.Username = ""; c.Password = ""; c.APIKey = "" }

// ... then apply env values as before
```

**Why this approach:** The alternative (all-or-nothing guard — clear all auth when any env var is set) would break the confirmed-valid case where `KIBANA_PASSWORD` is in env and `username` is in the provider schema. Method-scoped clearing correctly handles this.

### Decision 2: Transport switch as defense-in-depth

Change `transport.RoundTrip` from the current multi-branch `if` chain (mixing `Header.Set` and `Header.Add`) to a single `switch` statement using `Header.Set` throughout:

```go
switch {
case t.BearerToken != "":
    req.Header.Set("Authorization", "Bearer "+t.BearerToken)
case t.APIKey != "":
    req.Header.Set("Authorization", "ApiKey "+t.APIKey)
case t.Username != "":
    req.SetBasicAuth(t.Username, t.Password)
}
```

Priority order: `BearerToken > APIKey > BasicAuth`. This preserves the existing implicit "last wins" order (BearerToken already overwrote everything via `Set`; APIKey and BasicAuth were both applied previously). Both the Kibana client and the Fleet client (which uses `kibanaoapi.NewClientWithLabel`) benefit from this change.

### Decision 3: Diagnostic warnings

After the final config is assembled, emit a `diag.AddWarning` when more than one auth method group is set. This surfaces the previously-silent precedence decision.

```go
func authMethodCount(c kibanaOapiConfig) int {
    count := 0
    if c.Username != "" || c.Password != "" { count++ }
    if c.APIKey != "" { count++ }
    if c.BearerToken != "" { count++ }
    return count
}

if authMethodCount(config) > 1 {
    diags.AddWarning(
        "Multiple Kibana authentication methods configured",
        "More than one of username/password, api_key, or bearer_token is set "+
            "in the resolved Kibana configuration. Only one will be used. "+
            "Check your provider configuration and environment variables for "+
            "conflicting auth settings.",
    )
}
```

All relevant functions (`newProviderKibanaOapiConfigFromFramework`, `newKibanaOapiConfigFromFramework`, `newFleetConfigFromFramework`) already return `fwdiags.Diagnostics`, so no signature changes are required.

### Decision 4: `withFleetBlockFallback` is already correct

`withFleetBlockFallback` (`kibana_oapi.go:101`) uses "only fill if empty" guards (`if k.Username == ""`), so it cannot re-introduce a cleared auth method. No changes are needed there.

## Open questions

- **Same-level method conflict**: If two conflicting auth methods are set within the same priority level — e.g. both `KIBANA_API_KEY` and `KIBANA_USERNAME` set in env, or both `api_key` and `username` in the provider Kibana block — which wins? The transport switch implements `BearerToken > APIKey > BasicAuth` as the tie-breaker. Should this be the documented tie-breaker, or should it produce an error rather than silently dropping one method?
- **`TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT` for auth**: This flag currently only governs URL priority (`RESOURCE > ENV` for URL when set). Should its semantics be extended to auth fields in a follow-up, or should a separate flag cover this?
- **Fleet-level multi-auth warning**: `newFleetConfigFromFramework` already returns `fwdiags.Diagnostics`. Should it also emit a warning when Fleet config ends up with multiple auth methods, or is a warning at the Kibana layer sufficient since Fleet derives from Kibana?
