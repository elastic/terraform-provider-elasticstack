variable "calendar_id" {
  type = string
}

variable "missing_job_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar_job" "test" {
  calendar_id = var.calendar_id
  job_id      = var.missing_job_id
}
