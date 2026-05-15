## Why

The Kibana SLO API now permits longer custom SLO identifiers than the provider schema currently accepts. The provider rejects otherwise valid configurations at plan time because `elasticstack_kibana_slo.slo_id` is limited to 36 characters.

## What Changes

- Increase the maximum allowed length of `elasticstack_kibana_slo.slo_id` from 36 to 48 characters.
- Update the resource description and validation messaging so the documented SLO ID constraints match provider behavior.
- Add or update acceptance coverage to prove a 48-character SLO ID is accepted.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `kibana-slo`: Allow user-supplied SLO IDs up to 48 characters while preserving the existing minimum length and allowed character set.

## Impact

Affected code is concentrated in `internal/kibana/slo/`, especially the resource schema and tests. This change affects Terraform validation for the Kibana SLO resource and aligns provider behavior with the current Kibana SLO API.