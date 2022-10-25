provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "example" {
  pipeline_id = "test_pipeline"
  description = "This is an example pipeline"

  pipeline = <<-EOF
  input{}
  filter{}
  output{}
EOF

  pipeline_metadata = {
    "type"    = "logstash_pipeline"
    "version" = 1
  }

  pipeline_batch_delay         = 50
  pipeline_batch_size          = 125
  pipeline_ecs_compatibility   = "disabled"
  pipeline_ordered             = "auto"
  pipeline_plugin_classloaders = false
  pipeline_unsafe_shutdown     = false
  pipeline_workers             = 1
  queue_checkpoint_acks        = 1024
  queue_checkpoint_retry       = true
  queue_checkpoint_writes      = 1024
  queue_drain                  = false
  queue_max_bytes_number       = 1
  queue_max_bytes_units        = "gb"
  queue_max_events             = 0
  queue_page_capacity          = "64mb"
  queue_type                   = "persisted"
}

output "pipeline" {
  value = elasticstack_elasticsearch_logstash_pipeline.example.pipeline_id
}
