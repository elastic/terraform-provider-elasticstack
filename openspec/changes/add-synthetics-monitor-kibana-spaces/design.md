## Context

See `proposal.md` for motivation. Kibana's monitor API accepts a `spaces` array while monitor routes remain scoped by a separate owning `space_id`. Kibana adds the owning space to a monitor's returned visibility set, so the API response is not always byte-for-byte identical to the requested value.

## Goals / Non-Goals

**Goals:**

- Represent monitor visibility as an optional Terraform list while retaining the distinction between omitted and explicitly empty input.
- Send the configured values through the generated Kibana monitor types for every monitor variant.
- Keep Terraform state stable when Kibana's only response-side change is automatic inclusion of the owning space.

**Non-Goals:**

- Manage Kibana space resources or permissions.
- Treat visibility changes as an ownership or resource replacement change.
- Infer or validate general Kibana space membership beyond values returned by Kibana.

## Decisions

### Preserve omitted and empty values separately

The generated API field and request mapper use a pointer to a string slice. A null Terraform collection maps to a nil pointer, while a known empty collection maps to a pointer to an empty slice. This lets creation and updates that retain an omitted value omit `spaces`, while a list-to-omitted update sends `spaces: []` to clear the previously managed additional visibility.

The alternative of converting both states to a plain slice was rejected because it would collapse those two API behaviors.

### Add the API field to shared monitor definitions and transform metadata

The Kibana OpenAPI transform defines `spaces` for the monitor response and shared request fields, then the generated client is refreshed. This makes the field available consistently to HTTP, TCP, ICMP, and browser request variants.

The alternative of defining local request-only wire structs was rejected because it would bypass the provider's generated Kibana client contract.

### Reconcile only Kibana's implicit owning-space addition

During response-to-state mapping, compare the configured visibility with the returned visibility after adding the owning `space_id` to the configured set if it is absent. If the two sets match, retain the configured list and its ordering. If no visibility was configured and Kibana returns only the owner, retain null. All other response differences update state from the API.

The alternative of always retaining configuration would hide external visibility changes. The alternative of always writing the API value would cause a perpetual diff whenever Kibana adds the owner.

## Risks / Trade-offs

- [Kibana returns spaces in an order different from the configuration] → compare sets only for the implicit-owner reconciliation case and retain configuration ordering when it matches.
- [An absent `spaces` field has a meaning distinct from an empty array] → cover both forms with request-mapping tests.
- [Generated client drift] → regenerate from the adjusted transform and verify the monitor package and generated-client tests.
