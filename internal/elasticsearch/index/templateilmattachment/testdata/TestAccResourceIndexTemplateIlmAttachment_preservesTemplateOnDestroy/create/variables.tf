variable "index_template" {
  type        = string
  description = "Name of the index template (component template will be <name>@custom)"
}

variable "policy_name" {
  type        = string
  description = "Name of the ILM policy"
}
