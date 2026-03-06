variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                    = var.name
  type                    = "machine_learning"
  enabled                 = false
  description             = "Updated minimal test ML security detection rule"
  severity                = "medium"
  risk_score              = 55
  from                    = "now-12m"
  to                      = "now"
  interval                = "10m"
  anomaly_threshold       = 80
  machine_learning_job_id = ["test-ml-job", "test-ml-job-2"]
}

