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

  pipeline_settings = {
    "pipeline.workers"        = 2
    "pipeline.batch.size"     = 250
    "pipeline.batch.delay"    = 100
    "queue.type"              = "persisted"
    "queue.max_bytes.number"  = 2
    "queue.max_bytes.units"   = "mb"
    "queue.checkpoint.writes" = 2048
  }
}

output "pipeline" {
  value = elasticstack_elasticsearch_logstash_pipeline.example.pipeline_id
}
