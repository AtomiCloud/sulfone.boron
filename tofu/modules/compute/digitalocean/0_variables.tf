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

variable "instance_type" {
  type    = string
  default = "s-1vcpu-1gb"
}

variable "region" {
  type = string
}

variable "user" {
  type = string
}

variable "ssh_key" {
  type = list(string)
}