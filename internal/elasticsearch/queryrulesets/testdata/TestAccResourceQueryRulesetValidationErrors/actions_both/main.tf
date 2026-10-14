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
          type = "always"
        }
      ]

      actions = {
        ids = ["doc-1"]
        docs = [
          {
            _index = "my-index"
            _id    = "42"
          }
        ]
      }
    },
  ]
}
