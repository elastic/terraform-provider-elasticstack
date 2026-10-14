variable "name" {
  type = string
}

variable "new_name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "cases"
  name       = var.new_name
  is_global  = false
  definition = <<-EOT
name: ${var.new_name}
type: keyword
label: Renamed Label
control: INPUT_TEXT
EOT
}
