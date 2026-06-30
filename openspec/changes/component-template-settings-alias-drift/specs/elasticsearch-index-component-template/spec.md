## ADDED Requirements

### Requirement: Dotted-key settings produce no perpetual diff (REQ-037)

The resource SHALL treat `template.settings` values written with dotted Elasticsearch keys (e.g. `{"index.lifecycle.name":"my-policy"}`) as semantically equal to their nested form (`{"index":{"lifecycle":{"name":"my-policy"}}}`). When the plan value and the prior state value differ only in key representation (dotted vs nested) but are semantically equal, the resource SHALL rewrite the plan value to match the prior state's canonical encoding via a `ModifyPlan` hook so that Terraform displays no diff and the post-apply consistency check succeeds. This requirement applies on both the normal plan/apply cycle and after `terraform import`.

Implementation note: the fix mirrors the `ModifyPlan` hook already present in `elasticstack_elasticsearch_index_template` (`internal/elasticsearch/index/template/modify_plan.go`) and relies on `IndexSettingsValue.SemanticallyEqual` for the equality check.

#### Scenario: Dotted-key settings produce no diff after apply

- GIVEN a component template applied with `template.settings = jsonencode({"index.lifecycle.name":"my-policy"})` (dotted keys)
- WHEN a subsequent `terraform plan` runs (settings unchanged in config)
- THEN the plan SHALL show no changes to `template.settings`

#### Scenario: Dotted-key settings produce no diff after re-import

- GIVEN a component template created in Terraform with dotted-key `template.settings`
- WHEN `terraform import` is run followed by `terraform plan`
- THEN the plan SHALL show no changes to `template.settings` and `ImportStateVerify` SHALL pass

#### Scenario: Switching from dotted to nested form produces no diff

- GIVEN state holding nested-form `template.settings` (canonical Elasticsearch echo)
- WHEN the configuration is changed to the semantically-equivalent dotted-key form
- THEN `terraform plan` SHALL show no diff for `template.settings`

### Requirement: Alias `routing`-only configurations apply cleanly (REQ-038)

The resource SHALL accept a `template.alias` block that specifies only `routing` (without explicit `index_routing` or `search_routing`) and apply it without producing a `Provider produced inconsistent result after apply` error. Elasticsearch splits the `routing` field into `index_routing` and `search_routing` on the GET response; the provider SHALL reconcile this split on read and at plan time so that subsequent plans show no diff.

Implementation note: the fix mirrors the alias infrastructure already present in `elasticstack_elasticsearch_index_template`: adopting `aliasutil.AliasObjectType` as the custom element type for the alias set, adding read-time alias reconciliation, and extending the `ModifyPlan` hook to include the alias plan reconcilers from `aliasutil`.

#### Scenario: Routing-only alias applies without error

- GIVEN a component template with a single alias block that sets only `routing = "shard_1"` (no explicit `index_routing` or `search_routing`)
- WHEN the resource is created
- THEN the apply SHALL succeed without a `Provider produced inconsistent result after apply` error

#### Scenario: Routing-only alias produces no perpetual diff

- GIVEN a component template created with a `routing`-only alias
- WHEN a subsequent `terraform plan` runs (config unchanged)
- THEN the plan SHALL show no changes to `template.alias`
