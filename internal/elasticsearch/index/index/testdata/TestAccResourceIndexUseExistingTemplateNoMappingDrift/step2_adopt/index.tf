variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.index_name
  index_patterns = [var.index_name]

  template {
    mappings = jsonencode({
      dynamic_templates = [
        {
          strings_as_ip = {
            match_mapping_type = "string"
            match              = "ip*"
            runtime = {
              type = "ip"
            }
          }
        },
      ]
      properties = {
        template_field = {
          type = "keyword"
        }
      }
    })
  }
}

# User mapping subset; cluster mapping includes template-injected fields (superset).
resource "elasticstack_elasticsearch_index" "test_adopt_template" {
  name             = var.index_name
  use_existing     = true
  number_of_shards = 1

  mappings = jsonencode({
    properties = {
      user_declared_field = {
        type = "keyword"
      }
    }
  })

  deletion_protection = false

  depends_on = [elasticstack_elasticsearch_index_template.test]
}
