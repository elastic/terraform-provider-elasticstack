## 1. Audit

- [ ] 1.1 Compare each control schema in `generated/kbapi/dashboards.json` (`kbn-controls-schemas-controls-group-schema-{esql,options-list,range-slider,time-slider}-control`) against the corresponding `*_control_config` schema in `internal/kibana/dashboard/schema.go`
- [ ] 1.2 Capture the audit as a short note in the change folder (or PR description) listing each missing attribute and its proposed TF representation
- [ ] 1.3 If audit surfaces additional gaps beyond `width`/`grow`, extend `specs/kibana-dashboard/spec.md` accordingly before implementation

## 2. Schema additions

- [ ] 2.1 Add `width` (string, validator: `small`/`medium`/`large`) and `grow` (bool) to all four `*_control_config` schemas
- [ ] 2.2 Apply any additional attributes identified by the audit
- [ ] 2.3 Update descriptions under `internal/kibana/dashboard/descriptions/`

## 3. Models and mapping

- [ ] 3.1 Extend each control panel model with `Width` and `Grow` (and any audit additions)
- [ ] 3.2 Map to/from the API panel body in each control's read/write helpers
- [ ] 3.3 Apply REQ-009 null-preservation semantics on read

## 4. Tests

- [ ] 4.1 Per-control unit tests for `width`/`grow` round-trip and null-preservation
- [ ] 4.2 Validator unit test for invalid `width` enum value
- [ ] 4.3 At least one acceptance test per control type exercising `width` and `grow`
- [ ] 4.4 Run `make build`, `go vet ./...`, `go test ./internal/kibana/dashboard/...` (`TF_ACC=1` for acceptance)

## 5. Spec sync

- [ ] 5.1 Run `make check-openspec`
