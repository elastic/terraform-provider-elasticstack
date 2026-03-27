## 1. Spec

- [x] 1.1 Keep delta spec aligned with proposal.md / design.md
- [ ] 1.2 On completion, sync delta into canonical spec or archive

## 2. Implementation

- [x] 2.1 Add `time_slider_control_config` schema block to `schema.go` with optional float32 percentage attributes and optional bool `is_anchored`
- [x] 2.2 Add float32 range validators (0.0 ≤ value ≤ 1.0) for `start_percentage_of_time_range` and `end_percentage_of_time_range`
- [x] 2.3 Extend `panelModel` struct in `models_panels.go` to carry the new config block and route it through the panel dispatcher
- [x] 2.4 Create `models_time_slider_control_panel.go` with converter implementing null-preservation read-back semantics
- [x] 2.5 Add schema validation enforcing that `time_slider_control_config` is only present on `type = "time_slider_control"` panels and conflicts with all other typed config blocks
- [x] 2.6 Update resource descriptions/documentation
- [x] 2.7 Extend `panelConfigValidator` description for `time_slider_control` + `config_json` (schema allowlist remains the sole diagnostic; no duplicate object-level error)

## 3. Testing

- [x] 3.1 Add acceptance tests for `time_slider_control` panel lifecycle (create with config, create without config, update, import), including non-dyadic percentages (e.g. `0.1` / `0.9`) and a plan-only step with `ExpectNonEmptyPlan: false` after apply to guard refresh drift
- [x] 3.2 Add unit tests for the converter, including exact float32 round-trip for non-dyadic values, null-preservation read-back, and percentage boundary validation
- [x] 3.3 Add negative acceptance and unit coverage rejecting practitioner-authored `config_json` for `time_slider_control`
