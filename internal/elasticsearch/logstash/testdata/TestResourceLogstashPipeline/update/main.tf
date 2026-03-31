variable "pipeline_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = var.pipeline_id
  description = "Updated description of Logstash Pipeline"
  pipeline    = "input{} \nfilter{} \noutput{}"
  username    = "test_user"

  pipeline_batch_delay         = 100
  pipeline_batch_size          = 250
  pipeline_ecs_compatibility   = "disabled"
  pipeline_metadata = jsonencode({
    type    = "logstash_pipeline"
    version = 2
  })
  pipeline_ordered             = "auto"
  pipeline_plugin_classloaders = false
  pipeline_unsafe_shutdown     = false
  pipeline_workers             = 2
  queue_checkpoint_acks        = 1024
  queue_checkpoint_retry       = true
  queue_checkpoint_writes      = 2048
  queue_drain                  = false
  queue_max_bytes              = "1mb"
  queue_max_events             = 0
  queue_page_capacity          = "64mb"
  queue_type                   = "memory"
}
