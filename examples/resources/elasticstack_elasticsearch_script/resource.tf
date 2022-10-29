provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "my_script" {
  script_id = "my_script"
  source    = "Math.log(_score * 2) + params['my_modifier']"
  context   = "score"
}
