variable "remote_cluster_alias" {
  type = string
}

variable "remote_proxy_address" {
  type = string
}

variable "leader_index_name" {
  type = string
}

variable "pattern_name" {
  type = string
}

variable "leader_index_patterns" {
  type = list(string)
}

variable "max_outstanding_read_requests" {
  type    = number
  default = null
}

variable "max_outstanding_write_requests" {
  type    = number
  default = null
}

variable "max_read_request_operation_count" {
  type    = number
  default = null
}

variable "max_read_request_size" {
  type    = string
  default = null
}

variable "max_retry_delay" {
  type    = string
  default = null
}

variable "max_write_buffer_count" {
  type    = number
  default = null
}

variable "max_write_buffer_size" {
  type    = string
  default = null
}

variable "max_write_request_operation_count" {
  type    = number
  default = null
}

variable "max_write_request_size" {
  type    = string
  default = null
}

variable "read_poll_timeout" {
  type    = string
  default = null
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

resource "elasticstack_elasticsearch_ccr_auto_follow_pattern" "test" {
  name                              = var.pattern_name
  remote_cluster                    = var.remote_cluster_alias
  leader_index_patterns             = var.leader_index_patterns
  max_outstanding_read_requests     = var.max_outstanding_read_requests
  max_outstanding_write_requests    = var.max_outstanding_write_requests
  max_read_request_operation_count  = var.max_read_request_operation_count
  max_read_request_size             = var.max_read_request_size
  max_retry_delay                   = var.max_retry_delay
  max_write_buffer_count            = var.max_write_buffer_count
  max_write_buffer_size             = var.max_write_buffer_size
  max_write_request_operation_count = var.max_write_request_operation_count
  max_write_request_size            = var.max_write_request_size
  read_poll_timeout                 = var.read_poll_timeout

  depends_on = [
    elasticstack_elasticsearch_cluster_settings.ccr_remote,
    elasticstack_elasticsearch_index.leader,
  ]
}
