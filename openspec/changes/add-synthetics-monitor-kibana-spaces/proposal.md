## Why

Terraform users can create and manage Kibana Synthetics monitors, but cannot declare which additional Kibana spaces can view a monitor. This forces manual post-provisioning changes and prevents Terraform state from representing the intended visibility.

## What Changes

- Add an optional `kibana_spaces` configuration to declare the Kibana spaces in which a Synthetics monitor is visible, including the all-spaces wildcard.
- Keep `space_id` as the monitor's owning space; this change controls visibility only.
- Preserve a deliberately empty visibility configuration and reconcile Kibana's automatic inclusion of the owning space without creating a Terraform diff.

This change does not add Kibana-space lifecycle management, alter the lifecycle of `space_id`, or change monitor behavior beyond its space visibility.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-synthetics-monitor`: Synthetics monitor configuration and state include declared Kibana space visibility.

## Impact

- The `elasticstack_kibana_synthetics_monitor` resource gains a new optional configuration attribute.
- Monitor create and update payloads, read/state behavior, generated API client bindings, resource documentation, and resource tests are affected.
