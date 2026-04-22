variable "pipeline_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = var.pipeline_id
  description = "Description of Logstash Pipeline"
  pipeline    = "input{} filter{} output{}"
  username    = "test_user"
}
