## MODIFIED Requirements

### Requirement: Read and delete behavior (REQ-013–REQ-016)

Delete SHALL call the Delete Lifecycle API with the policy name portion of `id`. **Before invoking the Delete Lifecycle API, Delete SHALL identify any indices whose `index.lifecycle.name` setting references the policy name and remove that reference by setting `index.lifecycle.name` to `null`.**

The process for removing in-use references SHALL be:

1. Query `GET /_all/_settings/index.lifecycle.name?flat_settings=true` to obtain the `index.lifecycle.name` setting for every index.
2. Filter to indices whose setting value equals the policy name being deleted.
3. If one or more indices match, issue `PUT /{indices}/_settings` with `{"index.lifecycle.name": null}` where `{indices}` is a comma-separated list of matched index names.
4. After clearing references, proceed with `DELETE /_ilm/policy/{policy_name}`.

If the settings-clear call returns an error, Delete SHALL surface that error as a Terraform diagnostic and SHALL NOT proceed with the Delete Lifecycle API call. If the subsequent Delete Lifecycle API call returns an error (for example, because a new index referencing the policy was created during the clear step), Delete SHALL surface the Elasticsearch error verbatim.

#### Scenario: ILM policy deleted while referenced by backing index

- GIVEN an ILM policy named `"my-policy"` exists
- AND an index `".ds-logs-test-default-2026.01.01-000001"` has `index.lifecycle.name` set to `"my-policy"`
- WHEN Delete runs for the ILM policy resource
- THEN the provider SHALL first set `index.lifecycle.name` to `null` on `.ds-logs-test-default-2026.01.01-000001`
- AND then call `DELETE /_ilm/policy/my-policy`
- AND the resource SHALL be destroyed successfully

#### Scenario: No indices reference the policy

- GIVEN an ILM policy named `"unused-policy"` exists
- AND no index has `index.lifecycle.name` set to `"unused-policy"`
- WHEN Delete runs for the ILM policy resource
- THEN the provider SHALL skip the settings-clear step
- AND call `DELETE /_ilm/policy/unused-policy` directly

#### Scenario: Settings-clear fails before delete

- GIVEN an ILM policy named `"my-policy"` exists
- AND an index referencing the policy exists
- AND the `PUT /_settings` call to clear the reference fails (e.g., index is closed or unavailable)
- WHEN Delete runs
- THEN the provider SHALL surface the settings-clear error as a Terraform diagnostic
- AND SHALL NOT call `DELETE /_ilm/policy/my-policy`
