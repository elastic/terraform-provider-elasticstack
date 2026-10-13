# Delta spec: `kibana-dashboard` — Lens XY duration-format version gate

Extends the main spec at `openspec/specs/kibana-dashboard/spec.md`.

## ADDED Requirements

### Requirement: Lens XY duration-format parameter version gate (REQ-056)

For every Lens XY chart `y[]` metric reachable at `panels[].vis_config.by_value.xy_chart_config`
(top-level `panels` and panels nested under `sections[].panels`), when that metric's
`config_json`-decoded `format` has `type` or `id` equal to `"duration"` and also sets `decimals`
and/or `compact` (directly on `format` or nested under `format.params`), the provider SHALL require
Elastic Stack **9.5.2** or later before issuing the create or update API request. When the connected
stack is older than 9.5.2 and is not Serverless, the provider SHALL return an error diagnostic that
identifies the offending panel, layer, and metric, and states that `format.type = "duration"` with
`decimals`/`compact` requires Elastic Stack 9.5.2 or later, instead of issuing the API request and
surfacing Kibana's raw validation error. The provider SHALL NOT perform this version check for
dashboards whose `y[]` metrics do not use this combination, and SHALL NOT change behavior for
`format.type` values other than `duration` (in particular `number` and `percent`, which already
accept `decimals`/`compact` on every supported Kibana version).

#### Scenario: Duration format with decimals and compact on an unsupported stack

- GIVEN a dashboard configuration with an XY chart `line` layer whose `y[]` metric has
  `config_json` containing `format = { type = "duration", from = "s", to = "ms", decimals = 0,
  compact = true }`
- AND the connected Elastic Stack is older than 9.5.2 and is not Serverless
- WHEN the provider creates or updates the dashboard
- THEN the provider SHALL return an error diagnostic naming the panel/layer/metric path and the
  minimum required version
- AND the provider SHALL NOT issue the underlying Kibana Dashboard API create/update request

#### Scenario: Duration format with decimals and compact on a supported stack

- GIVEN the same configuration as above
- AND the connected Elastic Stack is 9.5.2 or later
- WHEN the provider creates or updates the dashboard
- THEN the provider SHALL issue the API request unchanged and SHALL NOT return a version-gate error

#### Scenario: Duration format with decimals and compact on Serverless

- GIVEN the same configuration as above
- AND the connected stack is Kibana Serverless
- WHEN the provider creates or updates the dashboard
- THEN the provider SHALL NOT return a version-gate error, consistent with
  `EnforceMinVersion` always being satisfied on Serverless

#### Scenario: Duration format without decimals or compact

- GIVEN an XY chart `y[]` metric with `config_json` containing `format = { type = "duration", from
  = "s", to = "ms" }` and neither `decimals` nor `compact` set
- WHEN the provider creates or updates the dashboard on any supported Kibana version
- THEN the provider SHALL NOT perform the version check described in this requirement and SHALL NOT
  return a version-gate error

#### Scenario: Number or percent format with decimals and compact is unaffected

- GIVEN an XY chart `y[]` metric with `config_json` containing `format = { type = "number", decimals
  = 0, compact = true }`
- WHEN the provider creates or updates the dashboard on any supported Kibana version
- THEN the provider SHALL NOT perform the version check described in this requirement and SHALL NOT
  return a version-gate error

#### Scenario: Offending metric nested inside a section

- GIVEN the offending `y[]` metric from the first scenario is on a panel nested inside
  `sections[].panels` rather than a top-level panel
- AND the connected Elastic Stack is older than 9.5.2 and is not Serverless
- WHEN the provider creates or updates the dashboard
- THEN the provider SHALL return the same error diagnostic as the top-level-panel case, naming the
  section and panel path
