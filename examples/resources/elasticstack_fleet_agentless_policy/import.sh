# Import using the bare policy ID (assumes the default space)
terraform import elasticstack_fleet_agentless_policy.cspm_aws <fleet_agentless_policy_id>

# Or using the composite <space_id>/<policy_id> ID
terraform import elasticstack_fleet_agentless_policy.cspm_aws <space_id>/<fleet_agentless_policy_id>
