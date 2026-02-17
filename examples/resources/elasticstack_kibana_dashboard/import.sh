# Dashboard can be imported using the composite ID format: <space_id>/<dashboard_id>
# For example, to import a dashboard with ID "my-dashboard-id" from the default space:
terraform import elasticstack_kibana_dashboard.my_dashboard default/my-dashboard-id

# Bulk import via `terraform query` (Terraform >= 1.14)
#
# See `bulk_import.tfquery.hcl` in this directory.
# Note: dashboards are experimental in this provider, so enable them:
#
#   TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true terraform query -generate-config-out=generated.tf
#   # Copy generated blocks into your .tf files
#   terraform apply
