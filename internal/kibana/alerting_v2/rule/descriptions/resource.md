Creates and manages Kibana [Alerting v2 rules](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-alerting-v2): ES|QL rules that create alerts (`kind = "alert"`) or only store matching events (`kind = "signal"`).

**Prerequisites.** The target Elastic Stack must be version 9.6.0 or later and the Alerting v2 engine must be enabled in the target space through the `alerting:v2:enabled` Kibana advanced setting. Requests against a space where the engine is disabled fail with a diagnostic naming that setting.

**Update semantics.** Create and update both issue a single `PUT` whose body is built entirely from the configuration, so every apply is a full replacement of the rule's configuration (last write wins): attributes removed from the configuration are removed server-side, and edits made in the Kibana UI between a refresh and an apply are overwritten. `kind` is immutable server-side and changing it replaces the resource.

**Managed scope.** Terraform-managed rules are ES|QL rules. The resource does not expose `metadata.builder` (the Kibana UI form that authored a rule), the rule actors, or runtime state.
