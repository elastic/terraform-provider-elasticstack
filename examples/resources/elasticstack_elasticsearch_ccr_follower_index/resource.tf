provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ccr_follower_index" "my_follower" {
  name           = "follower-index"
  remote_cluster = "remote-cluster"
  leader_index   = "leader-index"

  status = "active"
}
