# Requires Terraform 1.14+

action "elasticstack_elasticsearch_snapshot_create" "manual" {
  config {
    repository = elasticstack_elasticsearch_snapshot_repository.backup.name
    snapshot   = "manual-snapshot-2024-01-01"

    indices              = ["logs-*"]
    include_global_state = false
    ignore_unavailable   = true
    partial              = false
    expand_wildcards     = "open"
    metadata             = jsonencode({ created_by = "terraform" })

    wait_for_completion = true

    timeouts {
      invoke = "60m"
    }
  }
}
