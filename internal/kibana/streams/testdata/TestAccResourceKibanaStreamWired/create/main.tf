variable "suffix" {
  description = "Random suffix to make the stream name unique."
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_stream" "wired" {
  name        = "logs.otel.testacc${var.suffix}"
  space_id    = "default"
  description = "Test wired stream"

  wired_config = {}
}
