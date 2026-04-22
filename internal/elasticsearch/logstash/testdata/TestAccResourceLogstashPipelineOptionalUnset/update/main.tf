variable "pipeline_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_optional" {
  pipeline_id = var.pipeline_id
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"
}
