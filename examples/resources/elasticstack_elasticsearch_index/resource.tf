provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "my-index"

  alias {
    name = "my_alias_1"
  }

  alias {
    name = "my_alias_2"
    filter = jsonencode({
      term = { "user.id" = "developer" }
    })
  }

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  settings {
    setting {
      name  = "index.number_of_replicas"
      value = "2"
    }
    setting {
      name  = "index.search.idle.after"
      value = "20s"
    }
  }
}
