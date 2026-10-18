variable "policy_name" {
  type        = string
  description = "Shared name for the agent policies in different spaces"
}

variable "space_id" {
  type        = string
  description = "Identifier of the non-default Kibana space"
}

variable "space_name" {
  type        = string
  description = "Name of the non-default Kibana space"
}
