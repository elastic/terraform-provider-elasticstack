# Requires Terraform 1.14+ for provider-defined actions.

action "elasticstack_elasticsearch_connector_sync_job_create" "trigger" {
  config {
    connector_id        = elasticstack_elasticsearch_connector.postgres.connector_id
    job_type            = "full"
    trigger_method      = "on_demand"
    wait_for_completion = false

    timeouts {
      invoke = "60m"
    }
  }
}
