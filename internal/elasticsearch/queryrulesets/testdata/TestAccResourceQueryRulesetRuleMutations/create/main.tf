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
          type = "always"
        }
      ]

      actions = {
        ids = ["doc-1"]
      }
    },
    {
      rule_id  = "rule-2"
      type     = "pinned"
      priority = 5

      criteria = [
        {
          type = "always"
        }
      ]

      actions = {
        ids = ["doc-2"]
      }
    },
    {
      rule_id = "rule-3"
      type    = "exclude"

      criteria = [
        {
          type = "always"
        }
      ]

      actions = {
        ids = ["doc-3"]
      }
    },
  ]
}
