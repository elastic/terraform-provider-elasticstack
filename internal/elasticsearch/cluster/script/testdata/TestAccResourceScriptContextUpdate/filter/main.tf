variable "script_id" {
  description = "The script ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "test" {
  script_id = var.script_id
  lang      = "painless"
  source    = "params['enabled'] == true"
  context   = "filter"
  params = jsonencode({
    enabled = true
  })
}
