## Context

The `elasticstack_fleet_output` resource has an `ssl` block that currently exposes three attributes: `certificate_authorities`, `certificate`, and `key`. The Fleet API's SSL type (`KibanaHTTPAPIsOutputSsl`) also carries a `verification_mode` field (enum: `certificate`, `full`, `none`, `strict`), but this is not plumbed through the provider's schema, intermediate model structs, or read/write mapping functions.

Users who need to configure `verification_mode` are forced to use `config_yaml`, which is sensitive, opaque to Terraform diff output, and not validated by the schema.

The generated API client already has the field; no client regeneration is needed.

## Goals / Non-Goals

**Goals:**

- Add `verification_mode` as an optional `ssl` block attribute across all four output types (Elasticsearch, Logstash, Kafka, Remote Elasticsearch)
- Correctly write `verification_mode` to the API on create and update
- Correctly read `verification_mode` back from the API into state on read

**Non-Goals:**

- Changes to any other SSL attributes
- Changes to Kafka-specific SSL handling (the Kafka `ssl` block is the same shared `ssl` top-level attribute)
- Data source support — the data source (`fleet-output-ds`) doesn't expose an `ssl` block, so no changes are needed there
- Changes to the generated `kbapi` client

## Decisions

### Decision: Extend `sslToObjectValue` signature vs. pass the full SSL struct

**Options:**

1. Add a `verificationMode *string` parameter to the existing `sslToObjectValue(ctx, cert, certAuths, key)` function.
2. Replace the function with one that accepts the full `*kbapi.KibanaHTTPAPIsOutputSsl` struct directly.

**Decision: Option 1 — add the parameter.**

Rationale: Option 1 is the smallest delta. The function already has individual field params, so adding one more is consistent. Option 2 would be cleaner long-term but introduces unnecessary coupling between the model layer and the generated client type, and it's a larger refactor with no other immediate benefit.

### Decision: Where to enforce valid values

The valid `verification_mode` values (`certificate`, `full`, `none`, `strict`) are an enum in the API. Enforce them in the Terraform schema with a `stringvalidator.OneOf(...)` validator. This gives clear plan-time errors rather than API-time errors.

### Decision: No computed/default

`verification_mode` is optional with no default. When not set it is null in state. The API treats absence and `full` the same (full verification is the default), so there is no useful computed behavior to add.

## Risks / Trade-offs

- **State upgrade risk**: The `ssl` block schema change adds a new optional attribute. Terraform Plugin Framework handles new optional attributes in existing state gracefully (they become null), so no state upgrade version bump is needed.
- **All four output-type fromAPI functions need updating**: Missing any one would cause `verification_mode` to silently drop on read for that output type. The tasks list covers all four explicitly.

## Open Questions

_(none)_
