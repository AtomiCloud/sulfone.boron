variable "landscape" {
  type = string
}

variable "platform" {
  type = string
}

variable "service" {
  type = string
}

variable "module" {
  type = string
}

variable "cidr" {
  type = string
}

variable "instance_type" {
  type    = string
  default = "t4g.medium"
}

variable "user" {
  type = string
}

variable "ssh_key" {
  type = list(string)
}