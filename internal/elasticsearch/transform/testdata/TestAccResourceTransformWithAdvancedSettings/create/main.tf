variable "transform_name" {
  type = string
}

variable "pipeline_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test" {
  name = var.pipeline_name

  processors = [
    jsonencode({
      set = {
        field = "_meta"
        value = "transformed"
      }
    })
  ]
}

resource "elasticstack_elasticsearch_transform" "test_advanced" {
  name        = var.transform_name
  description = "test advanced transform settings"

  source {
    indices = ["source_index_for_transform"]
    query   = jsonencode({ term = { status = "active" } })
    runtime_mappings = jsonencode({
      day_of_week = {
        type = "keyword"
        script = {
          source = "emit(doc['order_date'].value.dayOfWeekEnum.getDisplayName(TextStyle.FULL, Locale.ROOT))"
        }
      }
    })
  }

  destination {
    index    = "dest_index_for_transform_advanced"
    pipeline = elasticstack_elasticsearch_ingest_pipeline.test.name
  }

  metadata = jsonencode({
    owner = "test-team"
    env   = "ci"
  })

  pivot = jsonencode({
    group_by = {
      customer_id = {
        terms = {
          field = "customer_id"
        }
      }
    }
    aggregations = {
      total_sales = {
        sum = {
          field = "sales"
        }
      }
    }
  })

  sync {
    time {
      field = "order_date"
      delay = "20s"
    }
  }

  align_checkpoints     = true
  dates_as_epoch_millis = false
  deduce_mappings       = true
  docs_per_second       = 100
  num_failure_retries   = 5
  unattended            = false

  max_page_search_size = 2000
  frequency            = "5m"
  enabled              = false

  defer_validation = true
  timeout          = "1m"
}
