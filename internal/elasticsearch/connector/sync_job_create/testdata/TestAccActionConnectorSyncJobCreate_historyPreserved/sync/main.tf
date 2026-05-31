variable "connector_id" {
  type = string
}

variable "wait_for_completion" {
  type = bool
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "source" {
  connector_id = var.connector_id
  service_type = "postgresql"
  name         = "TF acc sync job action"
  description  = "acceptance test connector for sync job action"
  index_name   = "content-connector-${var.connector_id}"
  is_native    = false

  scheduling = {
    full = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
    incremental = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
    access_control = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
  }
}

action "elasticstack_elasticsearch_connector_sync_job_create" "sync" {
  config {
    connector_id        = elasticstack_elasticsearch_connector.source.connector_id
    wait_for_completion = var.wait_for_completion
  }
}

resource "terraform_data" "trigger" {
  depends_on = [elasticstack_elasticsearch_connector.source]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.elasticstack_elasticsearch_connector_sync_job_create.sync]
    }
  }
}
