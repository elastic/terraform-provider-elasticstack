# Terraform Provider Bug: Integration Policy Update Issue

**Date**: 2026-01-13
**Reporter**: James Garside
**Provider Version**: elastic/elasticstack v0.13.1
**Related PR**: [#1390 - Space-Aware Fleet](https://github.com/elastic/terraform-provider-elasticstack/pull/1390)
**Status**: Bug identified, fix needed

## Problem Summary

When updating `elasticstack_fleet_integration_policy` resources (e.g., changing `integration_version`), the Terraform provider produces inconsistent state errors, preventing in-place updates.

## Exact Error Messages

```
Error: Provider produced inconsistent result after apply

When applying changes to elasticstack_fleet_integration_policy.endpoint_hostnation[6], provider
"provider["registry.terraform.io/elastic/elasticstack"]" produced an unexpected new value:
.agent_policy_id: was cty.StringVal("61b8dd46-a55a-4221-8071-5d50824eb79c"), but now null.

This is a bug in the provider, which should be reported in the provider's own issue tracker.
```

```
Error: Provider produced inconsistent result after apply

When applying changes to elasticstack_fleet_integration_policy.endpoint_hostnation[6], provider
"provider["registry.terraform.io/elastic/elasticstack"]" produced an unexpected new value:
.input: block count changed from 0 to 1.

This is a bug in the provider, which should be reported in the provider's own issue tracker.
```

## Reproduction Case

### Scenario
DCM5 Blue Team deployment with 40 teams, each with 2 agent policies (HostNation + Deployed), each with 4 integrations (System, Endpoint, Windows, Auditd) = 320 integration policies.

### Trigger
Changing `auditd_integration_version` from `3.27.0` to `3.22.0` in variables.tf

### Configuration
```hcl
# variables.tf
variable "auditd_integration_version" {
  description = "Version of Auditd integration package"
  type        = string
  default     = "3.22.0"
}

# integrations.tf
resource "elasticstack_fleet_integration_policy" "auditd_hostnation" {
  count = var.team_count

  name            = "dcm5-auditd-${local.hostnation_namespaces[count.index]}"
  namespace       = local.hostnation_namespaces[count.index]
  agent_policy_id = elasticstack_fleet_agent_policy.hostnation[count.index].policy_id
  space_ids       = elasticstack_fleet_agent_policy.hostnation[count.index].space_ids

  integration_name    = "auditd"
  integration_version = var.auditd_integration_version  # Changed from 3.27.0 to 3.22.0
}
```

### Expected Behavior
Terraform should update the integration version in-place without recreating the resource.

### Actual Behavior
Provider loses `agent_policy_id` during update, causing inconsistent state error.

## Root Cause Analysis

### Hypothesis
PR #1390 introduced space-aware Fleet operations with a "request editor" pattern. The `Update()` method for `elasticstack_fleet_integration_policy` has a regression where:

1. **Read Phase**: Resource is read from space-aware API successfully with all fields
2. **Update Phase**: When building the update request, `agent_policy_id` is not preserved from state
3. **Response Handling**: API returns incomplete data or provider doesn't properly handle the response
4. **State Validation**: Terraform detects inconsistency between expected and actual `agent_policy_id`

### Evidence
- PR #1390 was merged November 4, 2025
- Bug appears in v0.13.1 (includes space-aware changes)
- Only affects **updates** (create and destroy work fine)
- Only affects resources with `space_ids` specified
- `agent_policy_id` is a computed field that should be preserved

### Affected Code Path

```
Terraform Update Flow:
┌─────────────────────────────────────────────────┐
│ 1. Read current state (space-aware API)        │
│    GET /s/{space}/api/fleet/package_policies   │
│    Response includes agent_policy_id            │
└─────────────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────┐
│ 2. Plan changes                                 │
│    Detect: integration_version changed          │
└─────────────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────┐
│ 3. Update resource ❌ BUG HERE                  │
│    PUT /s/{space}/api/fleet/package_policies    │
│    Request missing agent_policy_id              │
└─────────────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────┐
│ 4. Read updated state                           │
│    agent_policy_id = null ❌                    │
└─────────────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────┐
│ 5. State comparison                             │
│    Expected: "xxx", Actual: null                │
│    ERROR: Inconsistent state                    │
└─────────────────────────────────────────────────┘
```

## Proposed Fix

### Solution: Preserve Computed Fields During Update

**Location**: `internal/fleet/integration_policy/resource.go` (likely path)

**Approach**: Ensure computed fields like `agent_policy_id` are explicitly preserved when building update requests.

### Code Changes Required

#### 1. Update Method Enhancement

```go
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan, state models.IntegrationPolicy

    // Get plan and state
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // ✅ CRITICAL FIX: Preserve computed fields from state
    plan.ID = state.ID
    plan.AgentPolicyID = state.AgentPolicyID  // ❗ Key fix
    plan.PolicyID = state.PolicyID

    // Build update request
    updateReq := buildUpdateRequest(ctx, plan)

    // Execute space-aware update
    result, err := r.client.UpdateIntegrationPolicy(ctx, updateReq)
    if err != nil {
        resp.Diagnostics.AddError("Failed to update integration policy", err.Error())
        return
    }

    // ✅ Validate response completeness
    if result.AgentPolicyID == "" {
        resp.Diagnostics.AddError(
            "API returned incomplete data",
            fmt.Sprintf("agent_policy_id missing from response for policy %s", plan.ID.ValueString()),
        )
        return
    }

    // Set state
    resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}
```

#### 2. Request Builder Fix

```go
func buildUpdateRequest(ctx context.Context, plan models.IntegrationPolicy) *fleet.UpdatePackagePolicyRequest {
    tflog.Debug(ctx, "Building update request", map[string]any{
        "id":              plan.ID.ValueString(),
        "agent_policy_id": plan.AgentPolicyID.ValueString(),
    })

    return &fleet.UpdatePackagePolicyRequest{
        ID:            plan.ID.ValueString(),
        Name:          plan.Name.ValueString(),
        Namespace:     plan.Namespace.ValueString(),
        PolicyID:      plan.AgentPolicyID.ValueString(), // ❗ Ensure this is set
        Package: &fleet.Package{
            Name:    plan.IntegrationName.ValueString(),
            Version: plan.IntegrationVersion.ValueString(),
        },
        Inputs:   buildInputs(ctx, plan),
        SpaceIDs: extractSpaceIDs(plan.SpaceIDs),
    }
}
```

#### 3. API Client Validation

```go
func (c *Client) UpdateIntegrationPolicy(ctx context.Context, req *UpdatePackagePolicyRequest) (*PackagePolicy, error) {
    // Determine operational space (default-first pattern from PR #1390)
    operationalSpace := determineOperationalSpace(req.SpaceIDs)

    // Build space-aware URL
    url := fmt.Sprintf("/s/%s/api/fleet/package_policies/%s", operationalSpace, req.ID)

    // Log request for debugging
    log.Debug(ctx, "Updating integration policy", map[string]any{
        "url":       url,
        "policy_id": req.PolicyID,
        "name":      req.Name,
    })

    // Validate request has required fields
    if req.PolicyID == "" {
        return nil, fmt.Errorf("policy_id (agent_policy_id) is required for update")
    }

    // Execute PUT request
    var result PackagePolicy
    err := c.put(ctx, url, req, &result)
    if err != nil {
        return nil, err
    }

    // ✅ Validate response completeness
    if result.PolicyID == "" {
        return nil, fmt.Errorf("API returned incomplete response: missing policy_id")
    }

    return &result, nil
}
```

### Alternative Solution: RequiresReplace

If the above fix doesn't resolve the issue, mark `integration_version` as requiring replacement:

```go
"integration_version": schema.StringAttribute{
    MarkdownDescription: "Version of the integration package",
    Required:            true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(), // Force recreation on version change
    },
},
```

**Note**: This is less desirable as it causes service disruption.

## Implementation Plan

### Phase 1: Investigation (2 hours)

1. **Clone provider repository**
   ```bash
   git clone https://github.com/elastic/terraform-provider-elasticstack.git
   cd terraform-provider-elasticstack
   git checkout v0.13.1
   ```

2. **Locate the integration policy resource**
   ```bash
   # Find the main resource file
   find . -name "*integration_policy*.go" | grep -v test

   # Likely locations:
   # internal/fleet/integration_policy/resource.go
   # internal/providers/elasticstack/fleet/integration_policy.go
   ```

3. **Add debug logging**
   ```go
   // In Update() method
   tflog.Debug(ctx, "Update method called", map[string]any{
       "plan_id":              plan.ID.ValueString(),
       "plan_agent_policy_id": plan.AgentPolicyID.ValueString(),
       "state_agent_policy_id": state.AgentPolicyID.ValueString(),
   })

   // After API call
   tflog.Debug(ctx, "API response received", map[string]any{
       "response_policy_id": result.PolicyID,
   })
   ```

4. **Run test to reproduce**
   ```bash
   cd DCM5/prod
   TF_LOG=DEBUG terraform apply 2>&1 | tee debug.log
   ```

5. **Analyze debug output**
   ```bash
   grep -A5 "Update method called" debug.log
   grep -A5 "API response received" debug.log
   ```

### Phase 2: Implementation (3-4 hours)

1. **Create fix branch**
   ```bash
   git checkout -b fix/integration-policy-update-preserves-agent-policy-id
   ```

2. **Apply code changes**
   - Modify `Update()` method to preserve `agent_policy_id`
   - Update `buildUpdateRequest()` helper
   - Add validation to API client
   - Add debug logging

3. **Add unit tests**
   ```go
   // In integration_policy_test.go
   func TestIntegrationPolicyUpdate_PreservesAgentPolicyID(t *testing.T) {
       // Arrange
       state := models.IntegrationPolicy{
           ID:              types.StringValue("test-id"),
           AgentPolicyID:   types.StringValue("agent-policy-123"),
           IntegrationName: types.StringValue("auditd"),
           IntegrationVersion: types.StringValue("3.22.0"),
       }

       plan := state
       plan.IntegrationVersion = types.StringValue("3.23.0")

       // Act
       req := buildUpdateRequest(context.Background(), plan)

       // Assert
       assert.Equal(t, "agent-policy-123", req.PolicyID)
   }
   ```

4. **Add acceptance test**
   ```go
   func TestAccFleetIntegrationPolicy_versionUpdate(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               // Step 1: Create with version 3.22.0
               {
                   Config: testAccFleetIntegrationPolicyConfig("3.22.0"),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test", "integration_version", "3.22.0"),
                       resource.TestCheckResourceAttrSet("elasticstack_fleet_integration_policy.test", "agent_policy_id"),
                   ),
               },
               // Step 2: Update to version 3.23.0 (should not recreate)
               {
                   Config: testAccFleetIntegrationPolicyConfig("3.23.0"),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test", "integration_version", "3.23.0"),
                       resource.TestCheckResourceAttrSet("elasticstack_fleet_integration_policy.test", "agent_policy_id"),
                   ),
               },
           },
       })
   }
   ```

### Phase 3: Testing (2-3 hours)

1. **Run unit tests**
   ```bash
   go test ./internal/fleet/... -v
   ```

2. **Run acceptance tests**
   ```bash
   TF_ACC=1 go test ./internal/fleet/... -v -run TestAccFleetIntegrationPolicy_versionUpdate
   ```

3. **Integration test with DCM5 config**
   ```bash
   cd DCM5/prod

   # Clean slate
   ./cleanup-existing-resources.sh

   # Initial deployment
   terraform apply -auto-approve

   # Update version in variables.tf
   sed -i '' 's/3.22.0/3.23.0/' variables.tf

   # Apply update (should succeed without errors)
   terraform apply -auto-approve

   # Verify no resources recreated
   terraform plan  # Should show "No changes"
   ```

4. **Validation checklist**
   - [ ] `agent_policy_id` preserved during update
   - [ ] No state inconsistency errors
   - [ ] Works with multiple spaces (DCM5 scenario)
   - [ ] Works with default space only
   - [ ] All existing tests still pass
   - [ ] New tests added and passing

### Phase 4: Contribution (2 hours)

1. **Commit changes**
   ```bash
   git add .
   git commit -m "[Bug] Fix agent_policy_id null during integration policy update

   When updating elasticstack_fleet_integration_policy resources
   (e.g., changing integration_version), the Update method was not
   preserving the agent_policy_id field from state, causing Terraform
   to detect inconsistent state.

   This fix ensures computed fields like agent_policy_id are explicitly
   preserved during the update transformation and validates API responses
   for completeness.

   Fixes regression from PR #1390 (space-aware Fleet)
   Reported by: James Garside

   Changes:
   - Preserve agent_policy_id from state in Update method
   - Add validation for required fields in API requests
   - Add validation for API response completeness
   - Add debug logging for troubleshooting
   - Add unit tests for update preservation logic
   - Add acceptance test for version update scenario"
   ```

2. **Push branch**
   ```bash
   git push origin fix/integration-policy-update-preserves-agent-policy-id
   ```

3. **Open Pull Request**
   - Title: `[Bug] Fix agent_policy_id null during integration policy update`
   - Reference PR #1390 (space-aware Fleet)
   - Include reproduction steps
   - Link related issues
   - Request review from PR #1390 reviewers
   - Tag as bug fix for v0.13.2

4. **File GitHub Issue** (if not already filed)
   - Title: `elasticstack_fleet_integration_policy update fails with agent_policy_id null`
   - Labels: `bug`, `fleet`, `regression`
   - Include reproduction case
   - Reference PR #1390

## Current Workaround

While waiting for the upstream fix, use targeted destroy/recreate:

```bash
cd DCM5/prod

# Option 1: Destroy and recreate all integration policies
terraform destroy \
  -target=elasticstack_fleet_integration_policy.system_hostnation \
  -target=elasticstack_fleet_integration_policy.system_deployed \
  -target=elasticstack_fleet_integration_policy.endpoint_hostnation \
  -target=elasticstack_fleet_integration_policy.endpoint_deployed \
  -target=elasticstack_fleet_integration_policy.windows_hostnation \
  -target=elasticstack_fleet_integration_policy.windows_deployed \
  -target=elasticstack_fleet_integration_policy.auditd_hostnation \
  -target=elasticstack_fleet_integration_policy.auditd_deployed

terraform apply

# Option 2: Use cleanup script (if no agents enrolled yet)
./cleanup-existing-resources.sh
terraform apply
```

## Testing Checklist

### Local Development Testing
- [ ] Unit tests pass
- [ ] Acceptance tests pass
- [ ] Debug logging shows correct values
- [ ] Request includes agent_policy_id
- [ ] Response validated for completeness

### Integration Testing
- [ ] DCM5 40-team deployment succeeds
- [ ] Version update works without errors
- [ ] No resources recreated unnecessarily
- [ ] agent_policy_id preserved in state
- [ ] Works with multiple space_ids
- [ ] Works with default space only

### Regression Testing
- [ ] Create new integration policy works
- [ ] Delete integration policy works
- [ ] Update other fields (name, namespace) works
- [ ] Space-aware operations still work
- [ ] Non-space-aware operations still work

## Success Criteria

1. ✅ Integration version updates complete without state inconsistency
2. ✅ `agent_policy_id` preserved during all update operations
3. ✅ Space-aware updates function correctly
4. ✅ No resource recreation required for version changes
5. ✅ All existing tests pass
6. ✅ New tests added covering update scenarios
7. ✅ Fix merged upstream in v0.13.2 or later

## Timeline

| Phase | Duration | Target Date |
|-------|----------|-------------|
| Investigation | 2 hours | Day 1 Morning |
| Implementation | 3-4 hours | Day 1 Afternoon |
| Testing | 2-3 hours | Day 1 Evening |
| PR Submission | 1 hour | Day 2 Morning |
| Review Cycle | 3-7 days | Week 1-2 |
| Provider Release | 2-4 weeks | Month 1 |

## Files to Create/Modify

### Provider Repository
- `internal/fleet/integration_policy/resource.go` - Main fix location
- `internal/fleet/integration_policy/resource_test.go` - Unit tests
- `internal/fleet/integration_policy/acc_test.go` - Acceptance tests
- `internal/fleet/client.go` - API client validation
- `CHANGELOG.md` - Document fix

### DCM5 Repository
- `prod/TROUBLESHOOTING.md` - Document workaround
- `claudedocs/terraform-provider-integration-policy-bug-fix.md` - This document

## Related Resources

### Documentation
- [Terraform Plugin Framework - Resource Update](https://developer.hashicorp.com/terraform/plugin/framework/resources/update)
- [Fleet Package Policy API](https://www.elastic.co/guide/en/fleet/current/fleet-api-docs.html)
- [PR #1390 - Space-Aware Fleet](https://github.com/elastic/terraform-provider-elasticstack/pull/1390)

### Similar Issues
- [Issue #1469 - Inconsistent result in kibana_action_connector](https://github.com/elastic/terraform-provider-elasticstack/issues/1469)
- [Issue #689 - Integration policy diff issues](https://github.com/elastic/terraform-provider-elasticstack/issues/689)
- [Issue #531 - Integration policy state issues](https://github.com/elastic/terraform-provider-elasticstack/issues/531)

### Provider Links
- [GitHub Repository](https://github.com/elastic/terraform-provider-elasticstack)
- [Issue Tracker](https://github.com/elastic/terraform-provider-elasticstack/issues)
- [Terraform Registry](https://registry.terraform.io/providers/elastic/elasticstack/latest)

## Notes

- This bug is a regression from PR #1390 which added space-aware Fleet support
- The bug only affects **updates**, not creates or deletes
- The bug is directly related to your contribution, so you're well-positioned to fix it
- The fix should be straightforward - preserve computed fields during updates
- Consider this a high-priority fix since it blocks standard Terraform workflows

## Next Steps

1. Tomorrow morning: Start with Phase 1 (Investigation)
2. Verify the hypothesis with debug logging
3. Implement the fix if hypothesis is confirmed
4. Submit PR with comprehensive tests
5. Monitor review process and respond to feedback

---

**Document Status**: Ready for implementation
**Last Updated**: 2026-01-13
**Author**: James Garside (with Claude assistance)
