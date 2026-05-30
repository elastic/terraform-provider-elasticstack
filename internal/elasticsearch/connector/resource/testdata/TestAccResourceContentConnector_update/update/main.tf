variable "connector_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "test" {
  connector_id = var.connector_id
  service_type = "postgresql"
  name         = "TF acc updated"
  description  = "updated description"
  index_name   = "content-connector-upd-${var.connector_id}"
  language     = "en"

  pipeline = {
    name                   = "ent-search-generic-ingestion"
    extract_binary_content = false
    reduce_whitespace      = false
    run_ml_inference       = true
  }

  scheduling = {
    full = {
      enabled  = false
      interval = "0 15 * * * ?"
    }
    incremental = {
      enabled  = true
      interval = "0 45 * * * ?"
    }
    access_control = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
  }

  features = {
    document_level_security = {
      enabled = true
    }
    incremental_sync = {
      enabled = false
    }
    sync_rules = {
      basic = {
        enabled = false
      }
      advanced = {
        enabled = true
      }
    }
  }
}
