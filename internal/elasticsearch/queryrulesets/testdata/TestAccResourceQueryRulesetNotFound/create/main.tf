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
      rule_id = "rule-1"
      type    = "pinned"

      criteria = [
        {
          type     = "exact"
          metadata = "query"
          values   = jsonencode(["test"])
        }
      ]

      actions = {
        ids = ["doc-1"]
      }
    },
  ]
}
