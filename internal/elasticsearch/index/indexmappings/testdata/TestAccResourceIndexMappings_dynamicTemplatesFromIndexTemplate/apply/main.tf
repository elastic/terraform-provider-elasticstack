variable "index_name" {
  type = string
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
          template_default = {
            match_mapping_type = "string"
            mapping = {
              type = "keyword"
            }
          }
        }
      ]
    })
  }
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false

  depends_on = [elasticstack_elasticsearch_index_template.test]
}

resource "elasticstack_elasticsearch_index_mappings" "test" {
  index = elasticstack_elasticsearch_index.test.name

  mappings = jsonencode({
    dynamic_templates = [
      {
        text_ja_example = {
          path_match = "hoge.example_field.freetext"
          mapping = {
            type = "text"
          }
        }
      }
    ]
  })
}
