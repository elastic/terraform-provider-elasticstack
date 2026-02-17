// Bulk list/import Kibana dashboards.
//
// Requirements:
// - Terraform >= 1.14 (for `terraform query`)
// - Dashboards are experimental in this provider:
//   set `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` when running `terraform query`
//
// Usage:
//   TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true terraform query
//   TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true terraform query -generate-config-out=generated.tf
//
// Then copy the generated `resource` and `import` blocks into your configuration
// and run `terraform apply` to import.

list "elasticstack_kibana_dashboard" "all" {
  provider = elasticstack

  // Default is false. Set to true to have Terraform request full resource data.
  // This enables generating more complete `resource` blocks.
  include_resource = true

  // Optional: override Terraform's default (100).
  // limit = 1000

  config {
    // Defaults to "default".
    space_id = "default"

    // Kibana paging (server-side).
    per_page = 100

    // Optional filters.
    // search        = "my-dashboard*"
    // tags_included = ["tag-id-1"]
    // tags_excluded = ["tag-id-2"]
  }
}

