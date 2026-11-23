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
    {
      rule_id = "rule-3"
      type    = "pinned"

      criteria = [
        {
          type = "always"
        }
      ]

      actions = {
        ids = ["doc-always"]
      }
    },
    {
      rule_id = "rule-4"
      type    = "pinned"

      criteria = [
        {
          type     = "gt"
          metadata = "popularity"
          values   = jsonencode([100])
        }
      ]

      actions = {
        ids = ["doc-numeric"]
      }
    },
  ]
}

data "elasticstack_elasticsearch_query_ruleset" "test" {
  ruleset_id = elasticstack_elasticsearch_query_ruleset.test.ruleset_id
}
