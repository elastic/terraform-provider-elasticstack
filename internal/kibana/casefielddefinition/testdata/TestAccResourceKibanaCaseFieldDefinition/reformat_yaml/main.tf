variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "cases"
  name       = var.name
  is_global  = false
  definition = <<-EOT
control: INPUT_TEXT

label: "Renamed Label"
type: keyword
name: ${var.name}
EOT
}
