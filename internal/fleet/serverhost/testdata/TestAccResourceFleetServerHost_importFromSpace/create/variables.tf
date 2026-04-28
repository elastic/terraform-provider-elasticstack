variable "name" {
  description = "The server host name"
  type        = string
}

variable "space_id" {
  description = "The Kibana space ID to create the server host in"
  type        = string
}

variable "space_name" {
  description = "The Kibana space display name"
  type        = string
}
