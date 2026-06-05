variable "remote_cluster_alias" {
  type = string
}

variable "remote_proxy_address" {
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

resource "elasticstack_elasticsearch_index_lifecycle" "leader_ilm" {
  name = var.data_stream_name

  hot {
    min_age = "1h"
    set_priority {
      priority = 10
    }
    rollover {
      max_age = "1d"
    }
    readonly {}
  }

  delete {
    min_age = "2d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_index_template" "leader_ds_template" {
  name = var.data_stream_name

  index_patterns = ["${var.data_stream_name}*"]

  template {
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.leader_ilm.name
    })
  }

  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "leader" {
  name = var.data_stream_name

  depends_on = [
    elasticstack_elasticsearch_cluster_settings.ccr_remote,
    elasticstack_elasticsearch_index_template.leader_ds_template,
  ]
}

resource "elasticstack_elasticsearch_ccr_follower_index" "test" {
  name           = var.follower_index_name
  remote_cluster = var.remote_cluster_alias
  leader_index   = var.data_stream_name

  depends_on = [
    elasticstack_elasticsearch_cluster_settings.ccr_remote,
    elasticstack_elasticsearch_data_stream.leader,
  ]
}
