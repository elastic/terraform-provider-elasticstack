variable "script_id" {
  description = "The script ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "search_template_test" {
  script_id = var.script_id
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
