# Import using the bare managed integration ID (Fleet policy_id; default space)
terraform import elasticstack_fleet_managed_integration.cspm_aws <policy_id>

# Or using the composite <space_id>/<policy_id> ID (policy_id is stored as policy_id in state)
terraform import elasticstack_fleet_managed_integration.cspm_aws <space_id>/<policy_id>
