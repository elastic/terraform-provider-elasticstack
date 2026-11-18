variable "policy_name" {
  type        = string
  description = "Name for the agent policy"
}

variable "space_id" {
  type        = string
  description = "Identifier of the first Kibana space"
}

variable "space_name" {
  type        = string
  description = "Name of the first Kibana space"
}

variable "second_space_id" {
  type        = string
  description = "Identifier of the second Kibana space"
}

variable "second_space_name" {
  type        = string
  description = "Name of the second Kibana space"
}
