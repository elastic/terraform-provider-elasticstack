variable "connector_id" {
  type = string
}

variable "wait_for_completion" {
  type = bool
}

provider "elasticstack" {
  elasticsearch {}
}

action "elasticstack_elasticsearch_connector_sync_job_create" "sync" {
  config {
    connector_id        = var.connector_id
    wait_for_completion = var.wait_for_completion
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.elasticstack_elasticsearch_connector_sync_job_create.sync]
    }
  }
}
