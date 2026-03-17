variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                    = var.name
  type                    = "machine_learning"
  enabled                 = true
  description             = "Test ML validation bypass - neither index nor data_view_id required"
  severity                = "medium"
  risk_score              = 50
  anomaly_threshold       = 75
  machine_learning_job_id = ["test-ml-job"]
}

