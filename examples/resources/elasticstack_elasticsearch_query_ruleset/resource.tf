resource "elasticstack_elasticsearch_query_ruleset" "my_ruleset" {
  ruleset_id = "my-search-rules"

  rules = [
    {
      rule_id  = "pin-laptops"
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
      rule_id = "exclude-deprecated"
      type    = "exclude"

      criteria = [
        {
          type     = "contains"
          metadata = "query"
          values   = jsonencode(["deprecated"])
        }
      ]

      actions = {
        docs = [
          {
            _index = "products"
            _id    = "old-1"
          }
        ]
      }
    }
  ]
}
