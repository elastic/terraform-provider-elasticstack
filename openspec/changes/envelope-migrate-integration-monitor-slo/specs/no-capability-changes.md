# Spec Changes

This change contains no new or modified capability requirements. It is a pure implementation refactoring: migrating three existing resources from `*entitycore.ResourceBase` to `*entitycore.KibanaResource[T]`. External behaviour, schema attributes, and API contracts are unchanged.

Existing canonical specs are authoritative and unmodified:
- `openspec/specs/fleet-integration/spec.md`
- `openspec/specs/kibana-synthetics-monitor/spec.md`
- `openspec/specs/kibana-slo/spec.md`
