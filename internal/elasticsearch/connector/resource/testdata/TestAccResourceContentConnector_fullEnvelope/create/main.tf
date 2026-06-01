variable "connector_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "test" {
  connector_id = var.connector_id
  service_type = "postgresql"
  name         = "TF acc full envelope"
  description  = "full envelope acceptance test"
  index_name   = "content-connector-${var.connector_id}"
  language     = "en"

  pipeline = {
    name                   = "ent-search-generic-ingestion"
    extract_binary_content = true
    reduce_whitespace      = true
    run_ml_inference       = false
  }

  scheduling = {
    full = {
      enabled  = true
      interval = "0 0 * * * ?"
    }
    incremental = {
      enabled  = false
      interval = "0 30 * * * ?"
    }
    access_control = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
  }

  features = {
    document_level_security = {
      enabled = false
    }
    incremental_sync = {
      enabled = true
    }
    native_connector_api_keys = {
      enabled = false
    }
    sync_rules = {
      basic = {
        enabled = true
      }
    }
  }
}
