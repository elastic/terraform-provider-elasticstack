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
      rule_id = "rule-a"
      type    = "pinned"

      criteria = [
        {
          type     = "exact"
          metadata = "query"
          values   = jsonencode(["dog"])
        }
      ]

      actions = {
        ids = ["doc-a"]
      }
    },
    {
      rule_id = "rule-b"
      type    = "exclude"

      criteria = [
        {
          type     = "exact"
          metadata = "query"
          values   = jsonencode(["cat"])
        }
      ]

      actions = {
        ids = ["doc-b"]
      }
    },
  ]
}
