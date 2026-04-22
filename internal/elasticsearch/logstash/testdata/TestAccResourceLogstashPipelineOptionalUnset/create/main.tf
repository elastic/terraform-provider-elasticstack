variable "pipeline_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_optional" {
  pipeline_id                = var.pipeline_id
  pipeline                   = "input{} filter{} output{}"
  username                   = "test_user"
  description                = "Optional attrs set"
  pipeline_batch_delay       = 75
  pipeline_batch_size        = 150
  pipeline_workers           = 1
  pipeline_ecs_compatibility = "disabled"
  pipeline_ordered           = "auto"
  pipeline_unsafe_shutdown   = true
  queue_type                 = "memory"
  queue_max_bytes            = "256mb"
}
