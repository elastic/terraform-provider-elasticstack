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
      rule_id = "docs-rule"
      type    = "pinned"

      criteria = [
        {
          type     = "exact"
          metadata = "query"
          values   = jsonencode(["laptop"])
        }
      ]

      actions = {
        docs = [
          {
            _index = "my-index-v2"
            _id    = "99"
          },
          {
            _index = "other-index"
            _id    = "7"
          }
        ]
      }
    },
  ]
}
