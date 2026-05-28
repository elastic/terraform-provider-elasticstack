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
      rule_id = "rule-fuzzy"
      type    = "pinned"

      criteria = [
        {
          type     = "fuzzy"
          metadata = "query"
          values   = jsonencode(["lap"])
        }
      ]

      actions = {
        ids = ["doc-1"]
      }
    },
    {
      rule_id = "rule-prefix"
      type    = "pinned"

      criteria = [
        {
          type     = "prefix"
          metadata = "query"
          values   = jsonencode(["lap"])
        }
      ]

      actions = {
        ids = ["doc-2"]
      }
    },
    {
      rule_id = "rule-suffix"
      type    = "pinned"

      criteria = [
        {
          type     = "suffix"
          metadata = "query"
          values   = jsonencode(["top"])
        }
      ]

      actions = {
        ids = ["doc-3"]
      }
    },
    {
      rule_id = "rule-lt"
      type    = "pinned"

      criteria = [
        {
          type     = "lt"
          metadata = "popularity"
          values   = jsonencode([10])
        }
      ]

      actions = {
        ids = ["doc-4"]
      }
    },
    {
      rule_id = "rule-lte"
      type    = "pinned"

      criteria = [
        {
          type     = "lte"
          metadata = "popularity"
          values   = jsonencode([50])
        }
      ]

      actions = {
        ids = ["doc-5"]
      }
    },
    {
      rule_id = "rule-gte"
      type    = "pinned"

      criteria = [
        {
          type     = "gte"
          metadata = "popularity"
          values   = jsonencode([5])
        }
      ]

      actions = {
        ids = ["doc-6"]
      }
    },
    {
      rule_id = "rule-multi"
      type    = "pinned"

      criteria = [
        {
          type     = "exact"
          metadata = "query"
          values   = jsonencode(["laptop"])
        },
        {
          type     = "prefix"
          metadata = "query"
          values   = jsonencode(["lap"])
        }
      ]

      actions = {
        ids = ["doc-7"]
      }
    },
  ]
}
