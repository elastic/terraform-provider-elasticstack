provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "example" {
  pipeline_id    = "test_pipeline"
  description    = "This is an example pipeline"

  pipeline = <<-EOF
  input{}
  filter{}
  output{}
EOF

  pipeline_metadata = jsonencode({
    "type"    = "logstash_pipeline"
    "version" = "1"
  })

  pipeline_settings = jsonencode({
    "pipeline.workers": 1,
    "pipeline.batch.size": 125,
    "pipeline.batch.delay": 50,
    "queue.type": "memory",
    "queue.max_bytes.number": 1,
    "queue.max_bytes.units": "gb",
    "queue.checkpoint.writes": 1024
  })
}

output "pipeline" {
  value = elasticstack_elasticsearch_logstash_pipeline.example.pipeline_id
}
