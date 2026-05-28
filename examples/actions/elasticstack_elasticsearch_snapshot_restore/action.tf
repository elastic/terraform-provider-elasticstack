# Requires Terraform 1.14+

action "elasticstack_elasticsearch_snapshot_restore" "dr_restore" {
  config {
    repository = elasticstack_elasticsearch_snapshot_repository.backup.name
    snapshot   = "my-snapshot-20240101"

    indices              = ["logs-*"]
    include_global_state = false
    ignore_unavailable   = true
    partial              = false
    include_aliases      = true

    rename_pattern     = "logs-(.+)"
    rename_replacement = "restored-logs-$1"
    index_settings     = jsonencode({ "index.number_of_replicas" = 0 })

    wait_for_completion = true

    timeouts {
      invoke = "30m"
    }
  }
}
