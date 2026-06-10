variable "remote_cluster_alias" {
  type = string
}

variable "remote_proxy_address" {
  type = string
}

variable "leader_index_name" {
  type = string
}

variable "follower_index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_cluster_settings" "ccr_remote" {
  persistent {
    setting {
      name  = "cluster.remote.${var.remote_cluster_alias}.mode"
      value = "proxy"
    }
    setting {
      name  = "cluster.remote.${var.remote_cluster_alias}.proxy_address"
      value = var.remote_proxy_address
    }
  }
}

resource "elasticstack_elasticsearch_index" "leader" {
  name                = var.leader_index_name
  deletion_protection = false

  mappings = jsonencode({
    properties = {
      field = { type = "keyword" }
    }
  })

  depends_on = [elasticstack_elasticsearch_cluster_settings.ccr_remote]
}

resource "elasticstack_elasticsearch_ccr_follower_index" "test" {
  name                              = var.follower_index_name
  remote_cluster                    = var.remote_cluster_alias
  leader_index                      = var.leader_index_name
  max_outstanding_read_requests     = 15
  max_outstanding_write_requests    = 10
  max_read_request_operation_count  = 5120
  max_read_request_size             = "40mb"
  max_retry_delay                   = "500ms"
  max_write_buffer_count            = 512
  max_write_buffer_size             = "512mb"
  max_write_request_operation_count = 5120
  max_write_request_size            = "40mb"
  read_poll_timeout                 = "5s"

  depends_on = [
    elasticstack_elasticsearch_cluster_settings.ccr_remote,
    elasticstack_elasticsearch_index.leader,
  ]
}
