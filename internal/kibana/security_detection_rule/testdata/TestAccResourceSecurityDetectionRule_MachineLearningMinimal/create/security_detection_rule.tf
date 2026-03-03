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
  description             = "Minimal test ML security detection rule"
  severity                = "low"
  risk_score              = 21
  from                    = "now-6m"
  to                      = "now"
  interval                = "5m"
  anomaly_threshold       = 75
  machine_learning_job_id = ["test-ml-job"]
}

