variable "stream_name" {
  description = "Name of the pre-existing classic stream (data stream) to manage."
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_stream" "classic" {
  name     = var.stream_name
  space_id = "default"

  classic_config = {
    processing_steps = [
      jsonencode({
        action   = "grok"
        from     = "message"
        patterns = ["%%{GREEDYDATA:attributes.msg}"]
      })
    ]
  }
}
