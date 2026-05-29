provider "elasticstack" {
  kibana {}
}

# Unfiltered read: returns all cloud connectors visible in the default space.
data "elasticstack_fleet_cloud_connectors" "all" {}

# Filtered read: restrict results with Fleet KQL (kuery query parameter).
data "elasticstack_fleet_cloud_connectors" "aws_only" {
  kuery = "fleet-cloud-connector.attributes.cloudProvider:aws"
}

output "cloud_connector_names" {
  description = "Names of all cloud connectors in the default space."
  value       = [for c in data.elasticstack_fleet_cloud_connectors.all.cloud_connectors : c.name]
}

output "aws_cloud_connector_names" {
  description = "Names of AWS cloud connectors returned by the filtered read."
  value       = [for c in data.elasticstack_fleet_cloud_connectors.aws_only.cloud_connectors : c.name]
}
