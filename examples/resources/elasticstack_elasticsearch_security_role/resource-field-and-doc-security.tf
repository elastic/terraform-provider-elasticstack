provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "field_and_doc_security" {
  name = "field_and_doc_security"

  indices {
    names      = ["customer-orders-*"]
    privileges = ["read", "view_index_metadata"]

    field_security {
      grant  = ["customer_id", "tenant_id", "order_id", "order_total", "order_status", "@timestamp"]
      except = ["customer_email", "customer_phone", "customer_ssn"]
    }

    query = jsonencode({
      term = {
        tenant_id = "tenant-a"
      }
    })
  }
}
