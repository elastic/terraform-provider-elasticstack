variable "policy_name" {
  type = string
}

variable "preset" {
  type    = string
  default = null
}

variable "config_yaml" {
  type    = string
  default = null
}

variable "host" {
  type    = string
  default = "https://elasticsearch:9200"
}
