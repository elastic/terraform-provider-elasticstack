provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ccr_auto_follow_pattern" "my_pattern" {
  name                  = "logs-auto-follow"
  remote_cluster        = "remote-cluster"
  leader_index_patterns = ["logs-*"]

  active = true
}
