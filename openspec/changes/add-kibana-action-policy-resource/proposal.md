# Proposal

## Why

Kibana Alerting v2 separates notification routing from rules: **action policies** (`/api/alerting/v2/action_policies`, stable since Kibana 9.6.0) decide which alerts are notified, how they are grouped and throttled, and which workflows receive them. The provider today only manages v1 alerting (`elasticstack_kibana_alerting_rule`), so teams adopting Alerting v2 cannot declare their notification routing as code, cannot review it in pull requests, and have no drift detection when someone edits a policy in the UI.

Source: [elastic/rna-program#674](https://github.com/elastic/rna-program/issues/674), part of the Alerting v2 Terraform epic [elastic/rna-program#655](https://github.com/elastic/rna-program/issues/655). The action policy surface is settled ahead of the v2 rule resource ([elastic/rna-program#673](https://github.com/elastic/rna-program/issues/673)) because the rule resource depends on how rules and policies are linked.

## What Changes

- Add a new Terraform resource, `elasticstack_kibana_action_policy`, that manages a single Alerting v2 action policy in a Kibana space: create, read, update, delete, and import.
- A practitioner can declare the policy's name, description, notification destinations, which alerts it applies to, how matching alerts are grouped, and how often notifications are sent. Workflows are the only destination type the API accepts.
- A practitioner can optionally choose the policy identifier, so the same policy identifier can be used across environments. The provider refuses to take over an existing policy with that identifier; existing policies are brought under management through import.
- A practitioner can enable or disable a policy from Terraform. Newly created policies are enabled unless configured otherwise.
- The resource documents the link between rules and policies. A rule does not point at a policy: a policy selects alerts by matching the rule's routing tags and/or a KQL condition on each alert, and a policy with no matching criteria applies to all alerts in its space. The v2 rule resource design builds on this.
- The resource documentation states that notifications run with the credentials of whoever last wrote the policy, which for Terraform is the provider's Kibana credentials. Every update re-binds the policy to those credentials.
- The resource is opt-in. It is only available when experimental resources are enabled (`TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`), and its documentation marks it as **technical preview**.
- The resource is limited to stacks that ship the stable action policy API (Kibana 9.6.0 or newer). Its documentation lists the remaining prerequisites: an Enterprise license and Alerting v2 enabled. Alerting v2 is not available in serverless Elasticsearch or Security projects.
- Add documentation, an example, and acceptance tests for the resource.

Non-goals:

- No data source for reading or listing action policies. It is deferred.
- No v2 rule resource. It is designed separately in elastic/rna-program#673.
- No management of snoozing. Snoozing is a temporary, operational action taken outside configuration, and the provider ignores it.
- No Terraform control over the policy's credentials beyond what every write already does. There is no separate key-rotation action.
- No bulk operations and no execution history.
- No graduation out of experimental; that is a separate, later decision.
- No support for Kibana versions before 9.6.0, whose experimental action policy API used a different, incompatible shape.

## Capabilities

### New Capabilities
- `kibana-action-policy`: Schema, identity and import, create/read/update/delete behavior, enabled-state management, rule-to-policy routing semantics, version and availability gating, error handling, and experimental registration for the `elasticstack_kibana_action_policy` resource.

### Modified Capabilities
<!-- None. This is a net-new resource; registering it in the experimental resource set does not change requirements of any existing capability. -->

## Impact

- The shared Kibana resource layer gains an opt-in option so a resource can keep a successfully written object in state when a follow-up step fails. It is off by default, and existing resources are unaffected.
- A new Kibana resource in the provider and a thin wrapper over the Kibana client for the action policy endpoints. The generated Kibana client already includes these endpoints, so no client regeneration is expected.
- The provider's experimental resource registration. The resource is invisible to users who do not opt in.
- User-facing documentation and an example under the provider's docs and examples.
- Acceptance tests require a Kibana 9.6.0+ stack with an Enterprise (or trial) license and Alerting v2 enabled. CI stacks that do not meet this skip the tests.
- The sibling design, elastic/rna-program#673 (v2 rule resource), must expose the rule's routing tags so that policies can target its alerts.
- The Alerting v2 epic (elastic/rna-program#655) blocker "minimum Kibana version" is resolved as 9.6.0 for action policies.
