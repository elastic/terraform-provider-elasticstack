variable "pipeline_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_persisted" {
  pipeline_id = var.pipeline_id
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"

  queue_type              = "persisted"
  queue_max_bytes         = "512mb"
  queue_max_events        = 2000
  queue_page_capacity     = "128mb"
  queue_checkpoint_acks   = 512
  queue_checkpoint_writes = 1024
  queue_checkpoint_retry  = false
  queue_drain             = true
}
