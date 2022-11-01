provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "my_script" {
  script_id = "my_script"
  lang      = "painless"
  source    = "Math.log(_score * 2) + params['my_modifier']"
  context   = "score"
}

resource "elasticstack_elasticsearch_script" "my_search_template" {
  script_id = "my_search_template"
  lang      = "mustache"
  source = jsonencode({
    query = {
      match = {
        message = "{{query_string}}"
      }
    }
    from = "{{from}}"
    size = "{{size}}"
  })
  params = jsonencode({
    query_string = "My query string"
  })
}
