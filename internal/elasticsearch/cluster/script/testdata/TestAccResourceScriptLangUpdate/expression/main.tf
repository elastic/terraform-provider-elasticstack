variable "script_id" {
  description = "The script ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "test" {
  script_id = var.script_id
  lang      = "expression"
  source    = "_score * multiplier"
  context   = "score"
  params = jsonencode({
    multiplier = 3
  })
}
