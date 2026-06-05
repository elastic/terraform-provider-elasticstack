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
  name                          = var.pattern_name
  remote_cluster                = var.remote_cluster_alias
  leader_index_patterns         = var.leader_index_patterns
  max_outstanding_read_requests = var.max_outstanding_read_requests

  depends_on = [
    elasticstack_elasticsearch_cluster_settings.ccr_remote,
    elasticstack_elasticsearch_index.leader,
  ]
}
