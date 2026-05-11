## 1. Read-path preservation

- [ ] 1.1 Add an unexported `rawConfig` (or equivalently scoped private value) to `panelModel` for storing the raw API panel payload
- [ ] 1.2 In `mapPanelFromAPI` (`internal/kibana/dashboard/models_panels.go`), replace the destructive `default` branch with logic that preserves `id`, `grid`, `type`, and the raw API config JSON for unrecognized panel types
- [ ] 1.3 Ensure the preserved payload normalizes through the existing `config_json` semantic-equality path so refreshes do not produce diffs from key reordering or whitespace

## 2. Write-path round-trip

- [ ] 2.1 In the panel write dispatcher, when the panel `type` matches no typed block and the `panelModel` has a preserved raw payload, re-marshal that payload into the API request unchanged
- [ ] 2.2 When the panel `type` matches no typed block and no preserved payload exists, return the existing "unsupported panel type" error diagnostic with a message clarifying the type is not yet typed

## 3. Tests

- [ ] 3.1 Unit test in `models_panels_test.go` covering an unknown-type panel: read produces preserved fields, second write replays the payload, validation rejects user-authored unknown types
- [ ] 3.2 Acceptance test that creates a dashboard via the Kibana API containing a panel type the resource does not type, imports it into Terraform, and verifies a no-op plan
- [ ] 3.3 Run `make build`, `go vet ./...`, and `go test ./internal/kibana/dashboard/...` (with `TF_ACC=1` for the new acceptance test)

## 4. Spec sync

- [ ] 4.1 Run `make check-openspec` to validate the change
- [ ] 4.2 After implementation lands, ensure `openspec/specs/kibana-dashboard/spec.md` is updated by archive
