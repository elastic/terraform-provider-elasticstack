## 1. Practitioner can use wildcard and multi-target expressions in `read_indices`

- [ ] 1.1 Add an index-expression resolver (Resolve Index API with `expand_wildcards=all`, `allow_no_indices=true`) in `internal/clients/elasticsearch/` returning concrete indices and data streams, and rejecting alias and remote-cluster targets; verify with unit tests for wildcard, comma, exclusion, `_all`, no-match, alias-target and remote-target cases
- [ ] 1.2 Add the computed `concrete_indices` `set(string)` attribute to `read_indices` in `schema.go` and the model, updating attribute-type helpers; verify with a schema unit test and `make build`
- [ ] 1.3 Implement action computation: expand each `read_indices` element to per-target configs, deduplicate identical settings, error on conflicting settings or a target that is also the write index; verify with unit tests for overlapping-identical, overlapping-conflicting and write-index-collision cases
- [ ] 1.4 Rework create and update to use resolved targets and diff against the live Get Alias response per concrete name (no remove/add churn for attached targets, no Update Aliases call when no actions); verify with unit tests on the action builder
- [ ] 1.5 Rework read to store per-element `concrete_indices` as the intersection of resolved targets and actual alias members, and to surface uncovered members as singleton `read_indices` entries; verify with unit tests for full match, partial attachment, no match and unconfigured member
- [ ] 1.6 Rework delete to remove the alias from its live Get Alias members; verify with a unit test that no selector strings reach the API
- [ ] 1.7 Add acceptance test for wildcard create, read, update and delete that ends with a clean follow-up plan, reproducing the scenario from issue #5027 (alias on `traces-apm*`-style indices)

## 2. Practitioner sees membership drift in the plan

- [ ] 2.1 Implement `ModifyPlan` on the alias resource to resolve expressions and populate each element's planned `concrete_indices`, skipping resolution for unknown names; verify with unit tests including the unknown-name case
- [ ] 2.2 Add acceptance coverage for a newly matching but unattached target producing an in-place update, for an unconfigured alias member producing a removal plan, and for a target created after plan but before apply
- [ ] 2.3 Add acceptance coverage for no-match expressions, comma-separated, exclusion and `_all` expressions, and hidden and closed matching targets

## 3. Practitioner gets early errors for invalid `write_index` selectors and safe upgrades

- [ ] 3.1 Add `write_index.name` validation rejecting `*`, `?`, comma lists, leading `-` and `_all` without rejecting currently valid single-target names; verify with unit tests covering rejected syntax and representative accepted names (for example names with `.`, `-` in the middle, and `_`)
- [ ] 3.2 Decide and implement state compatibility for the new attribute (null-tolerant read or `StateUpgrader`); verify with an acceptance test that applies with the previous provider version and upgrades with a clean plan, and with an import test showing read after import followed by a wildcard-config plan
- [ ] 3.3 Update attribute descriptions, regenerate docs (`make docs-generate`), add a changelog entry, and verify `make check-docs` is clean

## 4. Final verification

- [ ] 4.1 Update `openspec/specs/elasticsearch-index-alias/spec.md` by syncing this change's delta, and verify `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate index-alias-wildcard-read-indices --type change` passes
- [ ] 4.2 Run `make build` and the alias package unit tests and verify both pass
