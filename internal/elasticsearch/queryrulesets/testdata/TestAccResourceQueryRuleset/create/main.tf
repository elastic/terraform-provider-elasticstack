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
      rule_id  = "rule-1"
      type     = "pinned"
      priority = 1

      criteria = [
        {
          type     = "exact"
          metadata = "query"
          values   = jsonencode(["laptop", "notebook"])
        }
      ]

      actions = {
        ids = ["doc-1", "doc-2"]
      }
    },
    {
      rule_id = "rule-2"
      type    = "exclude"

      criteria = [
        {
          type     = "contains"
          metadata = "query"
          values   = jsonencode(["deprecated"])
        }
      ]

      actions = {
        ids = ["doc-old"]
      }
    },
  ]
}
