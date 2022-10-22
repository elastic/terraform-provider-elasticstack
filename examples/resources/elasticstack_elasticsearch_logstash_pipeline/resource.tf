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

  pipeline_batch_delay    = 100
  pipeline_batch_size     = 250
  pipeline_workers        = 2
  queue_checkpoint_writes = 2048
  queue_max_bytes_number  = 2
  queue_max_bytes_units   = "mb"
  queue_type              = "persisted"
}

output "pipeline" {
  value = elasticstack_elasticsearch_logstash_pipeline.example.pipeline_id
}
