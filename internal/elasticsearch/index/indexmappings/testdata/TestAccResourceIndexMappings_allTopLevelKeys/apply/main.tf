variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_mappings" "test" {
  index = elasticstack_elasticsearch_index.test.name

  mappings = jsonencode({
    dynamic = false
    _source = {
      enabled = true
    }
    dynamic_templates = [
      {
        strings_as_keywords = {
          match_mapping_type = "string"
          mapping = {
            type = "keyword"
          }
        }
      }
    ]
    properties = {
      title = { type = "text" }
    }
    runtime = {
      day_of_week = {
        type   = "keyword"
        script = "emit(doc['@timestamp'].value.dayOfWeekEnum.getDisplayName(TextStyle.FULL, Locale.ROOT))"
      }
    }
  })
}
