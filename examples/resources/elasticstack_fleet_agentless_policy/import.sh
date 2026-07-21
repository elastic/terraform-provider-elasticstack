# Import using the bare policy ID (assumes the default space)
terraform import elasticstack_fleet_managed_integration.cspm_aws <fleet_managed_integration_id>

# Or using the composite <space_id>/<policy_id> ID
terraform import elasticstack_fleet_managed_integration.cspm_aws <space_id>/<fleet_managed_integration_id>
