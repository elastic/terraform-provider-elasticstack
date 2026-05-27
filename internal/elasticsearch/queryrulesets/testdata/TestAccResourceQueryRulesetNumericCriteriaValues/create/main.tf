variable "ruleset_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_query_ruleset" "test" {
  ruleset_id = var.ruleset_id

  rules = [
    {
      rule_id = "numeric-rule"
      type    = "pinned"

      criteria = [
        {
          type     = "gt"
          metadata = "popularity"
          values   = jsonencode([100])
        }
      ]

      actions = {
        ids = ["doc-1"]
      }
    },
  ]
}
