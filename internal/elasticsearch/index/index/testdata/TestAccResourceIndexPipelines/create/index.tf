variable "index_name" {
  description = "The index name"
  type        = string
}

variable "pipeline_name" {
  description = "The ingest pipeline name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = var.pipeline_name
  description = "Acceptance test pipeline"

  processors = []
}

resource "elasticstack_elasticsearch_index" "test_pipelines" {
  name             = var.index_name
  default_pipeline = elasticstack_elasticsearch_ingest_pipeline.test_pipeline.name
  final_pipeline   = elasticstack_elasticsearch_ingest_pipeline.test_pipeline.name

  deletion_protection = false
}
