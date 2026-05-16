## Context

The four control panel schemas in `generated/kbapi/dashboards.json` (`kbn-controls-schemas-controls-group-schema-{esql,options-list,range-slider,time-slider}-control`) each define panel-level `width` (string enum, default `medium`) and `grow` (boolean, default `false`). The Terraform resource currently exposes the inner `config` body for each but not these two outer attributes. A targeted audit during implementation will identify any further gaps in the inner `config` schemas.

## Goals / Non-Goals

**Goals:**
- Reach parity with the API for the panel-level `width` and `grow` controls.
- Use the same null-preservation pattern already applied across REQ-026 / REQ-027 / REQ-028 / REQ-029 so server-side defaults do not force users to manage them in HCL.
- Surface and close any other narrow control gaps found during implementation.

**Non-Goals:**
- Reshape the control schemas. Existing nested attributes stay as-is.
- Add new control kinds. None present in the spec.

## Decisions

- **Validation**: validate `width` against the API enum (`small`/`medium`/`large`) at plan time.
- **Default handling**: do not bake the API defaults into the TF schema. Keep optional, null-by-default; treat absence as "let Kibana decide". This matches how booleans like `use_global_filters` are currently treated.
- **Single change for all four controls**: shared design, shared validators, one cohesive review. The control schemas are similar enough that splitting per control fragments the audit and creates drift risk.
- **Audit deliverable**: implementation must produce a short audit note (in the change folder or PR description) listing each control schema's API attributes and the corresponding TF coverage, with any newly added attributes called out.

## Risks / Trade-offs

- [Risk] Audit reveals more substantial gaps than expected → Mitigation: expand the change in scope or split out a follow-up; the pinned-panels change is decoupled and won't be blocked.
- [Risk] Default-vs-unset semantics differ subtly between the four controls → Mitigation: explicit per-control unit tests using the same null-preservation pattern as existing REQ-027 import scenario.
