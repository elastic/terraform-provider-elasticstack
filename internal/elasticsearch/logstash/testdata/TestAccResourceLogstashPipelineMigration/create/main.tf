variable "pipeline_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_migration" {
  pipeline_id = var.pipeline_id
  description = "Migration test pipeline"
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"
}
