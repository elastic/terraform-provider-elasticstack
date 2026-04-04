variable "pipeline_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test_enums" {
  pipeline_id                = var.pipeline_id
  pipeline                   = "input{} filter{} output{}"
  username                   = "test_user"
  pipeline_ecs_compatibility = "v1"
  pipeline_ordered           = "false"
}
