variable "remote_cluster_alias" {
  type = string
}

variable "remote_proxy_address" {
  type = string
}

variable "leader_data_stream_name" {
  type = string
}

variable "data_stream_name" {
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

resource "elasticstack_elasticsearch_index_template" "leader_ds_template" {
  name           = var.leader_data_stream_name
  index_patterns = ["${var.leader_data_stream_name}*"]

  template {
    mappings = jsonencode({
      properties = {
        "@timestamp" = { type = "date" }
      }
    })
  }

  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "leader" {
  name = var.leader_data_stream_name

  depends_on = [
    elasticstack_elasticsearch_cluster_settings.ccr_remote,
    elasticstack_elasticsearch_index_template.leader_ds_template,
  ]
}

# A follower index that replicates a backing index of a remote data stream and
# attaches the result to a locally named data stream via data_stream_name. The
# leader data stream and local data stream names differ because the leader and
# follower share a cluster in the self-remote acceptance environment.
resource "elasticstack_elasticsearch_ccr_follower_index" "test" {
  name             = var.follower_index_name
  remote_cluster   = var.remote_cluster_alias
  leader_index     = elasticstack_elasticsearch_data_stream.leader.indices[0].index_name
  data_stream_name = var.data_stream_name

  depends_on = [
    elasticstack_elasticsearch_cluster_settings.ccr_remote,
    elasticstack_elasticsearch_data_stream.leader,
  ]
}
