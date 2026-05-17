## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-detection-rule-empty-list-null --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-security-detection-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [ ] 2.1 Add `reconcileEmptyListsFromPlan(reference, target *Data)` to `internal/kibana/security_detection_rule/models.go`. The function SHALL copy each of the seven affected list attributes from `reference` into `target` when `reference.X` is a known, non-null, empty list (`!X.IsNull() && X.IsKnown() && len(X.Elements()) == 0`) and `target.X` is null. Affected attributes: `Actions`, `ExceptionsList`, `SeverityMapping`, `RiskScoreMapping`, `RelatedIntegrations`, `Threat`, `ThreatMapping`.
- [ ] 2.2 Call `reconcileEmptyListsFromPlan(&data, readData)` in `Create()` (`create.go`) immediately after the `readData, diags := r.read(...)` call and nil-check, before `resp.State.Set`.
- [ ] 2.3 Call `reconcileEmptyListsFromPlan(&data, readData)` in `Read()` (`read.go`) immediately after the `readData, diags := r.read(...)` call and nil-check, before `resp.State.Set`. Here `data` holds prior state from `req.State`.
- [ ] 2.4 Call `reconcileEmptyListsFromPlan(&data, readData)` in `Update()` (`update.go`) immediately after the `readData, diags := r.read(...)` call and nil-check, before `resp.State.Set`. Here `data` holds the plan from `req.Plan`.
- [ ] 2.5 Extend the reconciliation logic for nested `Threat` entries so that `technique` / `subtechnique` are copied from the reference data only when the reference path is a known, non-null empty list and the post-read value is null. Preserve `null` when the practitioner omitted the nested attribute or set it to `null`.

## 3. Testing

- [ ] 3.1 Add an acceptance test step (or test function) in `acc_test.go` that creates a detection rule with all seven attributes set to `[]` (`actions = []`, `exceptions_list = []`, `severity_mapping = []`, `risk_score_mapping = []`, `related_integrations = []`, `threat = []`, `threat_mapping = []`). The `terraform apply` MUST succeed without "Provider produced inconsistent result after apply" errors.
- [ ] 3.2 Assert in the same test that all seven attributes are stored as empty lists (`[]`) — not null — in Terraform state after `apply`.
- [ ] 3.3 Add a second test step that runs `terraform plan` after the initial apply and asserts no-diff (plan is empty), confirming there is no perpetual-diff regression.
- [ ] 3.4 Add unit tests for `reconcileEmptyListsFromPlan` in `models_test.go`: verify that (a) a null reference does not overwrite a null target, (b) a null target with an empty-list reference is updated, and (c) a non-empty target is not overwritten.
- [ ] 3.5 Add unit tests for nested `Threat` reconciliation: verify that explicitly configured empty `technique` / `subtechnique` lists round-trip as `[]`, while omitted / `null` nested lists remain `null`.
